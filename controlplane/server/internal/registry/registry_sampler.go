package registry

import (
	"errors"
	"fmt"
	"sync"

	"github.com/neblic/platform/controlplane/control"
	"github.com/neblic/platform/controlplane/server/internal/defs"
	"github.com/neblic/platform/controlplane/server/internal/registry/storage"
	"github.com/neblic/platform/internal/pkg/rule"
	"github.com/neblic/platform/logging"
	samplerDefs "github.com/neblic/platform/sampler/defs"
)

var (
	ErrUnknownSampler         = errors.New("unknown sampler")
	ErrUnknownSamplerInstance = errors.New("unknown sampler instance")
)

type SamplerRegistry struct {
	samplers    map[defs.SamplerIdentifier]*defs.Sampler
	storage     storage.Storage[defs.SamplerIdentifier, *defs.Sampler]
	ruleBuilder *rule.Builder
	notifyDirty chan struct{}

	logger logging.Logger
	m      sync.RWMutex
}

func NewSamplerRegistry(logger logging.Logger, notifyDirty chan struct{}, storageOpts storage.Options) (*SamplerRegistry, error) {
	if logger == nil {
		logger = logging.NewNopLogger()
	}

	// Initalize storage
	var storageInstance storage.Storage[defs.SamplerIdentifier, *defs.Sampler]
	switch storageOpts.Type {
	case storage.NopType:
		storageInstance = storage.NewNop[defs.SamplerIdentifier, *defs.Sampler]()
	case storage.DiskType:
		var err error
		storageInstance, err = storage.NewDisk[defs.SamplerIdentifier, *defs.Sampler](storageOpts.Path, "config")
		if err != nil {
			return nil, fmt.Errorf("error initializing sampler registry disk storage: %v", err)
		}
	}

	// Initialize the dynamic rule builder
	ruleBuilder, err := rule.NewBuilder(samplerDefs.NewDynamicSchema())
	if err != nil {
		return nil, fmt.Errorf("could not initialize the dynamic rule builder")
	}

	// Populate registry data using storage data
	samplers := map[defs.SamplerIdentifier]*defs.Sampler{}
	storageInstance.Range(func(key defs.SamplerIdentifier, sampler *defs.Sampler) {

		// Initialize event rules
		eventRules := map[control.SamplerEventUID]*rule.Rule{}
		for eventUID, event := range sampler.Config.Events {
			rule, err := ruleBuilder.Build(event.Rule.Expression)
			if err != nil {
				logger.Error("rule cannot be built. Skipping it", "resouce", key.Resource, "name", key.Name, "error", err)
				continue
			}

			eventRules[eventUID] = rule
		}

		// Initialize instances (not persisted)
		sampler.Instances = map[control.SamplerUID]*defs.SamplerInstance{}

		// Initialize sampler
		sampler.EventRules = eventRules

		samplers[key] = sampler
	})

	return &SamplerRegistry{
		samplers:    samplers,
		storage:     storageInstance,
		ruleBuilder: ruleBuilder,
		notifyDirty: notifyDirty,
		logger:      logger,
		m:           sync.RWMutex{},
	}, nil
}

func (sr *SamplerRegistry) getSampler(resource string, name string) (*defs.Sampler, error) {
	sampler, ok := sr.samplers[defs.NewSamplerIdentifier(resource, name)]
	if !ok {
		return nil, ErrUnknownSampler
	}
	return sampler, nil
}

func (sr *SamplerRegistry) setSampler(resource string, name string, sampler *defs.Sampler) error {
	// Set sampler in the registry
	samplerIdentifier := defs.NewSamplerIdentifier(resource, name)
	sr.samplers[samplerIdentifier] = sampler

	// Store sampler in the storage
	err := sr.storage.Set(samplerIdentifier, sampler)

	return err
}

func (sr *SamplerRegistry) getInstance(resource string, name string, uid control.SamplerUID) (*defs.SamplerInstance, error) {
	sampler, ok := sr.samplers[defs.NewSamplerIdentifier(resource, name)]
	if !ok {
		return nil, ErrUnknownSampler
	}

	instance, ok := sampler.GetInstance(uid)
	if !ok {
		return nil, ErrUnknownSamplerInstance
	}

	return instance, nil
}

func (sr *SamplerRegistry) setInstance(resource string, name string, instance *defs.SamplerInstance) error {
	if instance.UID == "" {
		return fmt.Errorf("provided instance has an empty UID")
	}

	// Get sampler
	sampler, err := sr.getSampler(resource, name)
	if err != nil {
		return err
	}

	// Set sampler instance
	sampler.SetInstance(instance.UID, instance)

	// Set sampler
	err = sr.setSampler(resource, name, sampler)

	return err
}

func (sr *SamplerRegistry) sendDirtyNotification() {
	select {
	case sr.notifyDirty <- struct{}{}:
	default:
	}
}

// RangeRegisteredInstances locks the registry until all the instances have been processed
// CAUTION: do not perform any action that may require registry access or it may cause a deadlock

func (sr *SamplerRegistry) RangeRegisteredInstances(fn func(sampler *defs.Sampler, instance *defs.SamplerInstance) (carryon bool)) {
	sr.m.Lock()
	defer sr.m.Unlock()

	for _, sampler := range sr.samplers {
		for _, instance := range sampler.Instances {
			if instance.Status == defs.RegisteredStatus {
				carryon := fn(sampler, instance)
				if !carryon {
					return
				}
			}
		}
	}
}

func (sr *SamplerRegistry) GetRegisteredInstances() []*defs.SamplerInstance {
	sr.m.RLock()
	defer sr.m.RUnlock()

	var instances []*defs.SamplerInstance
	for _, sampler := range sr.samplers {
		for _, instance := range sampler.Instances {
			if instance.Status == defs.RegisteredStatus {
				instances = append(instances, instance)
			}
		}
	}

	return instances
}

func (sr *SamplerRegistry) Register(resource string, name string, uid control.SamplerUID, conn defs.SamplerConn) error {
	sr.m.Lock()
	defer sr.m.Unlock()

	// Get sampler if exits, create it otherwise
	sampler, err := sr.getSampler(resource, name)
	if err != nil {
		if err != ErrUnknownSampler {
			return fmt.Errorf("unknown error happened when getting the sampler")
		}

		sampler = defs.NewSampler(resource, name)
	}

	// Get instance if exists, create it otherwise
	instance, ok := sampler.GetInstance(uid)
	if !ok {
		instance = defs.NewSamplerInstance(uid, sampler)
		sampler.SetInstance(uid, instance)
	}

	if instance.Status == defs.RegisteredStatus {
		sr.logger.Error("reregistering an already registered sampler", "sampler_uid", uid)

	}

	instance.UID = uid
	instance.Conn = conn
	instance.Dirty = true
	instance.Status = defs.RegisteredStatus

	err = sr.setSampler(resource, name, sampler)

	return err
}

func (sr *SamplerRegistry) Deregister(resource string, name string, uid control.SamplerUID) error {
	sr.m.Lock()
	defer sr.m.Unlock()

	sampler, err := sr.getSampler(resource, name)
	if err != nil {
		return ErrUnknownSampler
	}

	_, ok := sampler.Instances[uid]
	if !ok {
		sr.logger.Error("deregistering unknown sampler, nothing to do", "sampler_uid", uid)
		return nil
	}

	delete(sampler.Instances, uid)

	return nil
}

func (sr *SamplerRegistry) GetSampler(resource string, name string) (*defs.Sampler, error) {
	sr.m.RLock()
	defer sr.m.RUnlock()

	return sr.getSampler(resource, name)
}

func (sr *SamplerRegistry) UpdateSamplerConfig(resource string, name string, update control.SamplerConfigUpdate) error {
	sr.m.Lock()
	defer sr.m.Unlock()

	// Get current version of the sampler config
	sampler, err := sr.getSampler(resource, name)
	if err != nil {
		return err
	}

	// Update Streams
	if update.Reset.Streams || sampler.Config.Streams == nil {
		sampler.Config.Streams = make(map[control.SamplerStreamUID]control.Stream)
	}
	for _, rule := range update.StreamUpdates {
		switch rule.Op {
		case control.StreamUpsert:
			sampler.Config.Streams[rule.Stream.UID] = rule.Stream
		case control.StreamDelete:
			delete(sampler.Config.Streams, rule.Stream.UID)
		default:
			sr.logger.Error(fmt.Sprintf("received unkown sampling rule update operation: %d", rule.Op))
		}
	}

	// Update LimiterIn
	if update.Reset.LimiterIn {
		sampler.Config.LimiterIn = nil
	}
	if update.LimiterIn != nil {
		sampler.Config.LimiterIn = update.LimiterIn
	}

	// Update SamplingIn
	if update.Reset.SamplingIn {
		sampler.Config.SamplingIn = nil
	}
	if update.SamplingIn != nil {
		sampler.Config.SamplingIn = update.SamplingIn
	}

	// Update LimiterOut
	if update.Reset.LimiterOut {
		sampler.Config.LimiterOut = nil
	}
	if update.LimiterOut != nil {
		sampler.Config.LimiterIn = update.LimiterOut
	}

	// Update Digests
	if update.Reset.Digests || sampler.Config.Digests == nil {
		sampler.Config.Digests = make(map[control.SamplerDigestUID]control.Digest)
	}
	for _, rule := range update.DigestUpdates {
		switch rule.Op {
		case control.DigestUpsert:
			sampler.Config.Digests[rule.Digest.UID] = rule.Digest
		case control.DigestDelete:
			delete(sampler.Config.Digests, rule.Digest.UID)
		default:
			sr.logger.Error(fmt.Sprintf("received unkown digest update operation: %d", rule.Op))
		}
	}

	// Update Events
	if update.Reset.Events || sampler.Config.Events == nil {
		sampler.Config.Events = make(map[control.SamplerEventUID]control.Event)
	}
	for _, update := range update.EventUpdates {
		switch update.Op {
		case control.EventUpsert:
			// Update config
			sampler.Config.Events[update.Event.UID] = update.Event

			// Update event rules
			rule, err := sr.ruleBuilder.Build(update.Event.Rule.Expression)
			if err != nil {
				return fmt.Errorf("invalid event rule '%s': %w", update.Event.Rule.Expression, err)
			}
			sampler.EventRules[update.Event.UID] = rule

		case control.EventDelete:
			// Delete from config
			delete(sampler.Config.Events, update.Event.UID)

			// Delete from event rules
			delete(sampler.EventRules, update.Event.UID)
		default:
			sr.logger.Error(fmt.Sprintf("received unkown event update operation: %d", update.Op))
		}
	}

	// Mark instances as dirty and notify
	for _, instance := range sampler.Instances {
		instance.Dirty = true
	}
	defer sr.sendDirtyNotification()

	// Store updated sampler configuration
	err = sr.setSampler(resource, name, sampler)
	if err != nil {
		sr.logger.Error("could not store the sampler configuration updates", "error", err)
	}

	return nil
}

func (sr *SamplerRegistry) DeleteSamplerConfig(resource string, name string) error {
	sr.m.Lock()
	defer sr.m.Unlock()

	sampler, err := sr.getSampler(resource, name)
	if err != nil {
		return err
	}

	// Delete sampler configuration
	sampler.Config = *control.NewSamplerConfig()

	// Mark instances as dirty and notify
	for _, instance := range sampler.Instances {
		instance.Dirty = true
	}
	defer sr.sendDirtyNotification()

	// Store sampler without configuration
	err = sr.setSampler(resource, name, sampler)
	if err != nil {
		sr.logger.Error("could not store the sampler configuration delete", "error", err)
	}

	return nil
}

func (sr *SamplerRegistry) UpdateStats(resource string, name string, uid control.SamplerUID, newStats control.SamplerSamplingStats) error {
	sr.m.Lock()
	defer sr.m.Unlock()

	instance, err := sr.getInstance(resource, name, uid)
	if err != nil {
		return err
	}

	instance.Stats = newStats
	err = sr.setInstance(resource, name, instance)

	return err
}
