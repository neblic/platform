package otelcolext

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/neblic/platform/controlplane/control"
	"github.com/neblic/platform/controlplane/server"
	"github.com/neblic/platform/controlplane/server/internal/registry"
	"github.com/neblic/platform/dataplane/digest"
	"github.com/neblic/platform/dataplane/sample"
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

type eventRule struct {
	rule   rule.Rule
	config control.Event
}

type transformer struct {
	eventRules map[control.SamplerEventUID]*eventRule
	digester   *digest.Digester
}

type neblic struct {
	cfg      *Config
	exporter *Exporter

	logger       *zap.Logger
	notifyErr    func(err error)
	s            *server.Server
	serverOpts   []server.Option
	ruleBuilder  *rule.Builder
	transformers map[samplerIdentifier]*transformer
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
				return nil, fmt.Errorf("bearer authentication enabled but token not configured")
			}
			serverOpts = append(serverOpts, server.WithAuthBearer(cfg.AuthConfig.BearerConfig.Token))
		case "":
			// nothing to do
		default:
			return nil, fmt.Errorf("invalid authentication type %s", cfg.AuthConfig.Type)
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
		notifyErr:    func(err error) { zapLogger.Error("error digesting sample", zap.Error(err)) },
		exporter:     NewExporter(nextConsumer),
		serverOpts:   serverOpts,
		ruleBuilder:  builder,
		transformers: map[samplerIdentifier]*transformer{},
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

func (n *neblic) configUpdater() {
	for serverEvent := range n.s.GetEvents() {

		// Get config from event
		var resource string
		var sampler string
		var config *control.SamplerConfig
		switch v := serverEvent.(type) {
		case registry.ConfigUpdate:
			resource = v.Resource
			sampler = v.Sampler
			config = &v.Config
		case registry.ConfigDelete:
			resource = v.Resource
			sampler = v.Sampler
		default:
			continue
		}

		// Forward configuration
		// In case of having deleted the configuration, send an empty byte array
		var configData []byte
		var err error
		if config != nil {
			configData, err = json.Marshal(config)
		}
		if err != nil {
			n.logger.Error("Could not marshal config to JSON. Config will not be forwarded downstream")
		} else {

			otlpLogs := sample.NewOTLPLogs()
			samplerOtlpLogs := otlpLogs.AppendSamplerOTLPLogs(resource, sampler)
			configOtlpLog := samplerOtlpLogs.AppendConfigOTLPLog()
			configOtlpLog.SetTimestamp(time.Now())
			configOtlpLog.SetSampleRawData(sample.JSONEncoding, configData)

			err := n.exporter.Export(context.Background(), otlpLogs)
			if err != nil {
				n.logger.Error("Could not export configuration", zap.Error(err))
			}
		}

		// Get transformer
		samplerIdentifier := samplerIdentifier{
			resource: resource,
			name:     sampler,
		}
		transformerInstance, ok := n.transformers[samplerIdentifier]
		if !ok {
			transformerInstance = new(transformer)
		}

		// Update event rules
		if config != nil && len(config.Events) > 0 {
			if transformerInstance.eventRules == nil {
				transformerInstance.eventRules = map[control.SamplerEventUID]*eventRule{}
			}
			for eventUID, event := range config.Events {
				rule, err := n.ruleBuilder.Build(event.Rule.Expression)
				if err != nil {
					n.logger.Error("Rule cannot be built. Skipping it", zap.String("resource", resource), zap.String("sampler", sampler), zap.Error(err))
					continue
				}

				transformerInstance.eventRules[eventUID] = &eventRule{
					rule:   *rule,
					config: event,
				}
			}
		} else {
			transformerInstance.eventRules = nil
		}

		// Update digester
		if config != nil && len(config.Digests) > 0 {
			if transformerInstance.digester == nil {
				transformerInstance.digester = n.newDigester(resource, sampler)
			}
			transformerInstance.digester.SetDigestsConfig(config.Digests)
		} else {
			if transformerInstance.digester != nil {
				transformerInstance.digester.DeleteDigestsConfig()
				transformerInstance.digester.Close()
			}
			transformerInstance.digester = nil
		}

		// Update transformers
		if transformerInstance.digester != nil || transformerInstance.eventRules != nil {
			n.transformers[samplerIdentifier] = transformerInstance
		} else {
			delete(n.transformers, samplerIdentifier)
		}
	}
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

	// Run configuration updater goroutine
	go n.configUpdater()

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

func (n *neblic) getRuntime(resource string, sampler string) (*transformer, bool) {
	samplerIdentifier := samplerIdentifier{
		resource: resource,
		name:     sampler,
	}
	transformer, ok := n.transformers[samplerIdentifier]

	return transformer, ok
}

// computeDigests asynchronous computes the digests for the input samples, generated digests are exported
// in the background
func (n *neblic) computeDigests(otlpLogs sample.OTLPLogs) {

	sample.RangeWithType[sample.RawSampleOTLPLog](otlpLogs, func(resource, sample string, log sample.RawSampleOTLPLog) {
		transformer, ok := n.getRuntime(resource, sample)
		if !ok || transformer.digester == nil {
			return
		}

		data, err := log.SampleData()
		if err != nil {
			n.logger.Error(err.Error())
			return
		}

		transformer.digester.ProcessSample(log.Streams(), data)
	})
}

// computeEvents returns a list of generated events based on the provided samples (including digests)
func (n *neblic) computeEvents(otlpLogs sample.OTLPLogs) (sample.OTLPLogs, error) {
	// List of computed events
	events := sample.NewOTLPLogs()
	var errs []error

	sample.RangeSamplers(otlpLogs, func(resource, sampler string, samplerLogs sample.SamplerOTLPLogs) {
		transformer, ok := n.getRuntime(resource, sampler)
		if !ok {
			n.logger.Error("Transformer not found. That should not happen. Skipping event evaluation",
				zap.String("resource", resource),
				zap.String("name", sampler),
			)
			return
		}
		if transformer.eventRules == nil {
			// No event rules to evaluate. Nothing to do
			return
		}

		firstSampleEvent := true
		var samplerEvents sample.SamplerOTLPLogs
		sample.RangeSamplerLogsWithType[sample.RawSampleOTLPLog](samplerLogs, func(rawSample sample.RawSampleOTLPLog) {
			for _, eventRule := range transformer.eventRules {
				if slices.Contains(rawSample.Streams(), eventRule.config.StreamUID) {
					data, err := rawSample.SampleData()
					if err != nil {
						n.logger.Error(err.Error())
						continue
					}

					ruleMatches, err := eventRule.rule.Eval(context.Background(), data)
					if err != nil {
						errs = append(errs, fmt.Errorf("eval(%s) -> %w", eventRule.config.Rule.Expression, err))
					}
					if ruleMatches {

						if firstSampleEvent {
							firstSampleEvent = false
							samplerEvents = events.AppendSamplerOTLPLogs(resource, sampler)
						}

						event := samplerEvents.AppendEventOTLPLog()
						event.SetUID(eventRule.config.UID)
						event.SetTimestamp(time.Now())
						event.SetStreams([]control.SamplerStreamUID{eventRule.config.StreamUID})
						event.SetSampleRawData(rawSample.SampleEncoding(), rawSample.SampleRawData())
						event.SetRuleExpression(eventRule.config.Rule.Expression)
					}
				}
			}
		})
	})

	return events, errors.Join(errs...)
}

func (n *neblic) ConsumeLogs(ctx context.Context, logs plog.Logs) error {
	// Convert complex OTLP logs structure to a simpler representation
	// samplerSamples := sample.OTLPLogsToSamples(ld)
	otlpLogs := sample.OTLPLogsFrom(logs)

	// Compute digests and append them to the sampler samples
	n.computeDigests(otlpLogs)

	// Compute events and append them to the sampler samples
	eventOtlpLogs, err := n.computeEvents(otlpLogs)
	if err != nil {
		n.logger.Error("Event evaluation finished with errors", zap.Error(err))
	}

	// Move generated events to the original OTLP logs
	eventOtlpLogs.MoveAndAppendTo(otlpLogs)

	return n.exporter.Export(ctx, otlpLogs)
}
