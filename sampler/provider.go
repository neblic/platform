package sampler

import (
	"context"
	"fmt"

	"github.com/neblic/platform/logging"
	"github.com/neblic/platform/sampler/defs"
	exporterotlp "github.com/neblic/platform/sampler/internal/sample/exporter/otlp"
	"github.com/neblic/platform/sampler/internal/sampler"
)

// Provider defines a sampler provider object capable of creating new samplers with a common configuration
type Provider struct {
	settings       Settings
	opts           *options
	sampleExporter *exporterotlp.Exporter
	logger         logging.Logger
}

// NewProvider creates a new sampler provider capable of creating new samplers.
// After initializing a provider, it is recommended to set it as the global provider for easier access.
//
//	provider := NewProvider(...)
//	global.SetSamplerProvider(provider)
//
// Then, to create new samplers do:
//
//	sampler, err := global.SamplerProvider().Sampler(...)
//
// or simply:
//
//	sampler, err := Sampler(...)
func NewProvider(ctx context.Context, settings Settings, opts ...Option) (defs.Provider, error) {
	setOpts := newDefaultOptions()
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
		return nil, fmt.Errorf("Invalid authorization type %s", setOpts.dataServerAuth.authType)
	}

	sampleExporter, err := exporterotlp.New(ctx, setOpts.logger, settings.DataServerAddr, exporterOpts)
	if err != nil {
		return nil, fmt.Errorf("couldn't initialize the OTLP samples exporter")
	}

	return &Provider{
		settings: settings,
		opts:     setOpts,

		sampleExporter: sampleExporter,
		logger:         setOpts.logger,
	}, nil
}

// Sampler creates a new sampler. See the interface comments for more details.
func (p *Provider) Sampler(name string, schema defs.Schema) (defs.Sampler, error) {
	samplerOpts := &sampler.Options{
		Name:     name,
		Resource: p.settings.ResourceName,
		Schema:   schema,

		ControlPlaneAddr: p.settings.ControlServerAddr,
		EnableTLS:        p.opts.controlServerTLSEnable,

		Exporter:  p.sampleExporter,
		RateLimit: p.opts.samplingRateLimit,
		RateBurst: p.opts.samplingRateBurst,

		UpdateStatsPeriod: p.opts.updateStatsPeriod,
	}

	switch p.opts.controlServerAuth.authType {
	case "bearer":
		samplerOpts.Auth.Type = "bearer"
		samplerOpts.Auth.Bearer.Token = p.opts.controlServerAuth.bearer.token
	case "":
		// nothing to do
	default:
		return nil, fmt.Errorf("Invalid authorization type %s", p.opts.controlServerAuth.authType)
	}

	return sampler.New(samplerOpts, p.logger)
}
