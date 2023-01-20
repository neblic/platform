package otelcolext

import (
	"github.com/google/uuid"
	"go.opentelemetry.io/collector/component"
)

const defaultEndpoint = "localhost:8899"

type BearerConfig struct {
	// Expected bearer token
	Token string `mapstructure:"token"`
}

type AuthConfig struct {
	// Configures server authentication, valid values: "bearer"
	Type string `mapstructure:"type"`

	// Bearer configuration
	BearerConfig *BearerConfig `mapstructure:"bearer"`
}

type TLSConfig struct {
	// Path to the TLS cert to use for TLS required connections. (optional)
	CertFile string `mapstructure:"cert_file"`

	// Path to the TLS key to use for TLS required connections. (optional)
	KeyFile string `mapstructure:"key_file"`
}

type Config struct {
	// Optional.
	UID string `mapstructure:"uid"`

	// Optional
	Endpoint string `mapstructure:"endpoint"`

	// Optional
	StoragePath string `mapstructure:"storage_path"`

	// Configures TLS (optional)
	// Default value is nil, which will disable TLS.
	TLSConfig *TLSConfig `mapstructure:"tls"`

	// Configures authentication (optional)
	// Default value is nil, which will disable authentication.
	AuthConfig *AuthConfig `mapstructure:"auth"`
}

var _ component.Config = (*Config)(nil)

// Validate checks if the extension configuration is valid
func (cfg *Config) Validate() error {
	return nil
}

func newDefaultSettings() *Config {
	return &Config{
		UID:        uuid.NewString(),
		Endpoint:   defaultEndpoint,
		TLSConfig:  nil,
		AuthConfig: nil,
	}
}
