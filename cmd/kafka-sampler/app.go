package main

import (
	"context"
	"fmt"
	"time"

	"github.com/neblic/platform/cmd/kafka-sampler/kafka"
	"github.com/neblic/platform/logging"
)

type KafkaSampler struct {
	ctx             context.Context
	logger          logging.Logger
	config          *Config
	consumerManager *kafka.ConsumerManager
}

func NewKafkaSampler(ctx context.Context, logger logging.Logger, config *Config) (*KafkaSampler, error) {
	consumerManager, err := kafka.NewConsumerManager(ctx, logger, &config.Kafka)
	if err != nil {
		return nil, fmt.Errorf("error creating kafka consumer manager: %w", err)
	}

	kafkaSampler := &KafkaSampler{
		ctx:             ctx,
		logger:          logger,
		config:          config,
		consumerManager: consumerManager,
	}

	return kafkaSampler, nil
}

func (r *KafkaSampler) Run() error {
	// Run first reconciliation
	err := r.consumerManager.Reconcile()
	if err != nil {
		return err
	}

	// In case of having a reconcile period of 0 nanoseconds, disable it
	if r.config.Kafka.TopicFilter.RefreshPeriod == 0 {
		r.logger.Warn("Topic refresh period is 0. Added/deleted topics won't be detected.")

		<-r.ctx.Done()
		return nil
	}

	// Execute periodic reconciliations
	ticker := time.NewTicker(r.config.Kafka.TopicFilter.RefreshPeriod)
	for {
		select {
		case <-r.ctx.Done():
			return nil
		case <-ticker.C:
			err := r.consumerManager.Reconcile()
			if err != nil {
				r.logger.Error("Error running reconciliation, it will continue trying", "error", err)
			}
		}
	}
}
