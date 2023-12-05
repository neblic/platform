package sarama

import (
	"context"
	"fmt"

	"github.com/Shopify/sarama"
	"github.com/neblic/platform/logging"
)

type ConsumerGroup struct {
	group   sarama.ConsumerGroup
	handler *SamplerHandler
}

func NewConsumerGroup(logger logging.Logger, servers []string, groupID string, config *Config) (*ConsumerGroup, error) {
	group, err := sarama.NewConsumerGroup(servers, groupID, config)
	if err != nil {
		return nil, fmt.Errorf("error creating saram kafka consumer group: %w", err)
	}

	return &ConsumerGroup{
		group:   group,
		handler: NewSamplerHandler(logger),
	}, nil
}

func (c *ConsumerGroup) Consume(ctx context.Context, topics []string) error {
	return c.group.Consume(ctx, topics, c.handler)
}

func (c *ConsumerGroup) Close() error {
	return c.group.Close()
}
