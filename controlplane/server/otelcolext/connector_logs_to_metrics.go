package otelcolext

import (
	"context"
	"errors"

	"github.com/neblic/platform/dataplane"
	"github.com/neblic/platform/dataplane/sample"
	"github.com/neblic/platform/internal/pkg/exporter"
	"github.com/neblic/platform/logging"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.uber.org/zap"
)

type logsToMetricsConnector struct {
	exporter  exporter.MetricsExporter
	logger    logging.Logger
	processor *dataplane.LogsToMetricsProcessor
}

func newLogsToMetricsConnector(_ *Config, logger *zap.Logger, nextConsumer consumer.Metrics) (*logsToMetricsConnector, error) {
	if nextConsumer == nil {
		return nil, component.ErrNilNextConsumer
	}

	return &logsToMetricsConnector{
		logger:   logging.FromZapLogger(logger),
		exporter: NewMetricsExporter(nextConsumer),
	}, nil
}

func (n *logsToMetricsConnector) Start(_ context.Context, _ component.Host) error {
	n.processor = dataplane.NewLogsToMetricsProcessor(n.logger)
	return nil
}

func (n *logsToMetricsConnector) Shutdown(_ context.Context) error {
	return nil
}

func (n *logsToMetricsConnector) Capabilities() consumer.Capabilities {
	return consumer.Capabilities{
		MutatesData: false,
	}
}

func (n *logsToMetricsConnector) ConsumeLogs(ctx context.Context, logs plog.Logs) error {
	otlpLogs := sample.OTLPLogsFrom(logs)
	otlpMetrics, errs := n.processor.Process(ctx, otlpLogs)

	err := n.exporter.Export(ctx, otlpMetrics)
	if err != nil {
		errs = errors.Join(errs, err)
	}

	return errs
}
