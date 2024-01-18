package sampler

import (
	"context"
	"fmt"

	"github.com/neblic/platform/internal/pkg/exporter"
	"github.com/neblic/platform/logging"
	exporterotlp "github.com/neblic/platform/sampler/internal/sample/exporter/otlp"
	"github.com/neblic/platform/sampler/internal/sampler"
	"github.com/neblic/platform/sampler/sample"
)

type Settings struct {
	// ResourceName sets the resource name common to all samplers created by this provider
	// For example, the service or the data pipeline operator name
	ResourceName string
	// ControlServerAddr specifies the address where the control server is listening at.
	// Format addr:port
	ControlServerAddr string
	// ControlServerAddr specifies the address where the data server is listening at.
	// By default, it sends the data samples encoded as OTLP logs following the OTLP gRPC
	// protocol, so it works with any OTLP gRPC logs compatible collector.
	// Format addr:port
	DataServerAddr string
}

type authBearerOption struct {
	token string
}

type authOption struct {
	authType string
	bearer   authBearerOption
}

type providerOptions struct {
	controlServerTLSEnable bool
	controlServerAuth      authOption

	dataServerTLSEnable bool
	dataServerAuth      authOption

	logger      logging.Logger
	samplersErr chan error
}

func newDefaultProviderOptions() *providerOptions {
	return &providerOptions{
		controlServerTLSEnable: false,
		dataServerTLSEnable:    false,
	}
}

type ProviderOption interface {
	apply(*providerOptions)
}

type funcProviderOption struct {
	f func(*providerOptions)
}

func (fco *funcProviderOption) apply(co *providerOptions) {
	fco.f(co)
}

func newFuncPoviderOption(f func(*providerOptions)) *funcProviderOption {
	return &funcProviderOption{
		f: f,
	}
}

// WithTLS enables TLS connections with the server
func WithTLS() ProviderOption {
	return newFuncPoviderOption(func(o *providerOptions) {
		o.controlServerTLSEnable = true
		o.dataServerTLSEnable = true
	})
}

// WithBearerAuth sets authorization based on a Bearer token
func WithBearerAuth(token string) ProviderOption {
	return newFuncPoviderOption(func(o *providerOptions) {
		o.controlServerAuth.authType = "bearer"
		o.controlServerAuth.bearer.token = token

		o.dataServerAuth.authType = "bearer"
		o.dataServerAuth.bearer.token = token
	})
}

// WithLogger provides a logger instance to log the Provider and Sampler
// activity
func WithLogger(l logging.Logger) ProviderOption {
	return newFuncPoviderOption(func(o *providerOptions) {
		o.logger = l
	})
}

// WithErrorChannel received a channel where Sampler errors will be sent.
// The avoid blocking the Sampler, it won't block if the channel is full so it is responsibility of
// the provider to ensure the channel has enough buffer to avoid losing errors.
func WithSamplerErrorChannel(errCh chan error) ProviderOption {
	return newFuncPoviderOption(func(o *providerOptions) {
		o.samplersErr = errCh
	})
}

// Provider defines a sampler provider object capable of creating new samplers.
type Provider interface {
	// Sampler creates a new sampler with the specified schema. It currently supports Dynamic and Proto schemas.
	// * A Dynamic schema does not enforce any structure to the sampled data and is compatible with all the Sample*()
	// methods. The downside, is that it is slower than the Proto schema since it needs to determine at runtime the sampled
	// data format.
	// * A Proto schema requires the caller to provide a proto message (type proto.Message) to define the sampler schema.
	// All sampled data is expected to be provided as proto messages with the sampler.SampleProto() method, and it should
	// be the same type as the one provided when defining the sampler schema.
	Sampler(name string, schema sample.Schema, opts ...Option) (Sampler, error)
}

// Provider defines a sampler provider object capable of creating new samplers with a common configuration
type provider struct {
	settings    Settings
	opts        *providerOptions
	samplersErr chan error

	sampleExporter exporter.Exporter
	logger         logging.Logger
}

// NewProvider creates a new sampler provider capable of creating new samplers.
// After initializing a provider, it is recommended to set it as the global provider for easier access.
//
//	provider := NewProvider(...)
//	sampler.SetProvider(provider)
//
// Then, to create new samplers do:
//
//	sampler, err := sampler.New(...)
//
// or simply:
//
//	sampler, err := Sampler(...)
func NewProvider(ctx context.Context, settings Settings, opts ...ProviderOption) (Provider, error) {
	setOpts := newDefaultProviderOptions()
	for _, opt := range opts {
		opt.apply(setOpts)
	}

	if setOpts.logger == nil {
		setOpts.logger = logging.NewNopLogger()
	}

	exporterOpts := &exporterotlp.Options{
		TLSEnable: setOpts.dataServerTLSEnable,
	}

	switch setOpts.dataServerAuth.authType {
	case "bearer":
		exporterOpts.Auth.Type = "bearer"
		exporterOpts.Auth.Bearer.Token = setOpts.dataServerAuth.bearer.token
	case "":
		// nothing to do
	default:
		return nil, fmt.Errorf("invalid authorization type %s", setOpts.dataServerAuth.authType)
	}

	sampleExporter, err := exporterotlp.New(ctx, setOpts.logger, settings.DataServerAddr, exporterOpts)
	if err != nil {
		return nil, fmt.Errorf("couldn't initialize the OTLP samples exporter: %v", err)
	}

	return &provider{
		settings:    settings,
		opts:        setOpts,
		samplersErr: setOpts.samplersErr,

		sampleExporter: sampleExporter,
		logger:         setOpts.logger,
	}, nil
}

// Sampler creates a new sampler. See the interface comments for more details.
func (p *provider) Sampler(name string, schema sample.Schema, opts ...Option) (Sampler, error) {
	setOpts := newDefaultOptions()
	for _, opt := range opts {
		opt.apply(setOpts)
	}

	samplerOpts := &sampler.Settings{
		Name:     name,
		Resource: p.settings.ResourceName,
		Schema:   schema,

		ControlPlaneAddr: p.settings.ControlServerAddr,
		EnableTLS:        p.opts.controlServerTLSEnable,

		InitialConfig: setOpts.initialConfig,
		Exporter:      p.sampleExporter,

		UpdateStatsPeriod: setOpts.updateStatsPeriod,

		ErrFwrder: p.samplersErr,
	}

	switch p.opts.controlServerAuth.authType {
	case "bearer":
		samplerOpts.Auth.Type = "bearer"
		samplerOpts.Auth.Bearer.Token = p.opts.controlServerAuth.bearer.token
	case "":
		// nothing to do
	default:
		return nil, fmt.Errorf("invalid authorization type %s", p.opts.controlServerAuth.authType)
	}

	return sampler.New(samplerOpts, p.logger)
}
