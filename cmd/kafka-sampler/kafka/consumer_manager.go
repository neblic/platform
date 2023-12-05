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
	"github.com/neblic/platform/logging"
)

var internalKafkaTopics = map[string]struct{}{"__consumer_offsets": {}, "__transaction_state": {}}

type groupProvider func(topic string) (ConsumerGroup, error)

type consumerInstance struct {
	wg     sync.WaitGroup
	ctx    context.Context
	cancel context.CancelFunc
	group  ConsumerGroup
	topic  string
}

type ConsumerManager struct {
	ctx           context.Context
	logger        logging.Logger
	config        *Config
	client        Client
	topicFilter   *filter.Filter
	groupProvider groupProvider
	consumers     map[string]*consumerInstance
}

func NewConsumerManager(ctx context.Context, logger logging.Logger, config *Config) (*ConsumerManager, error) {
	client, err := sarama.NewClient(config.Servers, &config.Sarama)
	if err != nil {
		return nil, err
	}

	filter, err := filter.New(&config.TopicFilter)
	if err != nil {
		return nil, err
	}

	return &ConsumerManager{
		ctx:         ctx,
		logger:      logger,
		config:      config,
		client:      client,
		topicFilter: filter,
		groupProvider: func(topic string) (ConsumerGroup, error) {
			h := md5.New()
			io.WriteString(h, topic)
			consumerGroup := fmt.Sprintf("%s-%s", config.ConsumerGroup, hex.EncodeToString(h.Sum(nil)[:8]))

			logger.Debug("Creating new consumer group", "topic", topic, "consumer_group", consumerGroup)
			return sarama.NewConsumerGroup(logger, config.Servers, consumerGroup, &config.Sarama)
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
			m.logger.Info("Starting group consume")
			err := g.group.Consume(g.ctx, []string{g.topic})
			if err != nil {
				m.logger.Error(fmt.Sprintf("Error consuming from %s: %v", g.topic, err))
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

	var errors error
	// Add new topics
	for topic := range topicsMap {
		if _, ok := m.consumers[topic]; !ok {
			m.logger.Info(fmt.Sprintf("Starting consumer for topic '%s'", topic))
			// Create consumer group
			// Each topic has its own consumer group so they are independently consumed
			// and all data ends up in the same sampler
			group, err := m.groupProvider(topic)
			if err != nil {
				errors = multierror.Append(errors, fmt.Errorf("cannot create '%s' topic consumer group: %w", topic, err))
				continue
			}

			// Initalize consumer
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

	// Remove old topics
	for topic, consumer := range m.consumers {

		// If the topic no longer exists, stop the consumer group
		if _, ok := topicsMap[topic]; !ok {
			m.logger.Info(fmt.Sprintf("Stopping consumer for topic '%s'", topic))

			// Cancel the consumer and wait until it finishes
			consumer.cancel()
			consumer.wg.Wait()

			consumer.group.Close()
			delete(m.consumers, topic)
		}
	}

	return errors
}

func (m *ConsumerManager) Reconcile() error {

	// Pull last version of the topics
	topics, err := m.pullTopics()
	if err != nil {
		return err
	}

	// Remove internal kafka topics
	externalTopics := make([]string, 0, len(topics))
	for _, topic := range topics {
		if _, ok := internalKafkaTopics[topic]; !ok {
			externalTopics = append(externalTopics, topic)
		}
	}

	// Filter topics based on the user provided rules.
	filteredTopics := m.topicFilter.EvaluateList(externalTopics)

	// Update internal status to reflect the list of pulled topics
	err = m.reconcile(filteredTopics)

	return err
}

func (m *ConsumerManager) Topics() []string {
	topics := make([]string, 0, len(m.consumers))
	for topic := range m.consumers {
		topics = append(topics, topic)
	}
	return topics
}
