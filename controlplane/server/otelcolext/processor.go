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
	"github.com/neblic/platform/dataplane/event"
	"github.com/neblic/platform/dataplane/sample"
	"github.com/neblic/platform/internal/pkg/rule"
	"github.com/neblic/platform/logging"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.uber.org/zap"
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
	eventor  *event.Eventor
	digester *digest.Digester
}

type neblic struct {
	cfg      *Config
	exporter *Exporter

	logger       *zap.Logger
	notifyErr    func(err error)
	s            *server.Server
	serverOpts   []server.Option
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

	return &neblic{
		cfg:          cfg,
		logger:       zapLogger,
		notifyErr:    func(err error) { zapLogger.Error("error digesting sample", zap.Error(err)) },
		exporter:     NewExporter(nextConsumer),
		serverOpts:   serverOpts,
		transformers: map[samplerIdentifier]*transformer{},
	}, nil
}

func (n *neblic) newDigester(resource string, sampler string, capabilities *control.CapabilitiesConfig) *digest.Digester {
	return digest.NewDigester(digest.Settings{
		ResourceName:   resource,
		SamplerName:    sampler,
		EnabledDigests: capabilities.CapableDigesters(),
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
			if transformerInstance.eventor == nil {
				settings := event.Settings{
					ResourceName: resource,
					SamplerName:  sampler,
				}
				eventor, err := event.NewEventor(settings)
				if err != nil {
					n.logger.Error("Could not create eventor", zap.Error(err))
					continue
				}
				transformerInstance.eventor = eventor
			}
			transformerInstance.eventor.SetEventsConfig(config.Events)
		} else {
			transformerInstance.eventor = nil
		}

		// Update digester
		if config != nil && len(config.Digests) > 0 {
			if transformerInstance.digester != nil {
				transformerInstance.digester.DeleteDigestsConfig()
				transformerInstance.digester.Close()
			}
			transformerInstance.digester = n.newDigester(resource, sampler, config.Capabilities)
			transformerInstance.digester.SetDigestsConfig(config.Digests)
		} else {
			if transformerInstance.digester != nil {
				transformerInstance.digester.DeleteDigestsConfig()
				transformerInstance.digester.Close()
			}
			transformerInstance.digester = nil
		}

		// Update transformers
		if transformerInstance.digester != nil || transformerInstance.eventor != nil {
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
	var errs error

	sample.RangeSamplers(otlpLogs, func(resource, sampler string, samplerLogs sample.SamplerOTLPLogs) {
		transformer, ok := n.getRuntime(resource, sampler)
		if !ok {
			n.logger.Error("Transformer not found. That should not happen. Skipping event evaluation",
				zap.String("resource", resource),
				zap.String("name", sampler),
			)
			return
		}
		if transformer.eventor == nil {
			// No event rules to evaluate. Nothing to do
			return
		}

		err := transformer.eventor.ProcessSample(samplerLogs)
		if err != nil {
			errs = errors.Join(errs, fmt.Errorf("resource %s, sampler%s -> %w", resource, sampler, err))
		}
	})

	return events, errs
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

	// Delete raw samples from exported data
	otlpLogs.RemoveOTLPLogIf(func(otlpLog any) bool {
		if _, ok := otlpLog.(sample.RawSampleOTLPLog); ok {
			return true
		}
		return false
	})

	return n.exporter.Export(ctx, otlpLogs)
}
