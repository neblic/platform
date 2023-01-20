package kafka

import "context"

type ConsumerGroup interface {
	Consume(ctx context.Context, topics []string) error
	Close() error
}
