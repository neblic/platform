package otelcolext

import (
	"context"

	dpsample "github.com/neblic/platform/dataplane/sample"
	"go.opentelemetry.io/collector/consumer"
)

type Exporter struct {
	consumer consumer.Logs
}

func NewExporter(c consumer.Logs) *Exporter {
	return &Exporter{
		consumer: c,
	}
}

func (e *Exporter) Export(ctx context.Context, otlpLogs dpsample.OTLPLogs) error {
	return e.consumer.ConsumeLogs(ctx, otlpLogs.Logs())
}

func (e *Exporter) Close(context.Context) error {
	return nil
}
