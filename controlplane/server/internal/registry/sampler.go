package registry

import (
	"errors"
	"fmt"
	"sync"

	"github.com/neblic/platform/controlplane/data"
	internalsampler "github.com/neblic/platform/controlplane/server/internal/defs/sampler"
	"github.com/neblic/platform/logging"
)

var (
	ErrUnknownSampler = errors.New("unknown sampler")
)

type Sampler struct {
	samplers    map[data.SamplerUID]*internalsampler.Sampler
	notifyDirty chan struct{}

	logger logging.Logger
	m      sync.Mutex
}

func NewSampler(logger logging.Logger, notifyDirty chan struct{}) (*Sampler, error) {
	if logger == nil {
		logger = logging.NewNopLogger()
	}

	return &Sampler{
		samplers:    make(map[data.SamplerUID]*internalsampler.Sampler),
		notifyDirty: notifyDirty,
		logger:      logger,
	}, nil
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

func (p *Sampler) getSamplers(name, resource string, uid data.SamplerUID) ([]*internalsampler.Sampler, error) {
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
	p.m.Lock()
	defer p.m.Unlock()

	var registeredSamplers []*internalsampler.Sampler
	for _, sampler := range p.samplers {
		if sampler.State == internalsampler.Registered {
			registeredSamplers = append(registeredSamplers, sampler)
		}
	}

	return registeredSamplers
}

func (p *Sampler) Register(uid data.SamplerUID, name, resource string, conn internalsampler.Conn) error {
	p.m.Lock()
	defer p.m.Unlock()

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

	p.setSampler(uid, sampler)

	return nil
}

func (p *Sampler) Deregister(uid data.SamplerUID) error {
	p.m.Lock()
	defer p.m.Unlock()

	_, err := p.getSampler(uid)
	if errors.Is(err, ErrUnknownSampler) {
		p.logger.Error("deregistering unknown sampler, nothing to do", "sampler_uid", uid)
		return nil
	} else if err != nil {
		return err
	}

	delete(p.samplers, uid)

	return nil
}

func (p *Sampler) UpdateStats(uid data.SamplerUID, newStats data.SamplerSamplingStats) error {
	p.m.Lock()
	defer p.m.Unlock()

	foundSampler, err := p.getSampler(uid)
	if err != nil {
		return err
	}

	foundSampler.Data.SamplingStats = newStats

	return nil
}

func (p *Sampler) MarkDirty(name, resource string, uid data.SamplerUID) error {
	p.m.Lock()
	defer p.m.Unlock()

	samplers, err := p.getSamplers(name, resource, uid)
	if err != nil {
		return err
	}

	for _, sampler := range samplers {
		sampler.Dirty = true
	}

	select {
	case p.notifyDirty <- struct{}{}:
	default:
	}

	return nil
}
