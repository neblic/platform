package exporter

import (
	"context"

	"github.com/neblic/platform/dataplane/metric"
	dpsample "github.com/neblic/platform/dataplane/sample"
)

type LogsExporter interface {
	Export(context.Context, dpsample.OTLPLogs) error
	Close(context.Context) error
}

type MetricsExporter interface {
	Export(context.Context, metric.Metrics) error
	Close(context.Context) error
}
