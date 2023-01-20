package registry

import (
	"errors"
	"fmt"
	"sync"

	"github.com/neblic/platform/controlplane/data"
	"github.com/neblic/platform/controlplane/server/internal/defs/sampler"
	internalsampler "github.com/neblic/platform/controlplane/server/internal/defs/sampler"
	"github.com/neblic/platform/logging"
)

var (
	ErrUnknownSampler = errors.New("unknown sampler")
)

type Sampler struct {
	samplers map[data.SamplerUID]*internalsampler.Sampler
	eventsCh chan *SamplerEvent

	logger logging.Logger
	sync.Mutex
}

func NewSampler(logger logging.Logger) (*Sampler, error) {
	if logger == nil {
		logger = logging.NewNopLogger()
	}

	return &Sampler{
		samplers: make(map[data.SamplerUID]*internalsampler.Sampler),
		logger:   logger,
	}, nil
}

func (p *Sampler) Events() chan *SamplerEvent {
	if p.eventsCh == nil {
		// eventsCh needs to be buffer to avoid a deadlock caused by methods accessing the registry in response to its events
		p.eventsCh = make(chan *SamplerEvent, 10)
	}

	return p.eventsCh
}

func (p *Sampler) getSampler(uid data.SamplerUID) (*internalsampler.Sampler, error) {
	foundSampler, ok := p.samplers[uid]
	if !ok {
		return nil, fmt.Errorf("%w, uid: %s", ErrUnknownSampler, uid)
	}

	return foundSampler, nil
}

func (p *Sampler) setSampler(uid data.SamplerUID, sampler *internalsampler.Sampler) {
	p.samplers[uid] = sampler
}

// GetSamplers returns all samplers that match the provided arguments.
// If UID is provided, it will return at most, one element
func (p *Sampler) GetSamplers(name, resource string, uid data.SamplerUID) ([]*internalsampler.Sampler, error) {
	p.Lock()
	defer p.Unlock()

	if uid != "" {
		sampler, err := p.getSampler(uid)

		return []*internalsampler.Sampler{sampler}, err
	}

	var samplers []*internalsampler.Sampler
	for _, sampler := range p.samplers {
		if sampler.Data.Name == name && sampler.Data.Resource == resource {
			samplers = append(samplers, sampler)
		}
	}

	if len(samplers) == 0 {
		return nil, fmt.Errorf("%w, name: %s", ErrUnknownSampler, name)
	}

	return samplers, nil
}

func (p *Sampler) GetRegisteredSamplers() []*internalsampler.Sampler {
	p.Lock()
	defer p.Unlock()

	var registeredSamplers []*internalsampler.Sampler
	for _, sampler := range p.samplers {
		if sampler.State == internalsampler.Registered {
			registeredSamplers = append(registeredSamplers, sampler)
		}
	}

	return registeredSamplers
}

func (p *Sampler) Register(uid data.SamplerUID, name, resource string, conn sampler.Conn) error {
	p.Lock()
	defer p.Unlock()

	knownSampler, err := p.getSampler(uid)
	if err != nil && !errors.Is(err, ErrUnknownSampler) {
		return err
	} else if err == nil {
		if knownSampler.State == internalsampler.Registered {
			p.logger.Error("reregistering an already registered sampler", "sampler_uid", uid)
		}
	}

	sampler := internalsampler.New(uid, name, resource, conn)
	sampler.State = internalsampler.Registered
	sampler.Dirty = true

	p.samplers[uid] = sampler

	p.sendEvent(&SamplerEvent{
		Action: SamplerRegistered,
		UID:    uid,
	})

	return nil
}

func (p *Sampler) Deregister(uid data.SamplerUID) error {
	p.Lock()
	defer p.Unlock()

	_, err := p.getSampler(uid)
	if errors.Is(err, ErrUnknownSampler) {
		p.logger.Error("deregistering unknown sampler, nothing to do", "sampler_uid", uid)

		return nil
	} else if err != nil {
		return err
	}

	delete(p.samplers, uid)

	p.sendEvent(&SamplerEvent{
		Action: SamplerDeregistered,
		UID:    uid,
	})

	return nil
}

func (p *Sampler) UpdateSamplerStats(uid data.SamplerUID, newStats data.SamplerSamplingStats) error {
	p.Lock()
	defer p.Unlock()

	foundSampler, err := p.getSampler(uid)
	if err != nil {
		return err
	}

	foundSampler.Data.SamplingStats = newStats

	return nil
}

func (p *Sampler) sendEvent(ev *SamplerEvent) {
	if p.eventsCh != nil {
		p.eventsCh <- ev
	}
}
