package registry

import (
	"fmt"

	data "github.com/neblic/platform/controlplane/data"
)

type Event interface {
	fmt.Stringer

	isEvent()
}

type ClientAction int

const (
	ClientRegistered ClientAction = iota
	ClientDeregistered
)

func (a ClientAction) String() string {
	switch a {
	case ClientRegistered:
		return "ClientRegistered"
	case ClientDeregistered:
		return "ClientDeregistered"
	default:
		return "Unknown"
	}
}

type ClientEvent struct {
	Action ClientAction
	UID    data.ClientUID
}

func (c *ClientEvent) String() string {
	return fmt.Sprintf("Client(Action: %s, UID: %s)", c.Action, c.UID)
}
func (*ClientEvent) isEvent() {}

type SamplerAction int

const (
	SamplerRegistered SamplerAction = iota
	SamplerDeregistered
)

func (a SamplerAction) String() string {
	switch a {
	case SamplerRegistered:
		return "SamplerRegistered"
	case SamplerDeregistered:
		return "SamplerDeregistered"
	default:
		return "Unknown"
	}
}

type SamplerEvent struct {
	Action SamplerAction
	UID    data.SamplerUID
}

func (p *SamplerEvent) String() string {
	return fmt.Sprintf("Sampler(Action: %s, UID: %s)", p.Action, p.UID)
}
func (*SamplerEvent) isEvent() {}

type ConfigAction int

const (
	ConfigUpdated ConfigAction = iota
	ConfigDeleted
)

func (a ConfigAction) String() string {
	switch a {
	case ConfigUpdated:
		return "ConfigUpdated"
	case ConfigDeleted:
		return "ConfigDeleted"
	default:
		return "Unknown"
	}
}

type ConfigEvent struct {
	Action          ConfigAction
	SamplerName     string
	SamplerResource string
	SamplerUID      data.SamplerUID
}

func (c *ConfigEvent) String() string {
	return fmt.Sprintf("Config(Action: %s, SamplerName: %s, SamplerResource: %s, SamplerUID: %s)", c.Action, c.SamplerName, c.SamplerResource, c.SamplerUID)
}
func (*ConfigEvent) isEvent() {}
