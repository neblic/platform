package otelcolext

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/neblic/platform/controlplane/server"
	"github.com/neblic/platform/dataplane"
	"github.com/neblic/platform/logging"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/connector"
)

type neblicConnector struct {
	onceStart           sync.Once
	onceConfigureGlobal sync.Once
	onceShutdown        sync.Once

	cfg *Config

	controlPlane *server.Server
	dataPlane    *dataplane.Processor
}

func newNeblicConnector() *neblicConnector {
	return &neblicConnector{
		onceStart:    sync.Once{},
		onceShutdown: sync.Once{},
	}
}

func (n *neblicConnector) CreateGlobal(_ context.Context, set connector.CreateSettings, componentCfg component.Config) error {

	var err error
	n.onceConfigureGlobal.Do(func() {
		n.cfg = componentCfg.(*Config)

		// Create control plane
		controlPlaneOptions := []server.Option{}
		if n.cfg.StoragePath != "" {
			controlPlaneOptions = append(controlPlaneOptions, server.WithDiskStorage(n.cfg.StoragePath))
		}
		if n.cfg.TLSConfig != nil {
			controlPlaneOptions = append(controlPlaneOptions, server.WithTLS(n.cfg.TLSConfig.CertFile, n.cfg.TLSConfig.KeyFile))
		}
		if n.cfg.AuthConfig != nil {
			switch n.cfg.AuthConfig.Type {
			case "bearer":
				if n.cfg.AuthConfig.BearerConfig == nil {
					err = fmt.Errorf("bearer authentication enabled but token not configured")
				}
				controlPlaneOptions = append(controlPlaneOptions, server.WithAuthBearer(n.cfg.AuthConfig.BearerConfig.Token))
			case "":
				// nothing to do
			default:
				err = fmt.Errorf("invalid authentication type %s", n.cfg.AuthConfig.Type)
				return
			}
		}
		controlPlaneOptions = append(controlPlaneOptions, server.WithLogger(logging.FromZapLogger(set.Logger)))
		n.controlPlane, err = server.New(n.cfg.UID, controlPlaneOptions...)
		if err != nil {
			return
		}

		// Create data plane
		dataPlaneSettings := &dataplane.Settings{
			Logger:       logging.FromZapLogger(set.Logger),
			ControlPlane: n.controlPlane,
		}
		n.dataPlane = dataplane.NewProcessor(dataPlaneSettings)
	})

	return err
}

func (n *neblicConnector) Start(_ context.Context, _ component.Host) error {
	var err error
	n.onceStart.Do(func() {
		err = n.controlPlane.Start(n.cfg.Endpoint)
		if err != nil {
			return
		}

		err = n.dataPlane.Start()
		if err != nil {
			return
		}
	})

	return err
}

func (n *neblicConnector) Shutdown(_ context.Context) error {
	var errs error

	n.onceShutdown.Do(func() {
		if n.controlPlane != nil {
			err := n.controlPlane.Stop(time.Second)
			errs = errors.Join(errs, err)
		}

		if n.dataPlane != nil {
			err := n.dataPlane.Stop()
			errs = errors.Join(errs, err)
		}
	})

	return errs
}
