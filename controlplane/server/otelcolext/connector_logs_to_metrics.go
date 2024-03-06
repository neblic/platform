package otelcolext

import (
	"context"

	"github.com/neblic/platform/dataplane/sample"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/plog"
)

type logsToMetricsConnector struct {
	*neblicConnector
}

func newLogsToMetricsConnector(neblicConnector *neblicConnector) *logsToMetricsConnector {
	return &logsToMetricsConnector{
		neblicConnector: neblicConnector,
	}
}

func (n *logsToMetricsConnector) Capabilities() consumer.Capabilities {
	return consumer.Capabilities{
		MutatesData: false,
	}
}

func (n *logsToMetricsConnector) ConsumeLogs(ctx context.Context, logs plog.Logs) error {
	n.dataPlane.TranslateStreamNamesToUIDs(sample.OTLPLogsFrom(logs))
	metrics, err := n.dataPlane.ComputeMetrics(sample.OTLPLogsFrom(logs))
	if err != nil {
		return err
	}
	n.dataPlane.MetricExporter.Export(ctx, metrics)

	return nil
}
