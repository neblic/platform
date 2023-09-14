package otelcolext

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/neblic/platform/controlplane/control"
	"github.com/neblic/platform/controlplane/server"
	"github.com/neblic/platform/controlplane/server/internal/registry"
	"github.com/neblic/platform/dataplane/sample"
	"github.com/neblic/platform/internal/pkg/data"
	"github.com/neblic/platform/internal/pkg/rule"
	"github.com/neblic/platform/logging"
	"github.com/neblic/platform/sampler/defs"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.uber.org/zap"
	"golang.org/x/exp/slices"
)

type neblic struct {
	cfg          *Config
	nextConsumer consumer.Logs

	logger          *zap.Logger
	s               *server.Server
	samplerRegistry *registry.SamplerRegistry
	serverOpts      []server.Option
	ruleBuilder     *rule.Builder
}

func newLogsProcessor(cfg *Config, zapLogger *zap.Logger, nextConsumer consumer.Logs) (*neblic, error) {
	serverOpts := []server.Option{}

	if cfg.UID == "" {
		cfg.UID = uuid.NewString()
	}

	if cfg.Endpoint == "" {
		cfg.Endpoint = defaultEndpoint
	}

	if cfg.StoragePath != "" {
		serverOpts = append(serverOpts, server.WithDiskStorage(cfg.StoragePath))
	}

	if cfg.TLSConfig != nil {
		serverOpts = append(serverOpts, server.WithTLS(cfg.TLSConfig.CertFile, cfg.TLSConfig.KeyFile))
	}

	if cfg.AuthConfig != nil {
		switch cfg.AuthConfig.Type {
		case "bearer":
			if cfg.AuthConfig.BearerConfig == nil {
				return nil, fmt.Errorf("Bearer authentication enabled but token not configured")
			}
			serverOpts = append(serverOpts, server.WithAuthBearer(cfg.AuthConfig.BearerConfig.Token))
		case "":
			// nothing to do
		default:
			return nil, fmt.Errorf("Invalid authentication type %s", cfg.AuthConfig.Type)
		}
	}

	serverOpts = append(serverOpts, server.WithLogger(logging.FromZapLogger(zapLogger)))

	if nextConsumer == nil {
		return nil, component.ErrNilNextConsumer
	}

	builder, err := rule.NewBuilder(defs.NewDynamicSchema())
	if err != nil {
		return nil, fmt.Errorf("could not initialize the rule builder: %w", err)
	}

	return &neblic{
		cfg:          cfg,
		logger:       zapLogger,
		nextConsumer: nextConsumer,
		serverOpts:   serverOpts,
		ruleBuilder:  builder,
	}, nil
}

func (n *neblic) Start(ctx context.Context, host component.Host) error {
	var err error
	n.s, err = server.New(n.cfg.UID, n.serverOpts...)
	if err != nil {
		return err
	}
	n.samplerRegistry = n.s.SamplerRegistry()

	return n.s.Start(n.cfg.Endpoint)
}

func (n *neblic) Shutdown(ctx context.Context) error {
	if n.s != nil {
		return n.s.Stop(time.Second)
	}

	return nil
}

func (n *neblic) Capabilities() consumer.Capabilities {
	return consumer.Capabilities{
		MutatesData: true,
	}
}

func (n *neblic) ConsumeLogs(ctx context.Context, ld plog.Logs) error {
	// Convert complex OTLP logs structure to a simpler representation
	samplerSamples := sample.OTLPLogsToSamples(ld)

	// List of new samples (containing events) to add
	samplerSamplesEvents := []sample.SamplerSamples{}

	// Event rule evaluation.
	// Generated events are stored in an auxiliary struct to avoid iterating an modifying the samplers at the same time
	var errs []error
	for _, samplerSample := range samplerSamples {
		sampler, err := n.samplerRegistry.GetSampler(samplerSample.ResourceName, samplerSample.SamplerName)
		if err != nil {
			n.logger.Error("could not get sampler config", zap.Error(err))
			continue
		}

		// stores all the generated events for the current sampler
		matchedSamples := []sample.Sample{}

		for eventUID, eventConfig := range sampler.Config.Events {

			// Get the event rule
			rule, ok := sampler.EventRules[eventUID]
			if !ok {
				n.logger.Error("Event in the configuration cannot be found in the even rules. That should not happen. Skipping event evaluation",
					zap.String("resource", samplerSample.ResourceName), zap.String("name", samplerSample.SamplerName), zap.String("expression", eventConfig.Rule.Expression))
				continue
			}

			for _, sampleData := range samplerSample.Samples {
				// Check if the sample matches the event
				if eventConfig.SampleType == sampleData.Type && slices.Contains(sampleData.Streams, eventConfig.StreamUID) {

					// The event matches the sample
					var record *data.Data
					switch sampleData.Encoding {
					case sample.JSONEncoding:
						record = data.NewSampleDataFromJSON(string(sampleData.Data))
					}
					if record == nil {
						n.logger.Error("Event could not be evaluated because data encoding is not supported", zap.String("encoding", sampleData.Encoding.String()))
						continue
					}

					// Evaluate rule
					ruleMatches, err := rule.Eval(context.Background(), record)
					if err != nil {
						errs = append(errs, fmt.Errorf("eval(%s) -> %w", eventConfig.Rule.Expression, err))
					}
					if ruleMatches {
						matchedSamples = append(matchedSamples, sample.Sample{
							Ts:       time.Now(),
							Type:     control.EventSampleType,
							Streams:  []control.SamplerStreamUID{eventConfig.StreamUID},
							Encoding: sampleData.Encoding,
							Data:     sampleData.Data,
							Metadata: map[sample.MetadataKey]string{
								sample.EventUID:  string(eventUID),
								sample.EventRule: eventConfig.Rule.Expression,
							},
						})
					}
				}
			}
		}

		// In case of having at least one event match, create a new entry in the data with the events.
		if len(matchedSamples) > 0 {
			samplerSamplesEvents = append(samplerSamplesEvents, sample.SamplerSamples{
				ResourceName: samplerSample.ResourceName,
				SamplerName:  samplerSample.SamplerName,
				Samples:      matchedSamples,
			})
		}
	}

	// Append new samplers (containing the generated events) to the data and convert it into OTLP logs
	samplerSamples = append(samplerSamples, samplerSamplesEvents...)
	ld = sample.SamplesToOTLPLogs(samplerSamples)

	if len(errs) > 0 {
		n.logger.Error("Event evaluation finished with errors", zap.Error(errors.Join(errs...)))
	}

	return n.nextConsumer.ConsumeLogs(ctx, ld)
}
