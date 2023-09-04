package control

import "github.com/neblic/platform/controlplane/protos"

type LimiterConfig struct {
	Limit int32
}

func NewLimiterFromProto(sr *protos.Limiter) LimiterConfig {
	if sr == nil {
		return LimiterConfig{}
	}

	return LimiterConfig{
		Limit: sr.Limit,
	}
}

func (sr LimiterConfig) ToProto() *protos.Limiter {
	return &protos.Limiter{
		Limit: sr.Limit,
	}
}
