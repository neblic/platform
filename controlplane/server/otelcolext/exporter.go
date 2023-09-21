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

func (e *Exporter) Export(ctx context.Context, logs []dpsample.SamplerSamples) error {
	return e.consumer.ConsumeLogs(ctx, dpsample.SamplesToOTLPLogs(logs))
}

func (e *Exporter) Close(context.Context) error {
	return nil
}
