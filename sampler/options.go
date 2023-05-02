package sampler

import (
	"time"

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

	limiterInLimit  int64
	limiterOutLimit int64

	updateStatsPeriod time.Duration

	logger logging.Logger
}

func newDefaultOptions() *options {
	return &options{
		controlServerTLSEnable: false,
		dataServerTLSEnable:    false,

		limiterInLimit:  100,
		limiterOutLimit: 10,

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
		o.limiterInLimit = l
	})
}

// WithLimiterOutLimit establishes the initial limiter out rate limit
// in samples per second
func WithLimiterOutLimit(l int64) Option {
	return newFuncOption(func(o *options) {
		o.limiterOutLimit = l
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
