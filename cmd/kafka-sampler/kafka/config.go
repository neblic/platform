package kafka

import (
	"time"

	"github.com/neblic/platform/cmd/kafka-sampler/filter"
	"github.com/neblic/platform/cmd/kafka-sampler/kafka/sarama"
)

type TopicsConfig struct {
	Max           int
	RefreshPeriod time.Duration
	Filter        filter.Config
}

type Config struct {
	Servers       []string
	ConsumerGroup string
	Sarama        sarama.Config
	Topics        TopicsConfig
}

func NewConfig() *Config {
	return &Config{
		Servers:       []string{"localhost:9092"},
		ConsumerGroup: "neblic-kafka-sampler",
		Sarama:        *sarama.NewConfig(),
		Topics: TopicsConfig{
			Max:           25,
			RefreshPeriod: time.Minute,
			Filter:        *filter.NewConfig(),
		},
	}
}

func (c *Config) Finalize() error {
	return sarama.Finalize(&c.Sarama)
}
