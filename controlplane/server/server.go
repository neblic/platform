package server

import (
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/neblic/platform/controlplane/control"
	"github.com/neblic/platform/controlplane/event"
	"github.com/neblic/platform/controlplane/protos"
	"github.com/neblic/platform/controlplane/server/internal/auth"
	protocolclient "github.com/neblic/platform/controlplane/server/internal/protocol/client"
	protocolsampler "github.com/neblic/platform/controlplane/server/internal/protocol/sampler"
	"github.com/neblic/platform/controlplane/server/internal/registry"
	"github.com/neblic/platform/logging"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
)

var (
	ErrAlreadyStarted = errors.New("server already started")
)

// keepAliveMinPeriod is set to the minimum period value accepted by grpc so
// the server never disconnects clients regardless of their keep alive settings.
const keepAliveMinPeriod = time.Duration(10) * time.Second

type Server struct {
	uid string

	lis        net.Listener
	grpcServer *grpc.Server
	protos.UnimplementedControlPlaneServer

	clientRegistry  *registry.ClientRegistry
	samplerRegistry *registry.SamplerRegistry
	opts            *options

	reconcileNow        chan struct{}
	reconciliationTimer *time.Ticker

	logger logging.Logger
}

var _ protos.ControlPlaneServer = (*Server)(nil)

func New(uid string, serverOptions ...Option) (*Server, error) {
	opts := newDefaultOptions()
	for _, opt := range serverOptions {
		opt.apply(opts)
	}

	s := &Server{
		uid:          uid,
		opts:         opts,
		reconcileNow: make(chan struct{}, 1),
		logger:       opts.logger,
	}

	// Initialize client registry
	var err error
	s.clientRegistry, err = registry.NewClientRegistry(s.logger)
	if err != nil {
		return nil, fmt.Errorf("error initializing client registry: %v", err)
	}

	// Initialize sampler registry
	s.samplerRegistry, err = registry.NewSamplerRegistry(s.logger, s.reconcileNow, *opts.storage)
	if err != nil {
		return nil, fmt.Errorf("error initializing sampler registry: %v", err)
	}

	s.reconciliationTimer = time.NewTicker(opts.reconciliationPeriod)
	go s.reconcileConfigLoop()

	return s, nil
}

func (s *Server) Start(listenAddr string) error {
	if s.grpcServer != nil {
		return ErrAlreadyStarted
	}

	var options []grpc.ServerOption
	options = append(options,
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
			MinTime: keepAliveMinPeriod,
		}),
	)

	if s.opts.tls.enable {
		creds, err := credentials.NewServerTLSFromFile(s.opts.tls.certPath, s.opts.tls.certKeyPath)
		if err != nil {
			return fmt.Errorf("error building gRPC server credentials: %w", err)
		}

		options = append(options, grpc.Creds(creds))
	}

	switch s.opts.auth.authType {
	case "bearer":
		options = append(options,
			grpc.StreamInterceptor(auth.NewAuthBearerStreamInterceptor(s.opts.auth.bearer.token)),
		)
	case "":
		// nothing to do
	default:
		return fmt.Errorf("invalid authentication type %s", s.opts.auth.authType)
	}

	grpcServer := grpc.NewServer(options...)
	protos.RegisterControlPlaneServer(grpcServer, s)

	lis, err := net.Listen("tcp", listenAddr)
	if err != nil {
		return fmt.Errorf("error listening at %s: %w", listenAddr, err)
	}

	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			s.logger.Error("error serving at %s: %w", listenAddr, err)

			if err := lis.Close(); err != nil {
				s.logger.Debug("error closing socket: %w", err)
			}
		}
	}()

	s.lis = lis
	s.grpcServer = grpcServer

	return nil
}

func (s *Server) Addr() net.Addr {
	return s.lis.Addr()
}

func (s *Server) GetSamplers(resourceName, samplerName string) ([]*control.Sampler, error) {
	serverSampler, err := s.samplerRegistry.GetSampler(resourceName, samplerName)
	if err != nil {
		return nil, err
	}

	samplers := []*control.Sampler{}
	for _, instance := range serverSampler.Instances {
		sampler := control.NewSampler(serverSampler.Name, serverSampler.Resource, instance.UID)
		sampler.Tags = serverSampler.Tags
		sampler.Config = instance.Sampler.Config
		sampler.SamplingStats = instance.Stats
		samplers = append(samplers, sampler)
	}

	return samplers, nil
}

func (s *Server) Events() chan event.Event {
	return s.samplerRegistry.Events()
}

func (s *Server) UpdateSamplerStats(resourceName, samplerName string, collectedSamples int64) error {
	return s.samplerRegistry.UpdateSamplerStats(resourceName, samplerName, collectedSamples)
}

func (s *Server) SamplerConn(stream protos.ControlPlane_SamplerConnServer) error {
	h := protocolsampler.New(s.logger, s.uid, s.samplerRegistry, s.opts.stream)

	return h.HandleStream(stream)
}

func (s *Server) ClientConn(stream protos.ControlPlane_ClientConnServer) error {
	h := protocolclient.New(s.logger, s.uid, s.clientRegistry, s.samplerRegistry, s.opts.stream)

	return h.HandleStream(stream)
}

func (s *Server) Stop(_ time.Duration) error {
	if s.grpcServer != nil {
		defer func() { s.grpcServer = nil }()

		// Just closing the server is enough to have all samplers and clients eventually
		// deregister when they detect that the connection has been terminated.
		//
		// A cleaner approach would be to notify all registered samplers and clients before closing
		// their connections, so they know for sure that the server is gone and it is not a transient
		// disconnection.
		s.grpcServer.Stop()
	}

	if s.samplerRegistry != nil {
		s.samplerRegistry.Close()
		s.samplerRegistry = nil
	}

	return nil
}

func (s *Server) reconcileSamplerConfigs() {
	start := time.Now()
	configsUpdated := 0

	instances := s.samplerRegistry.GetRegisteredInstances()
	for _, instance := range instances {
		if instance.Dirty {
			if err := instance.Conn.Configure(&instance.Sampler.Config); err != nil {
				s.logger.Error("Error configuring sampler", "error", err)
			}
			instance.Dirty = false
			configsUpdated++
		}
	}

	s.logger.Debug("Configuration reconciliation performed", "elapsed", time.Since(start).String(), "configs_updated", configsUpdated)
}

func (s *Server) reconcileConfigLoop() {
	for {
		select {
		case <-s.reconcileNow:
			s.reconcileSamplerConfigs()
		case <-s.reconciliationTimer.C:
			s.reconcileSamplerConfigs()
		}
	}
}
