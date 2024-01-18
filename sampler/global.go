// It implements a sampler and a global provider placeholder so samplers can be created before setting the global provider.
// Once the global provider is set, any prob ebuilt before setting it is initialized with the new global provider.
package sampler

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/neblic/platform/sampler/sample"
)

/*
 * Based on the go.opentelemetry.io/otel/internal/global package
 */

type samplerProviderPlaceholder struct {
	sync.Mutex
	samplers map[string]*samplerPlaceholder

	delegate Provider
}

var _ Provider = &samplerProviderPlaceholder{}

func newProveProviderPlaceholder() *samplerProviderPlaceholder {
	return &samplerProviderPlaceholder{}
}

func (p *samplerProviderPlaceholder) setDelegate(pp Provider) error {
	p.Lock()
	defer p.Unlock()

	p.delegate = pp

	if len(p.samplers) == 0 {
		return nil
	}

	var aggErr error
	for _, pph := range p.samplers {
		if err := pph.setDelegate(pp); err != nil {
			if aggErr != nil {
				aggErr = fmt.Errorf("%s; %w", aggErr, err)
			} else {
				aggErr = err
			}
		}
	}

	p.samplers = nil

	return aggErr
}

func (p *samplerProviderPlaceholder) Sampler(name string, schema sample.Schema, opts ...Option) (Sampler, error) {
	p.Lock()
	defer p.Unlock()

	if p.delegate != nil {
		return p.delegate.Sampler(name, schema)
	}

	if p.samplers == nil {
		p.samplers = make(map[string]*samplerPlaceholder)
	}

	if val, ok := p.samplers[name]; ok {
		return val, nil
	}

	pw := &samplerPlaceholder{name: name, schema: schema}
	p.samplers[name] = pw

	return pw, nil
}

var _ Sampler = &samplerPlaceholder{}

type samplerPlaceholder struct {
	name   string
	schema sample.Schema

	delegate atomic.Value
}

func (pw *samplerPlaceholder) setDelegate(pp Provider) error {
	sampler, err := pp.Sampler(pw.name, pw.schema)
	if err != nil {
		return fmt.Errorf("error creating sampler: %w", err)
	}
	pw.delegate.Store(sampler)

	return nil
}

func (pw *samplerPlaceholder) Sample(ctx context.Context, sample sample.Sample) bool {
	delegate := pw.delegate.Load()
	if delegate != nil {
		return delegate.(Sampler).Sample(ctx, sample)
	}

	return false
}

func (pw *samplerPlaceholder) Close() error {
	delegate := pw.delegate.Load()
	if delegate != nil {
		return delegate.(Sampler).Close()
	}

	return nil
}

type samplerProviderHolder struct {
	pp Provider
}

var (
	delegateSamplerProviderOnce sync.Once
	globalSamplerProvider       = defaultProvidervalue()
)

func defaultProvidervalue() *atomic.Value {
	v := &atomic.Value{}
	v.Store(samplerProviderHolder{pp: newProveProviderPlaceholder()})
	return v
}

// globalProvider returns the global sampler provider. It is allowed to obtain and use the global sampler provider without
// setting one using SetGlobalSamplerProvider, but the generated samplers won't work until it is set.
func globalProvider() Provider {
	return globalSamplerProvider.Load().(samplerProviderHolder).pp
}

// SetProvider sets the global provider. It can only be set once.
// It may return sampler initialization errors because the creation of samplers created with the default global provider
// are deferred until it is set
func SetProvider(pp Provider) error {
	current := globalProvider()

	if _, cOk := current.(*samplerProviderPlaceholder); cOk {
		if _, tpOk := pp.(*samplerProviderPlaceholder); tpOk && current == pp {
			return fmt.Errorf("can't set placeholder sampler provider as sampler provider")
		}
	}

	var err error
	delegateSamplerProviderOnce.Do(func() {
		if def, ok := current.(*samplerProviderPlaceholder); ok {
			err = def.setDelegate(pp)
		}
	})
	globalSamplerProvider.Store(samplerProviderHolder{pp: pp})

	return err
}
