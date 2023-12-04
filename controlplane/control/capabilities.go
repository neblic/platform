package control

import "github.com/neblic/platform/controlplane/protos"

type CapabilitiesConfig struct {
	StructDigests bool
	ValueDigests  bool
}

func (cc *CapabilitiesConfig) CapableDigesters() []DigestType {
	digesters := []DigestType{}
	if cc.StructDigests {
		digesters = append(digesters, DigestTypeSt)
	}
	if cc.ValueDigests {
		digesters = append(digesters, DigestTypeValue)
	}
	return digesters
}

func NewCapabilitiesFromProto(sr *protos.Capabilities) CapabilitiesConfig {
	if sr == nil {
		return CapabilitiesConfig{}
	}

	return CapabilitiesConfig{
		StructDigests: sr.StructDigests,
		ValueDigests:  sr.ValueDigests,
	}
}

func (sr CapabilitiesConfig) ToProto() *protos.Capabilities {
	return &protos.Capabilities{
		StructDigests: sr.StructDigests,
		ValueDigests:  sr.ValueDigests,
	}
}
