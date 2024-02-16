package control

import "github.com/neblic/platform/controlplane/protos"

type StreamCapabilities struct {
	Enabled bool
}

func NewStreamCapabilitiesFromProto(streamCapabilities *protos.StreamCapabilities) StreamCapabilities {
	if streamCapabilities == nil {
		return StreamCapabilities{}
	}

	return StreamCapabilities{
		Enabled: streamCapabilities.GetEnabled(),
	}
}

func (sc StreamCapabilities) ToProto() *protos.StreamCapabilities {
	return &protos.StreamCapabilities{
		Enabled: sc.Enabled,
	}
}

type LimiterCapabilities struct {
	Enabled bool
}

func NewLimiterCapabilitiesFromProto(limiterCapabilities *protos.LimiterCapabilities) LimiterCapabilities {
	if limiterCapabilities == nil {
		return LimiterCapabilities{}
	}

	return LimiterCapabilities{
		Enabled: limiterCapabilities.GetEnabled(),
	}
}

func (lc LimiterCapabilities) ToProto() *protos.LimiterCapabilities {
	return &protos.LimiterCapabilities{
		Enabled: lc.Enabled,
	}
}

type SamplingCapabilities struct {
	Enabled bool
	Types   []SamplingType
}

func NewSamplingCapabilitiesFromProto(samplingCapabilities *protos.SamplingCapabilities) SamplingCapabilities {
	if samplingCapabilities == nil {
		return SamplingCapabilities{}
	}

	var types []SamplingType
	for _, t := range samplingCapabilities.GetTypes() {
		types = append(types, SamplingType(t))
	}

	return SamplingCapabilities{
		Enabled: samplingCapabilities.GetEnabled(),
		Types:   types,
	}
}

func (sc SamplingCapabilities) ToProto() *protos.SamplingCapabilities {
	var types []protos.SamplingCapabilities_Type
	for _, t := range sc.Types {
		types = append(types, protos.SamplingCapabilities_Type(t))
	}

	return &protos.SamplingCapabilities{
		Enabled: sc.Enabled,
		Types:   types,
	}
}

type DigestCapabilities struct {
	Enabled bool
	Types   []DigestType
}

func NewDigestCapabilitiesFromProto(digestCapabilities *protos.DigestCapabilities) DigestCapabilities {
	if digestCapabilities == nil {
		return DigestCapabilities{}
	}

	var types []DigestType
	for _, t := range digestCapabilities.GetTypes() {
		types = append(types, DigestType(t))
	}

	return DigestCapabilities{
		Enabled: digestCapabilities.GetEnabled(),
		Types:   types,
	}
}

func (dc DigestCapabilities) ToProto() *protos.DigestCapabilities {
	var types []protos.DigestCapabilities_Type
	for _, t := range dc.Types {
		types = append(types, protos.DigestCapabilities_Type(t))
	}

	return &protos.DigestCapabilities{
		Enabled: dc.Enabled,
		Types:   types,
	}
}

type Capabilities struct {
	Stream     StreamCapabilities
	LimiterIn  LimiterCapabilities
	SamplingIn SamplingCapabilities
	LimiterOut LimiterCapabilities
	Digest     DigestCapabilities
}

func NewCapabilitiesFromProto(capabilities *protos.Capabilities) Capabilities {
	if capabilities == nil {
		return Capabilities{}
	}

	return Capabilities{
		Stream:     NewStreamCapabilitiesFromProto(capabilities.GetStream()),
		LimiterIn:  NewLimiterCapabilitiesFromProto(capabilities.GetLimiterIn()),
		SamplingIn: NewSamplingCapabilitiesFromProto(capabilities.GetSamplingIn()),
		LimiterOut: NewLimiterCapabilitiesFromProto(capabilities.GetLimiterOut()),
		Digest:     NewDigestCapabilitiesFromProto(capabilities.GetDigest()),
	}
}

func (c Capabilities) ToProto() *protos.Capabilities {
	return &protos.Capabilities{
		Stream:     c.Stream.ToProto(),
		LimiterIn:  c.LimiterIn.ToProto(),
		SamplingIn: c.SamplingIn.ToProto(),
		LimiterOut: c.LimiterOut.ToProto(),
		Digest:     c.Digest.ToProto(),
	}
}
