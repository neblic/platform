package sampler

import (
	"time"

	"github.com/neblic/platform/sampler/defs"
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

type Options struct {
	Name     string
	Resource string
	Schema   defs.Schema

	ControlPlaneAddr string
	EnableTLS        bool
	Auth             AuthOptions

	LimiterInLimit  int64
	Exporter        exporter.Exporter
	LimiterOutLimit int64

	UpdateStatsPeriod time.Duration
}
