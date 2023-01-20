package client

import (
	"time"

	"github.com/neblic/platform/controlplane/internal/stream"
	"github.com/neblic/platform/logging"
)

type options struct {
	streamOpts *stream.Options
	logger     logging.Logger
}

func newDefaultOptions() *options {
	return &options{
		streamOpts: stream.NewOptionsDefault(),
		logger:     logging.NewNopLogger(),
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

func WithBlock() Option {
	return newFuncOption(func(po *options) {
		po.streamOpts.Block = true
	})
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
