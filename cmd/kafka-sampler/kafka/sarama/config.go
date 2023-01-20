package sarama

import "github.com/Shopify/sarama"

type Config = sarama.Config

func NewConfig() *Config {
	return sarama.NewConfig()
}
