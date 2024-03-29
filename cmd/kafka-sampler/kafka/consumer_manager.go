package kafka

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"sync"

	multierror "github.com/hashicorp/go-multierror"
	"github.com/neblic/platform/cmd/kafka-sampler/filter"
	"github.com/neblic/platform/cmd/kafka-sampler/kafka/sarama"
	"github.com/neblic/platform/cmd/kafka-sampler/neblic"
	"github.com/neblic/platform/controlplane/control"
	"github.com/neblic/platform/logging"
	"github.com/neblic/platform/sampler"
)

type groupProvider func(topic string) (ConsumerGroup, error)

type consumerInstance struct {
	wg     sync.WaitGroup
	ctx    context.Context
	cancel context.CancelFunc
	group  ConsumerGroup
	topic  string
}

type ConsumerManager struct {
	ctx            context.Context
	logger         logging.Logger
	config         *Config
	client         Client
	internalFilter *filter.Filter
	topicFilter    *filter.Filter
	groupProvider  groupProvider
	consumers      map[string]*consumerInstance
}

func NewConsumerManager(ctx context.Context, logger logging.Logger, config *Config, neblicConfig *neblic.Config) (*ConsumerManager, error) {
	client, err := sarama.NewClient(config.Servers, &config.Sarama)
	if err != nil {
		return nil, err
	}

	// Create filter including well known internal topics
	// _schemas: Confluent Schema Registry
	// __consumer_offsets: Kafka group offsets
	// __transaction_state: Kafka transaction state
	// _confluent*: Confluent related topics
	// *-KSTREAM-*: Kafka Streams
	// *-changelog: Kafka Streams
	// __amazon_msk*: MSK
	internalFilterRegex, err := filter.NewRegex("(^_schemas$|^__consumer_offsets$|^__transaction_state$|^_confluent.*$|^.*-KSTREAM-.*$|^.*-changelog$|^__amazon_msk.*$)")
	if err != nil {
		return nil, err
	}
	internalFilter, err := filter.New(&filter.Config{
		Deny: internalFilterRegex,
	})
	if err != nil {
		return nil, err
	}

	filter, err := filter.New(&config.Topics.Filter)
	if err != nil {
		return nil, err
	}

	samplerOpts := []sampler.Option{
		sampler.WithInitialStructDigest(control.ComputationLocationSampler),
		sampler.WithInitalValueDigest(control.ComputationLocationSampler),
	}
	if neblicConfig.LimiterOutLimit != 0 {
		samplerOpts = append(samplerOpts, sampler.WithInitialLimiterOutLimit(int32(neblicConfig.LimiterOutLimit)))
	}
	if neblicConfig.UpdateStatsPeriod != 0 {
		samplerOpts = append(samplerOpts, sampler.WithUpdateStatsPeriod(neblicConfig.UpdateStatsPeriod))
	}

	return &ConsumerManager{
		ctx:            ctx,
		logger:         logger,
		config:         config,
		client:         client,
		internalFilter: internalFilter,
		topicFilter:    filter,
		groupProvider: func(topic string) (ConsumerGroup, error) {
			h := md5.New()
			io.WriteString(h, topic)
			consumerGroup := fmt.Sprintf("%s-%s", config.ConsumerGroup, hex.EncodeToString(h.Sum(nil)[:8]))

			logger.Debug("Creating new consumer group", "topic", topic, "consumer_group", consumerGroup)
			return sarama.NewConsumerGroup(logger, config.Servers, consumerGroup, &config.Sarama, samplerOpts)
		},
		consumers: map[string]*consumerInstance{},
	}, nil
}

func (m *ConsumerManager) runConsumerInstance(g *consumerInstance) {
	g.wg.Add(1)
	defer g.wg.Done()

	for {
		select {
		case <-g.ctx.Done():
			return
		default:
			m.logger.Debug("Starting group consume", "topic", g.topic)

			err := g.group.Consume(g.ctx, []string{g.topic})
			if err != nil {
				m.logger.Error("Consume error", "topic", g.topic, "error", err)
			}
		}
	}
}

func (m *ConsumerManager) pullTopics() ([]string, error) {
	err := m.client.RefreshMetadata()
	if err != nil {
		return nil, fmt.Errorf("cannot refresh kafka metadata: %w", err)
	}

	// Create map with all the topics
	topics, err := m.client.Topics()
	if err != nil {
		return topics, fmt.Errorf("cannot refresh kafka metadata: %w", err)
	}

	return topics, nil
}

func (m *ConsumerManager) reconcile(topics []string) error {
	// Create topic map to simplify the logic
	topicsMap := map[string]struct{}{}
	for _, topic := range topics {
		topicsMap[topic] = struct{}{}
	}

	// Remove old topics
	for topic, consumer := range m.consumers {

		// If the topic no longer exists, stop the consumer group
		if _, ok := topicsMap[topic]; !ok {
			m.logger.Info("Stopping consumer", "topic", topic)

			// Cancel the consumer and wait until it finishes
			consumer.cancel()
			consumer.wg.Wait()

			if err := consumer.group.Close(); err != nil {
				m.logger.Error("Error closing consumer group", "topic", topic, "error", err)
			}
			delete(m.consumers, topic)
		}
	}

	// Add new topics
	var errors error
	for topic := range topicsMap {
		if _, ok := m.consumers[topic]; !ok {

			if len(m.consumers) >= m.config.Topics.Max {
				m.logger.Warn("Ignoring topic because the maximum number has been reached", "topic", topic)
				continue
			}

			m.logger.Info("Starting consumer", "topic", topic)

			// Create consumer group
			// Each topic has its own consumer group so they are independently consumed
			// and all data ends up in the same sampler
			group, err := m.groupProvider(topic)
			if err != nil {
				errors = multierror.Append(errors, fmt.Errorf("cannot create '%s' topic consumer group: %w", topic, err))
				continue
			}

			// Initialize consumer
			ctx, cancel := context.WithCancel(m.ctx)
			consumerInstance := &consumerInstance{
				wg:     sync.WaitGroup{},
				ctx:    ctx,
				cancel: cancel,
				group:  group,
				topic:  topic,
			}
			go m.runConsumerInstance(consumerInstance)
			m.consumers[topic] = consumerInstance
		}
	}

	return errors
}

func (m *ConsumerManager) Reconcile() error {
	m.logger.Debug("Running topics reconcilitation")

	// Pull last version of the topics
	topics, err := m.pullTopics()
	if err != nil {
		return err
	}
	m.logger.Debug(fmt.Sprintf("Available topics: %s", topics))

	// Remove internal kafka topics
	internalAllowed, internalDenied := m.internalFilter.EvaluateList(topics)
	m.logger.Debug(fmt.Sprintf("Removed internal kafka topics: %v", internalDenied))

	// Filter topics based on the user provided rules.
	externalAllowed, externalDenied := m.topicFilter.EvaluateList(internalAllowed)
	m.logger.Debug(fmt.Sprintf("Removed user filtered topics: %v", externalDenied))
	m.logger.Debug(fmt.Sprintf("Assigned topics: %v", externalAllowed))

	// Update internal status to reflect the list of pulled topics
	err = m.reconcile(externalAllowed)

	return err
}

func (m *ConsumerManager) Topics() []string {
	topics := make([]string, 0, len(m.consumers))
	for topic := range m.consumers {
		topics = append(topics, topic)
	}
	return topics
}
