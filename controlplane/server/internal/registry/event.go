package registry

import "github.com/neblic/platform/controlplane/control"

type RegistryType uint

const (
	ClientRegistryType = iota
	SamplerRegistryType
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
	RegistryType         RegistryType
	SamplerRegistryEvent *SamplerRegistryEvent
}
