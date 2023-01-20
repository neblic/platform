package sampler

import (
	"time"

	"github.com/neblic/platform/sampler/defs"
	"github.com/neblic/platform/sampler/internal/sample/exporter"
)

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

	Exporter  exporter.Exporter
	RateLimit int64
	RateBurst int64

	UpdateStatsPeriod time.Duration
}
