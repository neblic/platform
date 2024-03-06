package otelcolext

import (
	"context"

	"github.com/neblic/platform/dataplane/sample"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/plog"
)

type logsToLogsConnector struct {
	*neblicConnector
}

func newLogsToLogsConnector(neblicConnector *neblicConnector) (*logsToLogsConnector, error) {
	return &logsToLogsConnector{
		neblicConnector: neblicConnector,
	}, nil
}

func (n *logsToLogsConnector) Capabilities() consumer.Capabilities {
	return consumer.Capabilities{
		MutatesData: true,
	}
}

func (n *logsToLogsConnector) ConsumeLogs(ctx context.Context, logs plog.Logs) error {
	otlpLogs := sample.OTLPLogsFrom(logs)

	n.dataPlane.TranslateStreamNamesToUIDs(otlpLogs)
	n.dataPlane.UpdateStats(otlpLogs)
	n.dataPlane.ComputeDigests(otlpLogs)
	n.dataPlane.ComputeEvents(otlpLogs)

	n.dataPlane.SampleExporter.Export(ctx, otlpLogs)
	return nil
}
