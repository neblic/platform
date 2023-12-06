package event

import (
	"fmt"

	"github.com/neblic/platform/controlplane/control"
)

type Event interface {
	fmt.Stringer

	isEvent()
}

type ConfigUpdate struct {
	Resource string
	Sampler  string
	Config   control.SamplerConfig
}

func (cu ConfigUpdate) String() string {
	return fmt.Sprintf("ConfigUpdate(Resource: %s, Sampler: %s, Config %v)", cu.Resource, cu.Sampler, cu.Config)
}
func (ConfigUpdate) isEvent() {}

type ConfigDelete struct {
	Resource string
	Sampler  string
}

func (cd ConfigDelete) String() string {
	return fmt.Sprintf("ConfigDelete(Resource: %s, Sampler: %s)", cd.Resource, cd.Sampler)
}
func (ConfigDelete) isEvent() {}
