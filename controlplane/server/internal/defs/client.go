package defs

import "github.com/neblic/platform/controlplane/control"

type Client struct {
	UID    control.ClientUID
	Status Status
}
