package defs

import (
	"github.com/neblic/platform/controlplane/control"
)

type SamplerConn interface {
	Configure(*control.SamplerConfig) error
}

type SamplerIdentifier struct {
	Resource string
	Name     string
}

func NewSamplerIdentifier(resource string, name string) SamplerIdentifier {
	return SamplerIdentifier{Resource: resource, Name: name}
}

type SamplerInstance struct {
	UID     control.SamplerUID
	Sampler *Sampler
	Conn    SamplerConn
	Dirty   bool
	Status  Status
	Stats   control.SamplerSamplingStats
}

func NewSamplerInstance(uid control.SamplerUID, sampler *Sampler) *SamplerInstance {
	return &SamplerInstance{
		UID:     uid,
		Sampler: sampler,
		Conn:    nil,
		Dirty:   true,
		Status:  UnknownStatus,
		Stats:   control.SamplerSamplingStats{},
	}
}

type Sampler struct {
	Resource       string
	Name           string
	Tags           control.Tags
	Capabilities   control.Capabilities
	Config         control.SamplerConfig
	Instances      map[control.SamplerUID]*SamplerInstance
	CollectorStats control.CollectorStats
}

func NewSampler(resource string, name string) *Sampler {
	return &Sampler{
		Resource:     resource,
		Name:         name,
		Config:       *control.NewSamplerConfig(),
		Capabilities: control.Capabilities{},
		Instances:    map[control.SamplerUID]*SamplerInstance{},
	}
}

func (s *Sampler) GetInstance(uid control.SamplerUID) (*SamplerInstance, bool) {
	instance, ok := s.Instances[uid]
	return instance, ok
}

func (s *Sampler) SetInstance(uid control.SamplerUID, samplerInstance *SamplerInstance) {
	s.Instances[uid] = samplerInstance
}
