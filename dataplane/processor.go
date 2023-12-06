package dataplane

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/neblic/platform/controlplane/control"
	controlEvent "github.com/neblic/platform/controlplane/event"
	"github.com/neblic/platform/controlplane/server"
	"github.com/neblic/platform/dataplane/digest"
	"github.com/neblic/platform/dataplane/event"
	"github.com/neblic/platform/dataplane/sample"
	"go.uber.org/zap"
)

type samplerIdentifier struct {
	resource string
	name     string
}

type transformer struct {
	eventor  *event.Eventor
	digester *digest.Digester
}

type Processor struct {
	ctx          context.Context
	ctxCancel    context.CancelFunc
	logger       *zap.Logger
	notifyErr    func(err error)
	exporter     Exporter
	controlplane *server.Server
	transformers map[samplerIdentifier]*transformer
}

func NewProcessor(logger *zap.Logger, controlplane *server.Server, exporter Exporter) *Processor {
	return &Processor{
		ctx:          nil,
		logger:       logger,
		notifyErr:    func(err error) { logger.Error("error digesting sample", zap.Error(err)) },
		exporter:     exporter,
		controlplane: controlplane,
		transformers: make(map[samplerIdentifier]*transformer),
	}
}

func (p *Processor) configUpdater() {
	eventsChan := p.controlplane.GetEvents()

	for serverEvent := range eventsChan {
		// Get config from event
		var resource string
		var sampler string
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

		p.UpdateConfig(resource, sampler, config)

	}
}

func (p *Processor) getRuntime(resource string, sampler string) (*transformer, bool) {
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

	sample.RangeWithType[sample.RawSampleOTLPLog](otlpLogs, func(resource, sample string, log sample.RawSampleOTLPLog) {
		transformer, ok := p.getRuntime(resource, sample)
		if !ok || transformer.digester == nil {
			return
		}

		data, err := log.SampleData()
		if err != nil {
			p.logger.Error(err.Error())
			return
		}

		transformer.digester.ProcessSample(log.Streams(), data)
	})
}

// computeEvents returns a list of generated events based on the provided samples (including digests)
func (p *Processor) computeEvents(otlpLogs sample.OTLPLogs) (sample.OTLPLogs, error) {
	// List of computed events
	events := sample.NewOTLPLogs()
	var errs error

	sample.RangeSamplers(otlpLogs, func(resource, sampler string, samplerLogs sample.SamplerOTLPLogs) {
		transformer, ok := p.getRuntime(resource, sampler)
		if !ok {
			p.logger.Error("Transformer not found. That should not happen. Skipping event evaluation",
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

func (p *Processor) UpdateConfig(resource string, sampler string, config *control.SamplerConfig) {
	// Forward configuration
	// In case of having deleted the configuration, send an empty byte array
	var configData []byte
	var err error
	if config != nil {
		configData, err = json.Marshal(config)
	}
	if err != nil {
		p.logger.Error("Could not marshal config to JSON. Config will not be forwarded downstream")
	} else {

		otlpLogs := sample.NewOTLPLogs()
		samplerOtlpLogs := otlpLogs.AppendSamplerOTLPLogs(resource, sampler)
		configOtlpLog := samplerOtlpLogs.AppendConfigOTLPLog()
		configOtlpLog.SetTimestamp(time.Now())
		configOtlpLog.SetSampleRawData(sample.JSONEncoding, configData)

		err := p.exporter.Export(context.Background(), otlpLogs)
		if err != nil {
			p.logger.Error("Could not export configuration", zap.Error(err))
		}
	}

	// Get transformer
	samplerIdentifier := samplerIdentifier{
		resource: resource,
		name:     sampler,
	}
	transformerInstance, ok := p.transformers[samplerIdentifier]
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
				p.logger.Error("Could not create eventor", zap.Error(err))
				return
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
		settings := digest.Settings{
			ResourceName:   resource,
			SamplerName:    sampler,
			EnabledDigests: config.Capabilities.NotCapableDigesters(),
			NotifyErr:      p.notifyErr,
			Exporter:       p.exporter,
		}
		transformerInstance.digester = digest.NewDigester(settings)
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
		p.transformers[samplerIdentifier] = transformerInstance
	} else {
		delete(p.transformers, samplerIdentifier)
	}
}

func (p *Processor) Process(ctx context.Context, logs sample.OTLPLogs) error {

	// Compute digests and append them to the sampler samples
	p.computeDigests(logs)

	// Compute events and append them to the sampler samples
	eventOtlpLogs, err := p.computeEvents(logs)
	if err != nil {
		p.logger.Error("Event evaluation finished with errors", zap.Error(err))
	}

	// Move generated events to the original OTLP logs
	eventOtlpLogs.MoveAndAppendTo(logs)

	// Delete raw samples from exported data
	logs.RemoveOTLPLogIf(func(otlpLog any) bool {
		if _, ok := otlpLog.(sample.RawSampleOTLPLog); ok {
			return true
		}
		return false
	})

	return p.exporter.Export(ctx, logs)
}
