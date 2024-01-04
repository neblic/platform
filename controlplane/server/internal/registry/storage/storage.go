package storage

import (
	"fmt"

	"github.com/neblic/platform/controlplane/control"
)

var (
	ErrUnknownSampler = fmt.Errorf("unknown sampler")
)

type Storage interface {
	GetSampler(resource string, sampler string) (control.SamplerConfig, error)
	RangeSamplers(func(resource string, sampler string, config control.SamplerConfig)) error
	SetSampler(resource string, sampler string, config control.SamplerConfig) error
	DeleteSampler(resource string, sampler string) error
}
