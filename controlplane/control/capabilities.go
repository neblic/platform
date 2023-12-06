package control

import "github.com/neblic/platform/controlplane/protos"

type CapabilitiesConfig struct {
	StructDigest bool
	ValueDigest  bool
}

func NewCapabilitiesFromProto(c *protos.Capabilities) CapabilitiesConfig {
	if c == nil {
		return CapabilitiesConfig{}
	}

	return CapabilitiesConfig{
		StructDigest: c.StructDigest,
		ValueDigest:  c.ValueDigest,
	}
}

func (cc CapabilitiesConfig) ToProto() *protos.Capabilities {
	return &protos.Capabilities{
		StructDigest: cc.StructDigest,
		ValueDigest:  cc.ValueDigest,
	}
}
