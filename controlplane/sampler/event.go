package sampler

import (
	"fmt"

	"github.com/neblic/platform/controlplane/control"
)

type Event interface {
	fmt.Stringer

	isEvent()
}

type ConfigUpdate struct {
	Config control.SamplerConfig
}

func (p ConfigUpdate) String() string {
	return fmt.Sprintf("ConfigUpdate(Config %v)", p.Config)
}
func (ConfigUpdate) isEvent() {}

type StateUpdate struct {
	State State
}

func (p StateUpdate) String() string {
	return fmt.Sprintf("StateUpcate(State: %s)", p.State)
}
func (StateUpdate) isEvent() {}
