package client

import (
	"fmt"

	"github.com/neblic/platform/controlplane/data"
)

type State int

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

type Client struct {
	State State
	Data  *data.Client
}

func New(uid data.ClientUID) *Client {
	return &Client{
		State: Unknown,
		Data:  data.NewClient(uid),
	}
}
