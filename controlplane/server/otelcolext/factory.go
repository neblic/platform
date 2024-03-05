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
	return connector.NewFactory(
		typeStr,
		createDefaultConfig,
		connector.WithLogsToLogs(createLogsToLogsConnector, component.StabilityLevelDevelopment),
		connector.WithLogsToMetrics(createLogsToMetricsConnector, component.StabilityLevelDevelopment),
	)
}

func createDefaultConfig() component.Config {
	return newDefaultSettings()
}

func createLogsToLogsConnector(_ context.Context, set connector.CreateSettings, cfg component.Config, nextConsumer consumer.Logs) (connector.Logs, error) {
	return newLogsToLogsConnector(cfg.(*Config), set.Logger, nextConsumer)
}

func createLogsToMetricsConnector(_ context.Context, set connector.CreateSettings, cfg component.Config, nextConsumer consumer.Metrics) (connector.Logs, error) {
	return newLogsToMetricsConnector(cfg.(*Config), set.Logger, nextConsumer)
}
