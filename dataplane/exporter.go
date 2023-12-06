package dataplane

import (
	"context"

	"github.com/neblic/platform/dataplane/sample"
	"go.opentelemetry.io/collector/consumer"
)

type Exporter interface {
	Export(ctx context.Context, otlpLogs sample.OTLPLogs) error
	Close(context.Context) error
}

type OTLPExporter struct {
	consumer consumer.Logs
}

func NewOTLPExporter(c consumer.Logs) *OTLPExporter {
	return &OTLPExporter{
		consumer: c,
	}
}

func (e *OTLPExporter) Export(ctx context.Context, otlpLogs sample.OTLPLogs) error {
	return e.consumer.ConsumeLogs(ctx, otlpLogs.Logs())
}

func (e *OTLPExporter) Close(context.Context) error {
	return nil
}
