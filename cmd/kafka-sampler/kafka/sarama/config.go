package sarama

import (
	"crypto/sha256"
	"crypto/sha512"

	"github.com/IBM/sarama"
)

type Config = sarama.Config

func NewConfig() *Config {
	c := sarama.NewConfig()

	// Setting default configuration
	c.Metadata.RefreshFrequency = 0
	c.Metadata.Full = false
	c.ClientID = "neblic-kafka-sampler"

	return c
}

func Finalize(c *Config) error {
	if c.Net.SASL.Mechanism == sarama.SASLTypeSCRAMSHA256 {
		c.Net.SASL.SCRAMClientGeneratorFunc = func() sarama.SCRAMClient {
			return &XDGSCRAMClient{HashGeneratorFcn: sha256.New}
		}
	}
	if c.Net.SASL.Mechanism == sarama.SASLTypeSCRAMSHA512 {
		c.Net.SASL.SCRAMClientGeneratorFunc = func() sarama.SCRAMClient {
			return &XDGSCRAMClient{HashGeneratorFcn: sha512.New}
		}
	}

	return c.Validate()
}
