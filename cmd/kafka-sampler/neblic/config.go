package neblic

import (
	"time"

	"github.com/neblic/platform/sampler"
)

type Options struct {
	Bearer            string
	TLS               bool
	SamplerLimit      uint
	UpdateStatsPeriod time.Duration
}

type Config struct {
	sampler.Settings `mapstructure:",squash"`
	Options          `mapstructure:",squash"`
}

func NewConfig() *Config {
	return &Config{
		Settings: sampler.Settings{
			ResourceName:      "kafkasampler",
			ControlServerAddr: "localhost:8899",
			DataServerAddr:    "localhost:4317",
		},
		Options: Options{
			UpdateStatsPeriod: time.Second * 15,
		},
	}
}
