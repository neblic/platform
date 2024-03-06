package otelcolext

import (
	"context"
	"errors"

	"github.com/neblic/platform/dataplane/metric"
	"github.com/neblic/platform/dataplane/sample"
	"github.com/neblic/platform/internal/pkg/exporter"
	"go.opentelemetry.io/collector/consumer"
)

type logsExporter struct {
	consumers []consumer.Logs
}

func NewLogsExporter(consumers ...consumer.Logs) exporter.LogsExporter {
	return &logsExporter{
		consumers: consumers,
	}
}

func (e *logsExporter) Export(ctx context.Context, otlpLogs sample.OTLPLogs) error {
	var errs error
	for _, consumer := range e.consumers {
		err := consumer.ConsumeLogs(ctx, otlpLogs.Logs())
		if err != nil {
			errors.Join(errs, err)
		}
	}
	return errs
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
