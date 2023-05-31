package otlp

import (
	"context"
	"fmt"

	dpsample "github.com/neblic/platform/dataplane/sample"
	"github.com/neblic/platform/logging"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config/configopaque"
	"go.opentelemetry.io/collector/config/configtls"
	collectorexporter "go.opentelemetry.io/collector/exporter"
	"go.opentelemetry.io/collector/exporter/otlpexporter"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/otel/metric/noop"
	"go.opentelemetry.io/otel/trace"
)

type Exporter struct { // implements sample.Exporter
	logsExporter collectorexporter.Logs
}

func New(ctx context.Context, logger logging.Logger, exportServerAddr string, opts *Options) (*Exporter, error) {
	if opts == nil {
		opts = newDefaultOptions()
	}

	factory := otlpexporter.NewFactory()
	cfg := factory.CreateDefaultConfig().(*otlpexporter.Config)
	cfg.GRPCClientSettings.Endpoint = exportServerAddr
	cfg.GRPCClientSettings.TLSSetting = configtls.TLSClientSetting{
		Insecure: !opts.TLSEnable,
	}

	// provider already checked that it is a valid type
	switch opts.Auth.Type {
	case "bearer":
		cfg.Headers["authorization"] = configopaque.String(fmt.Sprintf("Bearer %s", opts.Auth.Bearer.Token))
	}

	settings := collectorexporter.CreateSettings{
		TelemetrySettings: component.TelemetrySettings{
			Logger: logger.ZapLogger(),
			// The Exporter doesn't need to generate traces or metrics
			TracerProvider: trace.NewNoopTracerProvider(),
			MeterProvider:  noop.NewMeterProvider(),
		},
	}

	logsExporter, err := factory.CreateLogsExporter(ctx, settings, cfg)
	if err != nil {
		return nil, fmt.Errorf("couldn't create ProviderOTLP logs exporter: %w", err)
	}

	err = logsExporter.Start(ctx, &Host{})
	if err != nil {
		return nil, fmt.Errorf("couldn't start ProviderOTLP logs exporter: %w", err)
	}

	return &Exporter{
		logsExporter: logsExporter,
	}, nil
}

// Export internally perform samples batches
func (e *Exporter) Export(ctx context.Context, resourceSamples []dpsample.SamplerSamples) error {
	logs := dpsample.SamplesToOTLPLogs(resourceSamples)

	return e.exportLogs(ctx, logs)
}

func (e *Exporter) exportLogs(ctx context.Context, ld plog.Logs) error {
	err := e.logsExporter.ConsumeLogs(ctx, ld)
	if err != nil {
		return fmt.Errorf("error sending logs: %w", err)
	}

	return nil
}

func (e *Exporter) Close(ctx context.Context) error {
	if err := e.logsExporter.Shutdown(ctx); err != nil {
		return fmt.Errorf("error shutting down ProviderOTLP logs exporter: %w", err)
	}

	return nil
}
