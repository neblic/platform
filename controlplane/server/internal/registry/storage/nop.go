package storage

type Nop struct {
}

func NewNop() *Nop {
	return &Nop{}
}

func (d *Nop) GetSampler(resource string, sampler string) (SamplerEntry, error) {
	return SamplerEntry{Resource: resource, Name: sampler}, nil
}

func (d *Nop) RangeSamplers(_ func(entry SamplerEntry)) error {
	return nil
}

func (d *Nop) SetSampler(_ SamplerEntry) error {
	return nil
}

func (d *Nop) DeleteSampler(_ string, _ string) error {
	return nil
}
