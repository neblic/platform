package sample

import (
	"time"

	"github.com/neblic/platform/controlplane/control"
)

const (
	// resource attributes
	rlSamplerNameKey = "sampler_name"
	// sample attributes
	lrSampleStreamsUIDsKey = "stream_uids"
	lrSampleKey            = "sample_key"
	lrSampleTypeKey        = "sample_type"
	lrSampleEncodingKey    = "sample_encoding"
)

type Encoding uint8

const (
	UnknownEncoding Encoding = iota
	JSONEncoding
)

func (s Encoding) String() string {
	switch s {
	case UnknownEncoding:
		return "unknown"
	case JSONEncoding:
		return "json"
	default:
		return "unknown"
	}
}

func ParseSampleEncoding(enc string) Encoding {
	switch enc {
	case "json":
		return JSONEncoding
	default:
		return UnknownEncoding
	}
}

type MetadataKey string

const (
	EventUID  MetadataKey = "event_uid"
	EventRule MetadataKey = "event_rule"
	DigestUID MetadataKey = "digest_uid"
)

// Sample defines a sample to be exported
type Sample struct {
	Ts       time.Time
	Type     control.SampleType
	Streams  []control.SamplerStreamUID
	Encoding Encoding
	Key      string
	Data     []byte
	Metadata map[MetadataKey]string
}

// SamplerSamples contains a group of samples that originate from the same resource e.g. operator, service...
type SamplerSamples struct {
	ResourceName string
	SamplerName  string
	Samples      []Sample
}
