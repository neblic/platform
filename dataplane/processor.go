package dataplane

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/neblic/platform/controlplane/control"
	controlEvent "github.com/neblic/platform/controlplane/event"
	"github.com/neblic/platform/controlplane/server"
	"github.com/neblic/platform/dataplane/digest"
	"github.com/neblic/platform/dataplane/event"
	"github.com/neblic/platform/dataplane/sample"
	"github.com/neblic/platform/logging"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type samplerIdentifier struct {
	resource string
	name     string
}

type transformer struct {
	eventor  *event.Eventor
	digester *digest.Digester

	logger logging.Logger
}

type Processor struct {
	ctx          context.Context
	ctxCancel    context.CancelFunc
	logger       logging.Logger
	exporter     Exporter
	cpServer     *server.Server
	transformers map[samplerIdentifier]*transformer
}

func NewProcessor(logger logging.Logger, cpServer *server.Server, exporter Exporter) *Processor {
	return &Processor{
		ctx:          nil,
		logger:       logger,
		exporter:     exporter,
		cpServer:     cpServer,
		transformers: make(map[samplerIdentifier]*transformer),
	}
}

func (p *Processor) configUpdater() {
	p.logger.Debug("Starting configuration updater routine")

	for serverEvent := range p.cpServer.Events() {
		var logger logging.Logger

		// Get config from event
		var resource, sampler string
		var config *control.SamplerConfig
		switch v := serverEvent.(type) {
		case controlEvent.ConfigUpdate:
			resource = v.Resource
			sampler = v.Sampler
			config = &v.Config
		case controlEvent.ConfigDelete:
			resource = v.Resource
			sampler = v.Sampler
		default:
			continue
		}
		logger = p.logger.With("resource", resource, "sampler", sampler)
		logger.Debug("Received new event", "event", serverEvent)

		p.UpdateConfig(resource, sampler, config, logger)
	}
}

func (p *Processor) getTransformer(resource string, sampler string) (*transformer, bool) {
	samplerIdentifier := samplerIdentifier{
		resource: resource,
		name:     sampler,
	}
	transformer, ok := p.transformers[samplerIdentifier]

	return transformer, ok
}

// computeDigests asynchronous computes the digests for the input samples, generated digests are exported
// in the background
func (p *Processor) computeDigests(otlpLogs sample.OTLPLogs) {
	sample.RangeWithType[sample.RawSampleOTLPLog](otlpLogs, func(resource, sampler string, log sample.RawSampleOTLPLog) {
		transformer, ok := p.getTransformer(resource, sampler)
		if !ok || transformer.digester == nil {
			return
		}

		data, err := log.SampleData()
		if err != nil {
			// we use the transformer logger here because it is rate limited and has context keys
			transformer.logger.Error("Error getting sample data", "error", err)
			return
		}

		transformer.digester.ProcessSample(log.Streams(), data)
	})
}

// computeEvents appends generated events to the provided otlpLogs structure
func (p *Processor) computeEvents(otlpLogs sample.OTLPLogs) {
	// TODO: append to the events information about the sampler that matched it
	sample.RangeSamplers(otlpLogs, func(resource, sampler string, samplerLogs sample.SamplerOTLPLogs) {
		transformer, ok := p.getTransformer(resource, sampler)
		if !ok || transformer.eventor == nil {
			return
		}

		err := transformer.eventor.ProcessSample(samplerLogs)
		if err != nil {
			// we use the transformer logger here because it is rate limited and has context keys
			transformer.logger.Error("Error processing sample", "error", err)
		}
	})
}

func (p *Processor) Start() error {
	if p.ctx != nil {
		return errors.New("processor already started")
	}

	p.ctx, p.ctxCancel = context.WithCancel(context.Background())
	go p.configUpdater()

	return nil
}

func (p *Processor) Stop() error {
	if p.ctx != nil {
		p.ctxCancel()
	}

	for _, transformer := range p.transformers {
		if transformer.digester != nil {
			transformer.digester.Close()
		}
		if transformer.eventor != nil {
			transformer.eventor.Close()
		}
	}

	return nil
}

func (p *Processor) newEventor(resource, sampler string) (*event.Eventor, error) {
	settings := event.Settings{
		ResourceName: resource,
		SamplerName:  sampler,
	}
	return event.NewEventor(settings)
}

func (p *Processor) newDigester(resource, sampler string, digestTypes []control.DigestType, logger logging.Logger) *digest.Digester {
	// the digester performs async operations, so we need to make sure that errors are logged
	notifyErr := func(err error) { logger.Error("error digesting sample", "error", err) }

	return digest.NewDigester(digest.Settings{
		ResourceName:   resource,
		SamplerName:    sampler,
		EnabledDigests: digestTypes,
		NotifyErr:      notifyErr,
		Exporter:       p.exporter,
	})
}

func (p *Processor) UpdateConfig(resource, sampler string, config *control.SamplerConfig, logger logging.Logger) {
	// Forward configuration through data plane
	// In case of having deleted the configuration, send an empty byte array
	var configData []byte
	var err error
	if config != nil {
		configData, err = json.Marshal(config)
	}
	if err != nil {
		logger.Error("Could not marshal config to JSON. Config will not be forwarded downstream", "error", err)
	} else {
		otlpLogs := sample.NewOTLPLogs()
		samplerOtlpLogs := otlpLogs.AppendSamplerOTLPLogs(resource, sampler)
		configOtlpLog := samplerOtlpLogs.AppendConfigOTLPLog()
		configOtlpLog.SetTimestamp(time.Now())
		configOtlpLog.SetSampleRawData(sample.JSONEncoding, configData)

		err := p.exporter.Export(context.Background(), otlpLogs)
		if err != nil {
			logger.Error("Could not export configuration", "error", err)
		} else {
			logger.Debug("Configuration forwarded downstream", "config", config)
		}
	}

	// Get transformer
	samplerIdentifier := samplerIdentifier{
		resource: resource,
		name:     sampler,
	}

	tr, ok := p.transformers[samplerIdentifier]
	if !ok {
		tr = &transformer{
			// each transformer has its own rate limited logger to avoid spamming the logs
			logger: logging.FromZapLogger(
				zap.New(
					zapcore.NewSamplerWithOptions(
						logger.ZapLogger().Core(),
						time.Minute, 1, 0,
					),
				),
			),
		}
	}

	// Update eventor
	if config != nil && len(config.Events) > 0 {
		var eventor *event.Eventor

		if tr.eventor == nil {
			logger.Debug("Setting event configuration", "config", config.Events)
			eventor, err = p.newEventor(resource, sampler)
			if err != nil {
				logger.Error("Could not create eventor", "error", err)
			}
		}
		if eventor != nil {
			err := eventor.SetEventsConfig(config.Events)
			if err != nil {
				logger.Error("Could not set event config", "error", err)
			} else {
				tr.eventor = eventor
			}
		}
	} else {
		logger.Debug("No events configuration found. Disabling eventor")
		tr.eventor = nil
	}

	// Update digester
	if config != nil && len(config.Digests) > 0 {
		logger.Debug("Setting digest configuration", "config", config.Digests)

		if tr.digester == nil {
			tr.digester = p.newDigester(
				resource,
				sampler,
				config.DigestTypesByLocation(control.ComputationLocationCollector),
				tr.logger,
			)
		}
		tr.digester.SetDigestsConfig(config.Digests)
	} else {
		logger.Debug("No digests configuration found. Disabling digester")

		if tr.digester != nil {
			if err := tr.digester.Close(); err != nil {
				logger.Error("Error closing digester", "error", err)
			}
		}
		tr.digester = nil
	}

	// Update transformers
	if tr.digester != nil || tr.eventor != nil {
		p.transformers[samplerIdentifier] = tr
	} else {
		delete(p.transformers, samplerIdentifier)
	}
}

func (p *Processor) Process(ctx context.Context, logs sample.OTLPLogs) error {
	p.computeDigests(logs)
	p.computeEvents(logs)

	// Delete raw samples from exported data
	// TODO: allow the user to configure where the raw samples should be forwarded to
	logs.RemoveOTLPLogIf(func(otlpLog any) bool {
		if _, ok := otlpLog.(sample.RawSampleOTLPLog); ok {
			return true
		}
		return false
	})

	return p.exporter.Export(ctx, logs)
}
