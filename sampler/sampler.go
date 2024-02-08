package sampler

import (
	"context"

	"github.com/neblic/platform/sampler/sample"
)

// Sampler defines the sampler public interface
type Sampler interface {
	// Sample samples the given data sample. Returns true if the sample has been exported.
	Sample(ctx context.Context, sample sample.Sample) bool
	// Close closes all Sampler connections with the Control and Data planes. Once closed,
	// the Sampler can't be reused and none of its methods can be called.
	Close() error
}

// New creates a new sampler using the global provider.
// Since it uses the global sampler provider, it is necessary to set one for the samplers to work.
// See sampler.SetProvider(...) method for more details.
func New(name string, schema sample.Schema, opts ...Option) (Sampler, error) {
	return globalProvider().Sampler(name, schema, opts...)
}

var registeredSamplers = make(map[string]Sampler)

// Sample samples the given data sample using the given sampler name and schema.
// If a sampler with the given name is not registered, it will be created.
// Note that if there is any error creating the sampler, it wont't be reported and the sample will be silently discarded.
func Sample(ctx context.Context, name string, schema sample.Schema, sample sample.Sample) bool {
	var sampler Sampler
	if s, ok := registeredSamplers[name]; ok {
		sampler = s
	} else {
		var err error
		sampler, err = New(name, schema)
		if err != nil {
			// unfortunately, we have no way to report the error
			return false
		}
		registeredSamplers[name] = sampler
	}

	return sampler.Sample(ctx, sample)
}
