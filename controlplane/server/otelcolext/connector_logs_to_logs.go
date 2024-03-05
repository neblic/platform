package otelcolext

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/neblic/platform/controlplane/server"
	"github.com/neblic/platform/dataplane"
	"github.com/neblic/platform/dataplane/sample"
	"github.com/neblic/platform/internal/pkg/exporter"
	"github.com/neblic/platform/logging"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.uber.org/zap"
)

type logsToLogsConnector struct {
	cfg      *Config
	exporter exporter.LogsExporter

	logger     logging.Logger
	s          *server.Server
	serverOpts []server.Option
	processor  *dataplane.Processor
}

func newLogsToLogsConnector(cfg *Config, logger *zap.Logger, nextConsumer consumer.Logs) (*logsToLogsConnector, error) {
	serverOpts := []server.Option{}

	if cfg.StoragePath != "" {
		serverOpts = append(serverOpts, server.WithDiskStorage(cfg.StoragePath))
	}

	if cfg.TLSConfig != nil {
		serverOpts = append(serverOpts, server.WithTLS(cfg.TLSConfig.CertFile, cfg.TLSConfig.KeyFile))
	}

	if cfg.AuthConfig != nil {
		switch cfg.AuthConfig.Type {
		case "bearer":
			if cfg.AuthConfig.BearerConfig == nil {
				return nil, fmt.Errorf("bearer authentication enabled but token not configured")
			}
			serverOpts = append(serverOpts, server.WithAuthBearer(cfg.AuthConfig.BearerConfig.Token))
		case "":
			// nothing to do
		default:
			return nil, fmt.Errorf("invalid authentication type %s", cfg.AuthConfig.Type)
		}
	}

	serverOpts = append(serverOpts, server.WithLogger(logging.FromZapLogger(logger)))

	if nextConsumer == nil {
		return nil, component.ErrNilNextConsumer
	}

	return &logsToLogsConnector{
		cfg:        cfg,
		logger:     logging.FromZapLogger(logger),
		exporter:   NewLogsExporter(nextConsumer),
		serverOpts: serverOpts,
	}, nil
}

func (n *logsToLogsConnector) Start(_ context.Context, _ component.Host) error {
	var err error
	n.s, err = server.New(n.cfg.UID, n.serverOpts...)
	if err != nil {
		return err
	}
	err = n.s.Start(n.cfg.Endpoint)
	if err != nil {
		return err
	}

	n.processor = dataplane.NewProcessor(n.logger, n.s, n.exporter)
	err = n.processor.Start()
	if err != nil {
		return err
	}

	return nil
}

func (n *logsToLogsConnector) Shutdown(_ context.Context) error {
	var errs error
	if n.s != nil {
		err := n.s.Stop(time.Second)
		errs = errors.Join(errs, err)
	}

	if n.processor != nil {
		err := n.processor.Stop()
		errs = errors.Join(errs, err)
	}

	return errs
}

func (n *logsToLogsConnector) Capabilities() consumer.Capabilities {
	return consumer.Capabilities{
		MutatesData: true,
	}
}

func (n *logsToLogsConnector) ConsumeLogs(ctx context.Context, logs plog.Logs) error {
	otlpLogs := sample.OTLPLogsFrom(logs)
	err := n.processor.Process(ctx, otlpLogs)

	return err
}
