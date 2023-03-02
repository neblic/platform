package defs

import (
	"context"

	"google.golang.org/protobuf/proto"
)

// Sampler defines the sampler public interface
type Sampler interface {
	// SampleJSON samples a data sample encoded as a JSON string.
	SampleJSON(ctx context.Context, jsonSample string) (bool, error)
	// SampleNative samples a data sample represented as a Go struct.
	// Only exported fields will be part of the sample.
	SampleNative(ctx context.Context, nativeSample any) (bool, error)
	// SampleProto samples a data sampled encoded as a proto message. The protoSample parameter has to be the same
	// type as the proto message provided as schema when creating the sampler.
	SampleProto(ctx context.Context, protoSample proto.Message) (bool, error)
	// Close closes all Sampler connections with the Control and Data planes. Once closed,
	// the Sampler can't be reused and none of its methods can be called.
	Close() error
}
