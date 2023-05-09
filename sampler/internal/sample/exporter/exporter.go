package exporter

import (
	"context"
	"time"

	"github.com/neblic/platform/controlplane/data"
)

type SampleType uint8

const (
	UnknownSampleType SampleType = iota
	RawSampleType
	StructDigestSampleType
)

type SampleEncoding uint8

func (s SampleType) String() string {
	switch s {
	case UnknownSampleType:
		return "unknown"
	case RawSampleType:
		return "raw"
	case StructDigestSampleType:
		return "struct-digest"
	default:
		return "unknown"
	}
}

const (
	UnknownSampleEncoding SampleEncoding = iota
	JSONSampleEncoding
)

func (s SampleEncoding) String() string {
	switch s {
	case UnknownSampleEncoding:
		return "unknown"
	case JSONSampleEncoding:
		return "json"
	default:
		return "unknown"
	}
}

// Sample defines a sample to be exported
type Sample struct {
	Ts       time.Time
	Type     SampleType
	Streams  []data.SamplerStreamUID
	Encoding SampleEncoding
	Data     []byte
}

// SamplerSamples contains a group of samples that originate from the same resource e.g. operator, service...
type SamplerSamples struct {
	ResourceName string
	SamplerName  string
	Samples      []Sample
}

type Exporter interface {
	Export(context.Context, []SamplerSamples) error
	Close(context.Context) error
}
