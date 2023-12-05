package sarama

import (
	"github.com/Shopify/sarama"
)

type Client struct {
	config *Config
	sarama.Client
}

func NewClient(servers []string, config *Config) (*Client, error) {
	c, err := sarama.NewClient(servers, config)

	return &Client{
		config: config,
		Client: c,
	}, err
}
