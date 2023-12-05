package main

import (
	"github.com/neblic/platform/cmd/kafka-sampler/kafka"
	"github.com/neblic/platform/cmd/kafka-sampler/neblic"
)

type LoggingConfig struct {
	Level string
}

type Config struct {
	Logging LoggingConfig
	Kafka   kafka.Config
	Neblic  neblic.Config
}

func NewConfig() *Config {
	return &Config{
		Logging: LoggingConfig{
			Level: "info",
		},
		Kafka:  *kafka.NewConfig(),
		Neblic: *neblic.NewConfig(),
	}
}
