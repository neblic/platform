package exporter

import (
	"context"

	"github.com/neblic/platform/sampler/internal/sample"
)

type Exporter interface {
	Export(context.Context, []sample.ResourceSamples) error
}
