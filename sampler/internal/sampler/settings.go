package sampler

import (
	"fmt"
	"time"

	"github.com/neblic/platform/controlplane/control"
	"github.com/neblic/platform/internal/pkg/exporter"
	"github.com/neblic/platform/sampler/sample"
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
	Schema   sample.Schema

	ControlPlaneAddr string
	EnableTLS        bool
	Auth             AuthOptions

	SamplingIn    control.SamplingConfig
	InitialConfig control.SamplerConfigUpdate
	Tags          []control.Tag
	Exporter      exporter.Exporter

	UpdateStatsPeriod time.Duration

	ErrFwrder chan error
}

func (s *Settings) String() string {
	redactedSettings := *s
	redactedSettings.Auth.Bearer = AuthBearerOptions{Token: "REDACTED"}
	return fmt.Sprintf("%+v", redactedSettings)
}
