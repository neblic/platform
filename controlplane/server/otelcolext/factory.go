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
		connector.WithLogsToLogs(createLogsProcessor, component.StabilityLevelDevelopment),
	)
}

func createDefaultConfig() component.Config {
	return newDefaultSettings()
}

func createLogsProcessor(_ context.Context, set connector.CreateSettings, cfg component.Config, nextConsumer consumer.Logs) (connector.Logs, error) {
	return newLogsConnector(cfg.(*Config), set.Logger, nextConsumer)
}
