package sampler

import (
	"time"

	"github.com/neblic/platform/controlplane/control"
	"github.com/neblic/platform/controlplane/internal/stream"
	"github.com/neblic/platform/logging"
)

type options struct {
	streamOpts *stream.Options

	updateStatsPeriod time.Duration
	logger            logging.Logger

	initialConfig control.SamplerConfigUpdate
	capabilities  control.Capabilities
	tags          []control.Tag
}

func newDefaultSamplerOptions() *options {
	return &options{
		streamOpts: &stream.Options{
			ConnTimeout:        time.Second * time.Duration(30),
			ResponseTimeout:    time.Second * time.Duration(10),
			KeepAliveMaxPeriod: time.Second * time.Duration(10),
			ServerReqsQueueLen: 10,
		},
		logger:        logging.NewNopLogger(),
		initialConfig: control.NewSamplerConfigUpdate(),
	}
}

type Option interface {
	apply(*options)
}

type funcOption struct {
	f func(*options)
}

func (fpo *funcOption) apply(po *options) {
	fpo.f(po)
}

func newFuncOption(f func(*options)) *funcOption {
	return &funcOption{
		f: f,
	}
}

func WithConnTimeout(t time.Duration) Option {
	return newFuncOption(func(po *options) {
		po.streamOpts.ConnTimeout = t
	})
}

func WithResponseTimeout(t time.Duration) Option {
	return newFuncOption(func(po *options) {
		po.streamOpts.ResponseTimeout = t
	})
}

func WihKeepAliveMaxPeriod(t time.Duration) Option {
	return newFuncOption(func(po *options) {
		po.streamOpts.KeepAliveMaxPeriod = t
	})
}

func WithServerReqsQueueLen(l int) Option {
	return newFuncOption(func(po *options) {
		po.streamOpts.ServerReqsQueueLen = l
	})
}

func WithTLS() Option {
	return newFuncOption(func(po *options) {
		po.streamOpts.TLS.Enable = true
	})
}

func WithTLSCACert(path string) Option {
	return newFuncOption(func(po *options) {
		po.streamOpts.TLS.CACertPath = path
	})
}

func WithAuthBearer(token string) Option {
	return newFuncOption(func(po *options) {
		po.streamOpts.Auth.Type = "bearer"
		po.streamOpts.Auth.Bearer.Token = token
	})
}

func WithLogger(l logging.Logger) Option {
	return newFuncOption(func(po *options) {
		po.logger = l
	})
}

func WithInitialConfig(c control.SamplerConfigUpdate) Option {
	return newFuncOption(func(po *options) {
		po.initialConfig = c
	})
}

func WithCapabilities(c control.Capabilities) Option {
	return newFuncOption(func(po *options) {
		po.capabilities = c
	})
}

func WithTags(tags ...control.Tag) Option {
	return newFuncOption(func(po *options) {
		po.tags = tags
	})
}
