package sampler

import "github.com/neblic/platform/controlplane/internal/stream"

type State int

const (
	Unknown State = iota
	Unregistered
	Registering
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
		return "Unknown"
	}
}

func NewStateFromStreamState(st stream.State) State {
	switch st {
	case stream.Registering:
		return Registering
	case stream.Unregistered:
		return Unregistered
	case stream.Registered:
		return Registered
	case stream.Unknown:
		fallthrough
	default:
		return Unknown
	}
}
