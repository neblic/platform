package main

import (
	"time"

	"github.com/neblic/platform/cmd/kafka-sampler/kafka"
	"github.com/neblic/platform/cmd/kafka-sampler/neblic"
)

type Config struct {
	Verbose         bool
	Kafka           kafka.Config
	Neblic          neblic.Config
	ReconcilePeriod time.Duration
}

func NewConfig() *Config {
	return &Config{
		Verbose:         false,
		Kafka:           *kafka.NewConfig(),
		Neblic:          *neblic.NewConfig(),
		ReconcilePeriod: time.Minute,
	}
}
