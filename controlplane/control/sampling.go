package control

import (
	"github.com/neblic/platform/controlplane/protos"
)

type SamplingType int

const (
	UnknownSamplingType SamplingType = iota
	DeterministicSamplingType
)

type DeterministicSamplingConfig struct {
	SampleRate             int32
	SampleEmptyDeterminant bool
}

type SamplingConfig struct {
	SamplingType          SamplingType
	DeterministicSampling DeterministicSamplingConfig
}

func NewSamplingConfigFromProto(sr *protos.Sampling) SamplingConfig {
	if sr == nil {
		return SamplingConfig{}
	}

	var samplingType SamplingType
	switch sr.GetSampling().(type) {
	case *protos.Sampling_DeterministicSampling:
		samplingType = DeterministicSamplingType
		return SamplingConfig{
			SamplingType: samplingType,
			DeterministicSampling: DeterministicSamplingConfig{
				SampleRate:             sr.GetDeterministicSampling().GetSampleRate(),
				SampleEmptyDeterminant: sr.GetDeterministicSampling().GetSampleEmptyDeterminant(),
			},
		}
	default:
		return SamplingConfig{}
	}
}

func (sc SamplingConfig) ToProto() *protos.Sampling {
	switch sc.SamplingType {
	case DeterministicSamplingType:
		return &protos.Sampling{
			Sampling: &protos.Sampling_DeterministicSampling{
				DeterministicSampling: &protos.DeterministicSampling{
					SampleRate:             sc.DeterministicSampling.SampleRate,
					SampleEmptyDeterminant: sc.DeterministicSampling.SampleEmptyDeterminant,
				},
			},
		}
	case UnknownSamplingType:
		return &protos.Sampling{}
	default:
		return nil
	}
}
