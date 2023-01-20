package server

import (
	"time"

	"github.com/neblic/platform/controlplane/server/internal/protocol/stream"
	"github.com/neblic/platform/controlplane/server/internal/registry"
	"github.com/neblic/platform/logging"
)

type tlsOptions struct {
	// enable enables TLS on the server-side. If enabled, CertPath must be provided.
	// By default, it is disabled.
	enable bool
	// certPath points to the server certificate file in PEM format.
	certPath string
	// certKeyPath points to the certificate key file in PEM format.
	certKeyPath string
}

type authBearerOptions struct {
	// token specifies the expected bearer token
	token string
}

type authOptions struct {
	// authType specifies the authentication type, supported values: 'bearer'
	// If empty, no authentication enabled.
	authType string
	// bearer configures the bearer authentication type..
	bearer authBearerOptions
}

type options struct {
	stream               *stream.Options
	tls                  *tlsOptions
	auth                 *authOptions
	registry             *registry.Options
	reconciliationPeriod time.Duration
	logger               logging.Logger
}

func newDefaultOptions() *options {
	return &options{
		tls: &tlsOptions{
			enable: false,
		},
		auth: &authOptions{
			authType: "",
		},
		stream:               stream.NewOptionsDefault(),
		registry:             registry.NewOptionsDefault(),
		reconciliationPeriod: time.Second * time.Duration(5),
		logger:               logging.NewNopLogger(),
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

// WithTLS enables TLS communication encryption
func WithTLS(certFile, keyFile string) Option {
	return newFuncOption(func(po *options) {
		po.tls.enable = true
		po.tls.certPath = certFile
		po.tls.certKeyPath = keyFile
	})
}

// WithAuthBearer configures the server to validate connections based on the provided token
func WithAuthBearer(token string) Option {
	return newFuncOption(func(po *options) {
		po.auth.authType = "bearer"
		po.auth.bearer.token = token
	})
}

// WithDiskStorage enables the persistence of the internal state of the server
// to disk. Path contains the root folder where the data will be stored
func WithDiskStorage(path string) Option {
	return newFuncOption(func(po *options) {
		po.registry.StorageType = registry.DiskStorage
		po.registry.Path = path
	})
}

// WithReconciliationPeriod sets how often the server will try to reconcile its configuration with the
// connected samplers.
func WithReconciliationPeriod(p time.Duration) Option {
	return newFuncOption(func(po *options) {
		po.reconciliationPeriod = p
	})
}

// WithLogger changes the internal logger. In case of not setting it, the server does not output logs
func WithLogger(l logging.Logger) Option {
	return newFuncOption(func(po *options) {
		po.logger = l
	})
}
