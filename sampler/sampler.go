package sampler

import (
	"github.com/neblic/platform/sampler/defs"
	"github.com/neblic/platform/sampler/global"
)

// Sampler creates a new sampler using the global provider.
// Since it uses the global sampler provider, it is necessary to set one for the samplers to work.
// See global.SamplerProvider() and global.SetSamplerProvider(...) methods for more details.
func Sampler(name string, schema defs.Schema) (defs.Sampler, error) {
	return global.SamplerProvider().Sampler(name, schema)
}
