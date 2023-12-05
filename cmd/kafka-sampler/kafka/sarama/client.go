package sarama

import (
	"fmt"

	"github.com/Shopify/sarama"
)

type Client struct {
	config *Config
	sarama.Client
}

func NewClient(servers []string, config *Config) (*Client, error) {
	c, err := sarama.NewClient(servers, config)
	if err != nil {
		return nil, fmt.Errorf("error creating sarama kafka client: %w", err)
	}

	return &Client{
		config: config,
		Client: c,
	}, nil
}
