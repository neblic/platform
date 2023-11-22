package sampler

import (
	"time"

	"github.com/neblic/platform/controlplane/control"
	"github.com/neblic/platform/internal/pkg/exporter"
	"github.com/neblic/platform/sampler/defs"
)

const closeTimeout = time.Duration(2) * time.Second

type AuthBearerOptions struct {
	Token string
}

type AuthOptions struct {
	Type   string
	Bearer AuthBearerOptions
}

type Settings struct {
	Name     string
	Resource string
	Schema   defs.Schema

	ControlPlaneAddr string
	EnableTLS        bool
	Auth             AuthOptions

	SamplingIn    control.SamplingConfig
	InitialConfig control.SamplerConfigUpdate
	Exporter      exporter.Exporter

	UpdateStatsPeriod time.Duration

	ErrFwrder chan error
}
