package kafka

import (
	"github.com/neblic/platform/cmd/kafka-sampler/filter"
	"github.com/neblic/platform/cmd/kafka-sampler/kafka/sarama"
)

type Config struct {
	Servers       []string
	ConsumerGroup string
	Sarama        sarama.Config
	TopicFilter   filter.Config
}

func NewConfig() *Config {
	return &Config{
		Servers:       []string{"localhost:9092"},
		ConsumerGroup: "kafkasampler",
		Sarama:        *sarama.NewConfig(),
		TopicFilter:   *filter.NewConfig(),
	}
}
