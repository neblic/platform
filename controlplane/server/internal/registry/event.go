package registry

import "github.com/neblic/platform/controlplane/control"

type Type uint

const (
	ClientType = iota
	SamplerType
)

type Operation uint

const (
	UpsertOperation = iota
	DeleteOperation
)

type SamplerRegistryEvent struct {
	Resource string
	Sampler  string
	Config   *control.SamplerConfig
}

type Event struct {
	Operation
	RegistryType         Type
	SamplerRegistryEvent *SamplerRegistryEvent
}
