package otelcolext

import (
	"context"

	"github.com/neblic/platform/dataplane/metric"
	"github.com/neblic/platform/dataplane/sample"
	"github.com/neblic/platform/internal/pkg/exporter"
	"go.opentelemetry.io/collector/consumer"
)

type logsExporter struct {
	consumer consumer.Logs
}

func NewLogsExporter(c consumer.Logs) exporter.LogsExporter {
	return &logsExporter{
		consumer: c,
	}
}

func (e *logsExporter) Export(ctx context.Context, otlpLogs sample.OTLPLogs) error {
	return e.consumer.ConsumeLogs(ctx, otlpLogs.Logs())
}

func (e *logsExporter) Close(context.Context) error {
	return nil
}

type metricsExporter struct {
	consumer consumer.Metrics
}

func NewMetricsExporter(c consumer.Metrics) exporter.MetricsExporter {
	return &metricsExporter{
		consumer: c,
	}
}

func (e *metricsExporter) Export(ctx context.Context, otlpMetrics metric.Metrics) error {
	return e.consumer.ConsumeMetrics(ctx, otlpMetrics.Metrics())
}

func (e *metricsExporter) Close(context.Context) error {
	return nil
}
