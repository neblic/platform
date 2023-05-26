package defs

import (
	"context"

	"google.golang.org/protobuf/proto"
)

type SampleType uint8

const (
	UnknownSampleType SampleType = iota
	JSONSampleType
	NativeSampleType
	ProtoSampleType
)

type Sample struct {
	Determinant string

	Type   SampleType
	JSON   string
	Native any
	Proto  proto.Message
}

// JSONSample creates a data sample encoded as a JSON string.
// The JSON string must be a valid JSON object.
func JSONSample(json string, determinant string) Sample {
	return Sample{
		Determinant: determinant,

		Type: JSONSampleType,
		JSON: json,
	}
}

// NativeSample creates a data sample represented as a Go struct.
// Only exported fields will be part of the sample.
func NativeSample(native any, determinant string) Sample {
	return Sample{
		Determinant: determinant,

		Type:   NativeSampleType,
		Native: native,
	}
}

// ProtoSample creates a data sample encoded as a proto message. The protoSample parameter has to be the same
// type as the proto message provided as schema when creating the sampler.
func ProtoSample(proto proto.Message, determinant string) Sample {
	return Sample{
		Determinant: determinant,

		Type:  ProtoSampleType,
		Proto: proto,
	}
}

// Sampler defines the sampler public interface
type Sampler interface {
	// Sample samples the given data sample. Returns true if the sample has been exported.
	Sample(ctx context.Context, sample Sample) bool
	// Close closes all Sampler connections with the Control and Data planes. Once closed,
	// the Sampler can't be reused and none of its methods can be called.
	Close() error
}
