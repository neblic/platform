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
	Key string

	Type   Type
	JSON   string
	Native any
	Proto  proto.Message
}

// JSONSample creates a data sample encoded as a JSON string.
// The JSON string must be a valid JSON object.
func JSONSample(json string, key string) Sample {
	return Sample{
		Key: key,

		Type: JSONSampleType,
		JSON: json,
	}
}

// NativeSample creates a data sample represented as a Go struct.
// Only exported fields will be part of the sample.
func NativeSample(native any, key string) Sample {
	return Sample{
		Key: key,

		Type:   NativeSampleType,
		Native: native,
	}
}

// ProtoSample creates a data sample encoded as a proto message. The protoSample parameter has to be the same
// type as the proto message provided as schema when creating the sampler.
func ProtoSample(proto proto.Message, key string) Sample {
	return Sample{
		Key: key,

		Type:  ProtoSampleType,
		Proto: proto,
	}
}
