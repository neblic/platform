package storage

import "github.com/neblic/platform/controlplane/control"

type SamplerEntry struct {
	Resource     string
	Name         string
	Config       control.SamplerConfig
	Capabilities control.Capabilities
}
