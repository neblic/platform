package sarama

import (
	"fmt"

	"github.com/Shopify/sarama"
	"github.com/hashicorp/go-multierror"
	"github.com/neblic/platform/logging"
	"github.com/neblic/platform/sampler/defs"
	"github.com/neblic/platform/sampler/global"
)

// Handler represents a Sarama consumer group consumer
type SamplerHandler struct {
	logger   logging.Logger
	samplers map[string]defs.Sampler
}

func NewSamplerHandler(logger logging.Logger) *SamplerHandler {
	return &SamplerHandler{
		logger:   logger,
		samplers: map[string]defs.Sampler{},
	}
}

// Setup is run at the beginning of a new session, before ConsumeClaim
func (h *SamplerHandler) Setup(sess sarama.ConsumerGroupSession) error {
	var errors error

	// Initialize one sampler for each topic
	samplerProvider := global.SamplerProvider()
	for topic := range sess.Claims() {
		sampler, err := samplerProvider.Sampler(topic, defs.NewDynamicSchema())
		if err != nil {
			errors = multierror.Append(errors, fmt.Errorf("cannot initialize sampler for topic %s: %w", topic, err))
		}

		h.samplers[topic] = sampler
		h.logger.Info("Initialized sampler", "topic", topic)
	}

	return errors
}

// Cleanup is run at the end of a session, once all ConsumeClaim goroutines have exited
func (h *SamplerHandler) Cleanup(sarama.ConsumerGroupSession) error {
	// Clean samplers
	for topic, sampler := range h.samplers {
		h.logger.Info("Closing sampler", "topic", topic)

		if err := sampler.Close(); err != nil {
			h.logger.Error("Error closing sampler", "topic", topic, "error", err)
		}
	}

	h.samplers = map[string]defs.Sampler{}

	return nil
}

// ConsumeClaim must start a consumer loop of ConsumerGroupClaim's Messages().
func (h *SamplerHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	// NOTE:
	// Do not move the code below to a goroutine.
	// The `ConsumeClaim` itself is called within a goroutine, see:
	// https://github.com/Shopify/sarama/blob/main/consumer_group.go#L27-L29

	h.logger.Debug("Starting to consume messages", "topic", claim.Topic())
	for {
		select {
		case message := <-claim.Messages():
			sampler, ok := h.samplers[message.Topic]
			if !ok {
				return fmt.Errorf("received a message from an unexpected topic: %s. There isn't an initialized sampler for this topic.", message.Topic)
			}
			sampler.Sample(session.Context(), defs.JSONSample(string(message.Value), string(message.Key)))

			session.MarkMessage(message, "")

		// Should return when `session.Context()` is done.
		// If not, will raise `ErrRebalanceInProgress` or `read tcp <ip>:<port>: i/o timeout` when kafka rebalance. see:
		// https://github.com/Shopify/sarama/issues/1192
		case <-session.Context().Done():
			return nil
		}
	}
}
