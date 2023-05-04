package sampler

import (
	"time"

	"github.com/neblic/platform/controlplane/data"
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

	limiterIn  data.LimiterConfig
	samplingIn data.SamplingConfig
	limiterOut data.LimiterConfig

	updateStatsPeriod time.Duration

	logger      logging.Logger
	samplersErr chan error
}

func newDefaultOptions() *options {
	return &options{
		controlServerTLSEnable: false,
		dataServerTLSEnable:    false,

		limiterIn: data.LimiterConfig{
			Limit: 100,
		},
		samplingIn: data.SamplingConfig{},
		limiterOut: data.LimiterConfig{
			Limit: 10,
		},

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
func WithLimiterInLimit(l int64) Option {
	return newFuncOption(func(o *options) {
		o.limiterIn.Limit = l
	})
}

// WithDeterministicSamplingIn defines a deterministic sampling strategy which will be applied when a sample is received and before processing it in any way
// (e.g. before determining if a sample belongs to a stream which would require parsing it and evaluating the stream rules).
// Sampling is performed after the input limiter has been applied.
func WithDeterministicSamplingIn(samplingRate int32) Option {
	return newFuncOption(func(o *options) {
		o.samplingIn = data.SamplingConfig{
			SamplingType: data.DeterministicSamplingType,
			DeterministicSampling: data.DeterministicSamplingConfig{
				SampleRate: samplingRate,
			},
		}
	})
}

// WithLimiterOutLimit establishes the initial limiter out rate limit
// in samples per second
func WithLimiterOutLimit(l int64) Option {
	return newFuncOption(func(o *options) {
		o.limiterOut.Limit = l
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
