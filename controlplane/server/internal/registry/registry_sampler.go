package registry

import (
	"errors"
	"fmt"
	"sync"

	"github.com/neblic/platform/controlplane/control"
	"github.com/neblic/platform/controlplane/event"
	"github.com/neblic/platform/controlplane/server/internal/defs"
	"github.com/neblic/platform/controlplane/server/internal/registry/storage"
	"github.com/neblic/platform/logging"
)

var (
	ErrUnknownSampler         = errors.New("unknown sampler")
	ErrUnknownSamplerInstance = errors.New("unknown sampler instance")
)

type SamplerRegistry struct {
	samplers map[defs.SamplerIdentifier]*defs.Sampler
	storage  storage.Storage[defs.SamplerIdentifier, *defs.Sampler]

	eventsChan  chan event.Event
	notifyDirty chan struct{}

	logger logging.Logger
	m      sync.RWMutex
}

func NewSamplerRegistry(logger logging.Logger, notifyDirty chan struct{}, storageOpts storage.Options) (*SamplerRegistry, error) {
	if logger == nil {
		logger = logging.NewNopLogger()
	}

	// Initialize storage
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

	// Populate registry data using storage data
	samplers := map[defs.SamplerIdentifier]*defs.Sampler{}
	err := storageInstance.Range(func(key defs.SamplerIdentifier, sampler *defs.Sampler) {

		// Initialize instances (not persisted)
		sampler.Instances = map[control.SamplerUID]*defs.SamplerInstance{}
		samplers[key] = sampler
	})
	if err != nil {
		return nil, fmt.Errorf("error populating sampler registry from storage: %v", err)
	}

	return &SamplerRegistry{
		samplers:    samplers,
		storage:     storageInstance,
		eventsChan:  nil,
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

// RangeSamplers locks the registry until all the configs have been processed
// CAUTION: do not perform any action that may require registry access or it may cause a deadlock
func (sr *SamplerRegistry) RangeSamplers(fn func(sampler *defs.Sampler) (carryon bool)) {
	sr.m.Lock()
	defer sr.m.Unlock()

	for _, sampler := range sr.samplers {
		carryon := fn(sampler)
		if !carryon {
			return
		}
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

func (sr *SamplerRegistry) Register(resource string, name string, uid control.SamplerUID, conn defs.SamplerConn, initialConfig control.SamplerConfig) error {
	sr.m.Lock()
	defer sr.m.Unlock()

	// Get sampler if exits, create it otherwise
	sampler, err := sr.getSampler(resource, name)
	if err != nil {
		if err != ErrUnknownSampler {
			return fmt.Errorf("unknown error happened when getting the sampler")
		}

		sampler = defs.NewSampler(resource, name)
		sampler.Config = initialConfig
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

	// Send upsert event if necessary
	if sr.eventsChan != nil {
		sr.eventsChan <- event.ConfigUpdate{
			Resource: resource,
			Sampler:  name,
			Config:   sampler.Config,
		}
	}

	return err
}

func (sr *SamplerRegistry) Deregister(uid control.SamplerUID) error {
	sr.m.Lock()
	defer sr.m.Unlock()

	// Find sampler instance
	var found bool
	var sampler *defs.Sampler
	for _, sampler = range sr.samplers {
		_, ok := sampler.GetInstance(uid)
		if ok {
			found = true
			break
		}
	}
	if !found {
		return ErrUnknownSamplerInstance
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

	// Update sampler configuration
	sampler.Config.Merge(update)

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

	// Send upsert event if necessary
	if sr.eventsChan != nil {
		sr.eventsChan <- event.ConfigUpdate{
			Resource: resource,
			Sampler:  name,
			Config:   sampler.Config,
		}
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

	// Send delete event if necessary
	if sr.eventsChan != nil {
		sr.eventsChan <- event.ConfigDelete{
			Resource: resource,
			Sampler:  name,
		}
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

// Events returns a new channel that will be populated with the sampler configs. Events will contain the initial state
// and posterior updates
// CAUTION: Not reading from the returned channel until it gets closed will block the registry
func (sr *SamplerRegistry) Events() chan event.Event {
	if sr.eventsChan == nil {
		sr.eventsChan = make(chan event.Event)

		// Send config state to the created channel. That blocks the full registry until the goroutine finishes
		go func() {
			sr.RangeSamplers(func(sampler *defs.Sampler) (carryon bool) {
				sr.eventsChan <- event.ConfigUpdate{
					Resource: sampler.Resource,
					Sampler:  sampler.Name,
					Config:   sampler.Config,
				}

				return true
			})
		}()
	}

	return sr.eventsChan
}

func (sr *SamplerRegistry) Close() {
	if sr.eventsChan != nil {
		close(sr.eventsChan)
	}
}
