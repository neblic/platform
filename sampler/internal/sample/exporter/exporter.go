package exporter

import (
	"context"

	dpsample "github.com/neblic/platform/dataplane/sample"
)

type Exporter interface {
	Export(context.Context, []dpsample.SamplerSamples) error
	Close(context.Context) error
}
