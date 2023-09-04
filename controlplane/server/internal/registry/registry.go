package registry

import (
	"fmt"

	"github.com/neblic/platform/controlplane/server/internal/registry/storage"
	"github.com/neblic/platform/logging"
)

type Registry struct {
	Client  *ClientRegistry
	Sampler *SamplerRegistry
}

func NewRegistry(logger logging.Logger, notifyDirty chan struct{}, storageOpts storage.Options) (*Registry, error) {
	clientRegistry, err := NewClientRegistry(logger)
	if err != nil {
		return nil, fmt.Errorf("error initializing client registry: %w", err)
	}

	samplerRegistry, err := NewSamplerRegistry(logger, notifyDirty, storageOpts)
	if err != nil {
		return nil, fmt.Errorf("error initializing sampler registry: %w", err)
	}

	return &Registry{
		Client:  clientRegistry,
		Sampler: samplerRegistry,
	}, nil
}
