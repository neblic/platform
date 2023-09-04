package otelcolext

import (
	"context"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/processor"
)

const (
	// The value of processor "type" in configuration.
	typeStr = "neblic"
)

func NewFactory() processor.Factory {
	return processor.NewFactory(
		typeStr,
		createDefaultConfig,
		processor.WithLogs(createLogsProcessor, component.StabilityLevelDevelopment),
	)
}

func createDefaultConfig() component.Config {
	return newDefaultSettings()
}

// func(context.Context, CreateSettings, component.Config, consumer.Logs) (Logs, error)
func createLogsProcessor(_ context.Context, set processor.CreateSettings, cfg component.Config, nextConsumer consumer.Logs) (processor.Logs, error) {
	return newLogsProcessor(cfg.(*Config), set.Logger, nextConsumer)
}
