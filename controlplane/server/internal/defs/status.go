package defs

import "fmt"

type Status int

const (
	UnknownStatus Status = iota
	UnregisteredStatus
	RegisteredStatus
)

func (s Status) String() string {
	switch s {
	case UnknownStatus:
		return "Unknown"
	case UnregisteredStatus:
		return "Unregistered"
	case RegisteredStatus:
		return "Registered"
	default:
		return fmt.Sprintf("%d", int(s))
	}
}
