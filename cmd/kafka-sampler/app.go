package main

import (
	"context"
	"time"

	"github.com/neblic/platform/cmd/kafka-sampler/kafka"
	"github.com/neblic/platform/cmd/kafka-sampler/kafka/sarama"
	"github.com/neblic/platform/logging"
)

type KafkaSampler struct {
	ctx             context.Context
	logger          logging.Logger
	config          *Config
	client          kafka.Client
	consumerManager *kafka.ConsumerManager
}

func NewKafkaSampler(ctx context.Context, logger logging.Logger, config *Config) (*KafkaSampler, error) {
	client, err := sarama.NewClient(config.Kafka.Servers, &config.Kafka.Sarama)
	if err != nil {
		return nil, err
	}
	consumerManager, err := kafka.NewConsumerManager(ctx, logger, &config.Kafka)
	if err != nil {
		return nil, err
	}

	kafkaSampler := &KafkaSampler{
		ctx:             ctx,
		logger:          logger,
		config:          config,
		client:          client,
		consumerManager: consumerManager,
	}

	return kafkaSampler, nil
}

func (r *KafkaSampler) Run() error {

	// Run first reconciliation
	r.logger.Info("Running reconcilitation")
	err := r.consumerManager.Reconcile()
	if err != nil {
		return err
	}

	// In case of having a reconcile period of 0 nanoseconds, disable it
	if r.config.ReconcilePeriod == 0 {
		<-r.ctx.Done()
		return nil
	}

	// Execute periodic reconcilitaions
	ticker := time.NewTicker(r.config.ReconcilePeriod)
	for {
		select {
		case <-r.ctx.Done():
			return nil
		case <-ticker.C:
			r.logger.Info("Running reconcilitation")
			err := r.consumerManager.Reconcile()
			if err != nil {
				r.logger.Error("Error running reconciliation", "error", err)
			}
		}
	}
}
