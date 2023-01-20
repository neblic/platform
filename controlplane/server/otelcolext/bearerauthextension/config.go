package bearerauthextension

import "go.opentelemetry.io/collector/component"

type Config struct {
	Token string `mapstructure:"token"`
}

var _ component.Config = (*Config)(nil)

// Validate checks if the extension configuration is valid
func (cfg *Config) Validate() error {
	return nil
}
