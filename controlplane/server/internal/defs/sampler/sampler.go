package sampler

import (
	"fmt"

	"github.com/neblic/platform/controlplane/data"
)

type State int

type Conn interface {
	Configure(*data.SamplerConfig) error
}

const (
	Unknown State = iota
	Unregistered
	Registered
)

func (s State) String() string {
	switch s {
	case Unknown:
		return "Unknown"
	case Unregistered:
		return "Unregistered"
	case Registered:
		return "Registered"
	default:
		return fmt.Sprintf("%d", int(s))
	}
}

type Sampler struct {
	State State
	Data  *data.Sampler

	Dirty bool
	Conn  Conn
}

func New(uid data.SamplerUID, name, resource string, conn Conn) *Sampler {
	return &Sampler{
		State: Unknown,
		Data:  data.NewSampler(name, resource, uid),
		Conn:  conn,
	}
}
