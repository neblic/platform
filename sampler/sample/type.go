package sample

import "google.golang.org/protobuf/proto"

type Type uint8

const (
	UnknownSampleType Type = iota
	JSONSampleType
	NativeSampleType
	ProtoSampleType
)

type Sample struct {
	Type   Type
	JSON   string
	Native any
	Proto  proto.Message

	Options Options
}

// JSONSample creates a data sample encoded as a JSON string.
// The JSON string must be a valid JSON object.
func JSONSample(json string, sampleOpts ...Option) Sample {
	opts := Options{
		Size: len(json),
	}

	for _, opt := range sampleOpts {
		opt.apply(&opts)
	}

	return Sample{
		Type:    JSONSampleType,
		JSON:    json,
		Options: opts,
	}
}

// NativeSample creates a data sample represented as a Go struct.
// Only exported fields will be part of the sample.
func NativeSample(native any, sampleOpts ...Option) Sample {
	opts := Options{}
	for _, opt := range sampleOpts {
		opt.apply(&opts)
	}

	return Sample{
		Type:    NativeSampleType,
		Native:  native,
		Options: opts,
	}
}

// ProtoSample creates a data sample encoded as a proto message. The protoSample parameter has to be the same
// type as the proto message provided as schema when creating the sampler.
func ProtoSample(proto proto.Message, sampleOpts ...Option) Sample {
	opts := Options{}
	for _, opt := range sampleOpts {
		opt.apply(&opts)
	}

	return Sample{
		Type:    ProtoSampleType,
		Proto:   proto,
		Options: opts,
	}
}
