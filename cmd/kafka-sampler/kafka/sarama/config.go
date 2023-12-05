package sarama

import "github.com/Shopify/sarama"

type Config = sarama.Config

func NewConfig() *Config {
	c := sarama.NewConfig()

	// Setting default configuration
	c.Metadata.RefreshFrequency = 0
	c.Metadata.Full = false
	c.ClientID = "neblic-kafka-sampler"

	return c
}
