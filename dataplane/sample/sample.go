package sample

import (
	"time"

	"github.com/neblic/platform/controlplane/control"
)

const (
	// sample attributes
	OTLPLogSampleStreamUIDsKey  = "com.neblic.sample.stream.uids"
	OTLPLogSampleStreamNamesKey = "com.neblic.sample.stream.names"
	OTLPLogSampleKey            = "com.neblic.sample.key"
	OTLPLogSampleTypeKey        = "com.neblic.sample.type"
	OTLPLogSampleEncodingKey    = "com.neblic.sample.encoding"
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
	EventUID  MetadataKey = "com.neblic.event.uid"
	EventRule MetadataKey = "com.neblic.digest.rule"
	DigestUID MetadataKey = "com.neblic.digest.uid"
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
