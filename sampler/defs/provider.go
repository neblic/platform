package defs

// Provider defines a sampler provider object capable of creating new samplers.
type Provider interface {
	// Sampler creates a new sampler with the specified schema. It currently supports Dynamic and Proto schemas.
	// * A Dynamic schema does not enforce any structure to the sampled data and is compatible with all the Sample*()
	// methods. The downside, is that it is slower than the Proto schema since it needs to determine at runtime the sampled
	// data format.
	// * A Proto schema requires the caller to provide a proto message (type proto.Message) to define the sampler schema.
	// All sampled data is expected to be provided as proto messages with the sampler.SampleProto() method, and it should
	// be the same type as the one provided when defining the sampler schema.
	Sampler(name string, schema Schema) (Sampler, error)
}
