package storage

import "github.com/neblic/platform/controlplane/control"

type Nop struct {
}

func NewNop() *Nop {
	return &Nop{}
}

func (d *Nop) GetSampler(_ string, _ string) (control.SamplerConfig, error) {
	return control.SamplerConfig{}, nil
}

func (d *Nop) RangeSamplers(_ func(resource string, sampler string, config control.SamplerConfig)) error {
	return nil
}

func (d *Nop) SetSampler(_ string, _ string, _ control.SamplerConfig) error {
	return nil
}

func (d *Nop) DeleteSampler(_ string, _ string) error {
	return nil
}
