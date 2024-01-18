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
