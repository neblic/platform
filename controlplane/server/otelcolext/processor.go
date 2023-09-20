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
	"github.com/neblic/platform/dataplane/digest"
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

type samplerIdentifier struct {
	resource string
	name     string
}

type runtime struct {
	eventRules map[control.SamplerEventUID]*rule.Rule
	digester   *digest.Digester
}

type neblic struct {
	cfg      *Config
	exporter *Exporter

	logger      *zap.Logger
	notifyErr   func(err error)
	s           *server.Server
	serverOpts  []server.Option
	ruleBuilder *rule.Builder
	runtimes    map[samplerIdentifier]*runtime
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
		cfg:         cfg,
		logger:      zapLogger,
		notifyErr:   func(err error) { zapLogger.Error("error digesting sample", zap.Error(err)) },
		exporter:    NewExporter(nextConsumer),
		serverOpts:  serverOpts,
		ruleBuilder: builder,
		runtimes:    map[samplerIdentifier]*runtime{},
	}, nil
}

func (n *neblic) newDigester(resource string, sampler string) *digest.Digester {
	return digest.NewDigester(digest.Settings{
		ResourceName:   resource,
		SamplerName:    sampler,
		EnabledDigests: []control.DigestType{control.DigestTypeValue},
		NotifyErr:      n.notifyErr,
		Exporter:       n.exporter,
	})
}

func (n *neblic) Start(ctx context.Context, host component.Host) error {
	var err error
	n.s, err = server.New(n.cfg.UID, n.serverOpts...)
	if err != nil {
		return err
	}

	err = n.s.Start(n.cfg.Endpoint)
	if err != nil {
		return err
	}

	// Populate existing config from the controlplane
	n.runtimes = map[samplerIdentifier]*runtime{}
	n.s.RangeSamplersConfig(func(resource, sampler string, config control.SamplerConfig) (carryon bool) {
		// Initialize event rules
		eventRules := map[control.SamplerEventUID]*rule.Rule{}
		for eventUID, event := range config.Events {
			rule, err := n.ruleBuilder.Build(event.Rule.Expression)
			if err != nil {
				n.logger.Error("rule cannot be built. Skipping it", zap.String("resource", resource), zap.String("sampler", sampler), zap.Error(err))
				continue
			}

			eventRules[eventUID] = rule
		}

		// Initialize digest if exists config for it
		var digester *digest.Digester
		if len(config.Digests) > 0 {
			digester = n.newDigester(resource, sampler)
			digester.SetDigestsConfig(config.Digests)
		}

		n.runtimes[samplerIdentifier{resource: resource, name: sampler}] = &runtime{
			eventRules: eventRules,
			digester:   digester,
		}

		return true
	})

	// Run config update goroutine
	go func() {
		eventsChan, err := n.s.GetRegistryEvents()
		if err != nil {
			n.logger.Fatal(err.Error())
		}

		for registryEvent := range eventsChan {
			if registryEvent.RegistryType != registry.SamplerRegistryType {
				continue
			}

			// Get runtime
			resource := registryEvent.SamplerRegistryEvent.Resource
			sampler := registryEvent.SamplerRegistryEvent.Sampler
			config := registryEvent.SamplerRegistryEvent.Config
			samplerIdentifier := samplerIdentifier{
				resource: resource,
				name:     sampler,
			}
			runtimeInstance, ok := n.runtimes[samplerIdentifier]
			if !ok {
				runtimeInstance = new(runtime)
			}

			// Update event rules
			var eventRulesOperation registry.Operation
			if registryEvent.Operation == registry.UpsertOperation && len(config.Events) > 0 {
				eventRulesOperation = registry.UpsertOperation
			} else {
				eventRulesOperation = registry.DeleteOperation
			}
			switch eventRulesOperation {
			case registry.UpsertOperation:
				if runtimeInstance.eventRules == nil {
					runtimeInstance.eventRules = map[control.SamplerEventUID]*rule.Rule{}
				}
				for eventUID, event := range config.Events {
					rule, err := n.ruleBuilder.Build(event.Rule.Expression)
					if err != nil {
						n.logger.Error("rule cannot be built. Skipping it", zap.String("resource", resource), zap.String("sampler", sampler), zap.Error(err))
						continue
					}

					runtimeInstance.eventRules[eventUID] = rule
				}
			case registry.DeleteOperation:
				runtimeInstance.eventRules = nil
			}

			// Update digester
			var digesterOperation registry.Operation
			if registryEvent.Operation == registry.UpsertOperation && len(config.Digests) > 0 {
				digesterOperation = registry.UpsertOperation
			} else {
				digesterOperation = registry.DeleteOperation
			}
			switch digesterOperation {
			case registry.UpsertOperation:
				if runtimeInstance.digester == nil {
					runtimeInstance.digester = n.newDigester(resource, sampler)
				}
				runtimeInstance.digester.SetDigestsConfig(config.Digests)
			case registry.DeleteOperation:
				if runtimeInstance.digester != nil {
					runtimeInstance.digester.DeleteDigestsConfig()
					runtimeInstance.digester.Close()
				}
				runtimeInstance.digester = nil
			}

			// Update runtime
			if runtimeInstance.digester != nil || runtimeInstance.eventRules != nil {
				n.runtimes[samplerIdentifier] = runtimeInstance
			} else {
				delete(n.runtimes, samplerIdentifier)
			}
		}

	}()

	return nil
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

func (n *neblic) getRuntime(resource string, sampler string) (*runtime, bool) {
	samplerIdentifier := samplerIdentifier{
		resource: resource,
		name:     sampler,
	}
	runtime, ok := n.runtimes[samplerIdentifier]

	return runtime, ok
}

// computeDigests asynchronous computes the digests for the input samples, generated digests are exported
// in the background
func (n *neblic) computeDigests(samplerSamples []sample.SamplerSamples) {

	// Generated digests are stored in an auxiliary struct to avoid iterating an modifying the samplers at the same time
	for _, samplerSample := range samplerSamples {
		runtime, ok := n.getRuntime(samplerSample.ResourceName, samplerSample.SamplerName)
		if !ok || runtime.digester == nil {
			continue
		}

		for _, sampleData := range samplerSample.Samples {

			// Check the data encoding is supported
			var record *data.Data
			switch sampleData.Encoding {
			case sample.JSONEncoding:
				record = data.NewSampleDataFromJSON(string(sampleData.Data))
			}
			if record == nil {
				n.logger.Error("Digest could not be performed because data encoding is not supported", zap.String("encoding", sampleData.Encoding.String()))
				continue
			}

			// Digests can only be performed to raw samples
			if sampleData.Type != control.RawSampleType {
				continue
			}

			runtime.digester.ProcessSample(sampleData.Streams, record)
		}
	}
}

// computeEvents returns a list of generated events based on the provided samples (including digests)
func (n *neblic) computeEvents(samplerSamples []sample.SamplerSamples) ([]sample.SamplerSamples, error) {
	// List of computed events
	samplerSamplesEvents := []sample.SamplerSamples{}

	// Event rule evaluation.
	// Generated events are stored in an auxiliary struct to avoid iterating an modifying the samplers at the same time
	var errs []error
	for _, samplerSample := range samplerSamples {
		samplerConfig, err := n.s.GetSamplerConfig(samplerSample.ResourceName, samplerSample.SamplerName)
		if err != nil {
			n.logger.Error("could not get sampler config", zap.Error(err))
			continue
		}

		// stores all the generated events for the current sampler
		matchedSamples := []sample.Sample{}

		for eventUID, eventConfig := range samplerConfig.Events {

			// Get the event rule
			runtime, ok := n.getRuntime(samplerSample.ResourceName, samplerSample.SamplerName)
			if !ok {
				n.logger.Error("runtime not found. That should not happen. Skipping event evaluation",
					zap.String("resource", samplerSample.ResourceName),
					zap.String("name", samplerSample.SamplerName),
				)
				continue
			}
			if runtime.eventRules == nil {
				n.logger.Error("runtime found, but event rules has not been initialized. That should not happen. Skipping event evaluation",
					zap.String("resource", samplerSample.ResourceName),
					zap.String("name", samplerSample.SamplerName),
				)
				continue
			}

			rule, ok := runtime.eventRules[eventUID]
			if !ok {
				n.logger.Error("Event in the configuration cannot be found in the event rules. That should not happen. Skipping event evaluation",
					zap.String("resource", samplerSample.ResourceName),
					zap.String("name", samplerSample.SamplerName),
					zap.String("expression", eventConfig.Rule.Expression),
				)
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

	return samplerSamplesEvents, errors.Join(errs...)
}

func (n *neblic) ConsumeLogs(ctx context.Context, ld plog.Logs) error {
	// Convert complex OTLP logs structure to a simpler representation
	samplerSamples := sample.OTLPLogsToSamples(ld)

	// Compute digests and append them to the sampler samples
	n.computeDigests(samplerSamples)

	// Compute events and append them to the sampler samples
	samplerSamplesEvents, err := n.computeEvents(samplerSamples)
	if err != nil {
		n.logger.Error("Event evaluation finished with errors", zap.Error(err))
	}

	// Append new samplers (containing the generated events) to the data and convert it into OTLP logs
	samplerSamples = append(samplerSamples, samplerSamplesEvents...)

	return n.exporter.Export(ctx, samplerSamples)
}
