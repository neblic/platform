package sampler

import (
	"time"

	"github.com/google/uuid"
	"github.com/neblic/platform/controlplane/control"
	"github.com/neblic/platform/logging"
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

type options struct {
	controlServerTLSEnable bool
	dataServerTLSEnable    bool

	controlServerAuth authOption
	dataServerAuth    authOption

	samplingIn    control.SamplingConfig
	initialConfig control.SamplerConfigUpdate

	updateStatsPeriod time.Duration

	logger      logging.Logger
	samplersErr chan error
}

func newDefaultOptions() *options {
	initialConfig := control.NewSamplerConfigUpdate()
	initialConfig.LimiterIn = &control.LimiterConfig{Limit: 100}
	initialConfig.LimiterOut = &control.LimiterConfig{Limit: 10}
	streamUID := control.SamplerStreamUID(uuid.NewString())
	initialConfig.StreamUpdates = []control.StreamUpdate{
		{
			Op: control.StreamUpsert,
			Stream: control.Stream{
				UID:  streamUID,
				Name: "all",
				StreamRule: control.Rule{
					Lang:       control.SrlCel,
					Expression: "true",
				},
			},
		},
	}
	initialConfig.DigestUpdates = []control.DigestUpdate{
		{
			Op: control.DigestUpsert,
			Digest: control.Digest{
				UID:         control.SamplerDigestUID(uuid.NewString()),
				Name:        "all",
				StreamUID:   streamUID,
				FlushPeriod: time.Second * time.Duration(60),
				Type:        control.DigestTypeSt,
				St: control.DigestSt{
					MaxProcessedFields: int(100),
				},
			},
		},
	}

	return &options{
		controlServerTLSEnable: false,
		dataServerTLSEnable:    false,

		samplingIn:    control.SamplingConfig{},
		initialConfig: initialConfig,

		updateStatsPeriod: time.Second * time.Duration(5),
	}
}

type Option interface {
	apply(*options)
}

type funcOption struct {
	f func(*options)
}

func (fco *funcOption) apply(co *options) {
	fco.f(co)
}

func newFuncOption(f func(*options)) *funcOption {
	return &funcOption{
		f: f,
	}
}

// WithTLS enables TLS connections with the server
func WithTLS() Option {
	return newFuncOption(func(o *options) {
		o.controlServerTLSEnable = true
		o.dataServerTLSEnable = true
	})
}

// WithBearerAuth sets authorization based on a Bearer token
func WithBearerAuth(token string) Option {
	return newFuncOption(func(o *options) {
		o.controlServerAuth.authType = "bearer"
		o.controlServerAuth.bearer.token = token

		o.dataServerAuth.authType = "bearer"
		o.dataServerAuth.bearer.token = token
	})
}

// WithLimiterInLimit establishes the initial limiter in rate limit
// in samples per second
//
// Deprecated: Use WithInitialLimiterInLimit instead
func WithLimiterInLimit(l int32) Option {
	return WithInitialLimiterInLimit(l)
}

// WithInitialLimiterInLimit sets the initial limiter in rate limit. This configuration
// is only used the first time a sampler is registered with a server, posterior executions
// will use the configuration stored in the server and the provided configuration will be
// ignored.
func WithInitialLimiterInLimit(l int32) Option {
	return newFuncOption(func(o *options) {
		o.initialConfig.LimiterIn = &control.LimiterConfig{
			Limit: l,
		}
	})
}

// WithDeterministicSamplingIn defines a deterministic sampling strategy which will be applied when a sample is received and before processing it in any way
// (e.g. before determining if a sample belongs to a stream which would require parsing it and evaluating the stream rules).
// Sampling is performed after the input limiter has been applied.
func WithDeterministicSamplingIn(samplingRate int32) Option {
	return newFuncOption(func(o *options) {
		o.samplingIn = control.SamplingConfig{
			SamplingType: control.DeterministicSamplingType,
			DeterministicSampling: control.DeterministicSamplingConfig{
				SampleRate: samplingRate,
			},
		}
	})
}

// WithLimiterOutLimit establishes the initial limiter out rate limit
// in samples per second
//
// Deprecated: Use WithInitialLimiterOutLimit instead
func WithLimiterOutLimit(l int32) Option {
	return WithInitialLimiterOutLimit(l)
}

// WithInitialLimiterInLimit sets the initial limiter in rate limit. This configuration
// is only used the first time a sampler is registered with a server, posterior executions
// will use the configuration stored in the server and the provided configuration will be
// ignored.
func WithInitialLimiterOutLimit(l int32) Option {
	return newFuncOption(func(o *options) {
		o.initialConfig.LimiterOut = &control.LimiterConfig{
			Limit: l,
		}
	})
}

// WithoutDefaultInitialConfig avoids setting the default 'all' stream and digest. This configuration
// is only used the first time a sampler is registered with a server, posterior executions
// will use the configuration stored in the server and the provided configuration will be
// ignored.
func WithoutDefaultInitialConfig() Option {
	return newFuncOption(func(o *options) {
		o.initialConfig.StreamUpdates = []control.StreamUpdate{}
		o.initialConfig.DigestUpdates = []control.DigestUpdate{}
	})
}

// WithUpdateStatsPeriod specifies the period to send sampler stats to server
// If the provided period is less than a second, it will be set to 1 second.
func WithUpdateStatsPeriod(p time.Duration) Option {
	return newFuncOption(func(o *options) {
		if p < time.Second {
			p = time.Second
		}
		o.updateStatsPeriod = p
	})
}

// WithLogger provides a logger instance to log the Provider and Sampler
// activity
func WithLogger(l logging.Logger) Option {
	return newFuncOption(func(o *options) {
		o.logger = l
	})
}

// WithErrorChannel received a channel where Sampler errors will be sent.
// The avoid blocking the Sampler, it won't block if the channel is full so it is responsibility of
// the provider to ensure the channel has enough buffer to avoid losing errors.
func WithSamplerErrorChannel(errCh chan error) Option {
	return newFuncOption(func(o *options) {
		o.samplersErr = errCh
	})
}
