package otelcolext

import (
	"context"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/connector"
	"go.opentelemetry.io/collector/consumer"
)

const (
	// The value of connector "type" in configuration.
	typeStr = "neblic"
)

func NewFactory() connector.Factory {
	neblicConnector := newNeblicConnector()
	logsToLogsConnector := newLogsToLogsConnector(neblicConnector)
	logsToMetricsConnector := newLogsToMetricsConnector(neblicConnector)

	logsToLogsCreator := func(ctx context.Context, set connector.CreateSettings, cfg component.Config, nextConsumer consumer.Logs) (connector.Logs, error) {
		err := neblicConnector.CreateGlobal(ctx, set, cfg)
		if err != nil {
			return nil, err
		}

		if nextConsumer == nil {
			return nil, component.ErrNilNextConsumer
		}

		neblicConnector.dataPlane.SampleExporter = NewLogsExporter(nextConsumer)
		neblicConnector.dataPlane.DigestExporter = NewLogsExporter(logsToMetricsConnector, nextConsumer)
		neblicConnector.dataPlane.EventExporter = NewLogsExporter(logsToMetricsConnector, nextConsumer)

		return logsToLogsConnector, nil
	}

	logsToMetricsCreator := func(ctx context.Context, set connector.CreateSettings, cfg component.Config, nextConsumer consumer.Metrics) (connector.Logs, error) {
		err := neblicConnector.CreateGlobal(ctx, set, cfg)
		if err != nil {
			return nil, err
		}

		if nextConsumer == nil {
			return nil, component.ErrNilNextConsumer
		}

		neblicConnector.dataPlane.MetricExporter = NewMetricsExporter(nextConsumer)

		return logsToMetricsConnector, nil
	}

	return connector.NewFactory(
		typeStr,
		createDefaultConfig,
		connector.WithLogsToLogs(logsToLogsCreator, component.StabilityLevelDevelopment),
		connector.WithLogsToMetrics(logsToMetricsCreator, component.StabilityLevelDevelopment),
	)
}

func createDefaultConfig() component.Config {
	return newDefaultSettings()
}
