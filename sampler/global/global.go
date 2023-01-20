// Package global holds the global sampler provider.
//
// It also implements a sampler and a global provider placeholder so samplers can be created before setting the global provider.
// Once the global provider is set, any prob ebuilt before setting it is initialized with the new global provider.
package global

import (
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/neblic/platform/sampler/defs"
)

/*
	Based on the go.opentelemetry.io/otel/internal/global package
*/

type samplerProviderHolder struct {
	pp defs.Provider
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

// SamplerProvider returns the global sampler provider. It is allowed to obtain and use the global sampler provider without
// setting one using SetSamplerProvider, but the generated samplers won't work until it is set.
func SamplerProvider() defs.Provider {
	return globalSamplerProvider.Load().(samplerProviderHolder).pp
}

// SetSamplerProvider sets the global provider. It can only be set once.
// It may return sampler initialization errors because the creation of samplers created with the default global provider
// are deferred until it is set
func SetSamplerProvider(pp defs.Provider) error {
	current := SamplerProvider()

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
