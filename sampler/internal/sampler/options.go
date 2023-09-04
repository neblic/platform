package sampler

import (
	"time"

	"github.com/neblic/platform/controlplane/control"
	"github.com/neblic/platform/internal/pkg/rule"
	"github.com/neblic/platform/sampler/internal/sample/exporter"
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
	Schema   rule.Schema

	ControlPlaneAddr string
	EnableTLS        bool
	Auth             AuthOptions

	LimiterIn  control.LimiterConfig
	SamplingIn control.SamplingConfig
	Exporter   exporter.Exporter
	LimiterOut control.LimiterConfig

	UpdateStatsPeriod time.Duration

	ErrFwrder chan error
}
