package defs

import (
	"github.com/neblic/platform/controlplane/control"
	"github.com/neblic/platform/internal/pkg/rule"
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
	Sampler *Sampler                     `json:"-"`
	Stats   control.SamplerSamplingStats `json:"-"`
	Conn    SamplerConn                  `json:"-"`
	Dirty   bool                         `json:"-"`
	Status  Status                       `json:"-"`
}

func NewSamplerInstance(uid control.SamplerUID, sampler *Sampler) *SamplerInstance {
	return &SamplerInstance{
		UID:     uid,
		Sampler: sampler,
		Stats:   control.SamplerSamplingStats{},
		Conn:    nil,
		Dirty:   true,
		Status:  UnknownStatus,
	}
}

type Sampler struct {
	Resource  string
	Name      string
	Config    control.SamplerConfig
	Instances map[control.SamplerUID]*SamplerInstance

	// EventRules is used to evaluate event rules. It is computed from the config events
	EventRules map[control.SamplerEventUID]*rule.Rule `json:"-"`
}

func NewSampler(resource string, name string) *Sampler {
	return &Sampler{
		Resource:  resource,
		Name:      name,
		Config:    *control.NewSamplerConfig(),
		Instances: map[control.SamplerUID]*SamplerInstance{},

		EventRules: map[control.SamplerEventUID]*rule.Rule{},
	}
}

func (s *Sampler) GetInstance(uid control.SamplerUID) (*SamplerInstance, bool) {
	instance, ok := s.Instances[uid]
	return instance, ok
}

func (s *Sampler) SetInstance(uid control.SamplerUID, samplerInstance *SamplerInstance) {
	s.Instances[uid] = samplerInstance
}
