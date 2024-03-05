package dataplane

import (
	"context"
	"encoding/json"
	"errors"
	"sync/atomic"
	"time"

	"github.com/neblic/platform/controlplane/control"
	controlEvent "github.com/neblic/platform/controlplane/event"
	"github.com/neblic/platform/controlplane/server"
	"github.com/neblic/platform/dataplane/digest"
	"github.com/neblic/platform/dataplane/event"
	"github.com/neblic/platform/dataplane/sample"
	"github.com/neblic/platform/internal/pkg/exporter"
	"github.com/neblic/platform/logging"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type samplerIdentifier struct {
	resource string
	name     string
}

type handler struct {
	samplesCollected atomic.Uint64
	streamUIDs       map[string]control.SamplerStreamUID
	eventor          *event.Eventor
	digester         *digest.Digester

	logger logging.Logger
}

func newHandler(logger logging.Logger, resource string, sampler string) *handler {
	return &handler{
		samplesCollected: atomic.Uint64{},
		// each handler has its own rate limited logger to avoid spamming the logs
		logger: logging.FromZapLogger(
			zap.New(
				zapcore.NewSamplerWithOptions(
					logger.With("resource", resource, "sampler", sampler).ZapLogger().Core(),
					time.Minute, 1, 0,
				),
			),
		),
	}
}

type Processor struct {
	ctx       context.Context
	ctxCancel context.CancelFunc
	logger    logging.Logger
	exporter  exporter.LogsExporter
	cpServer  *server.Server
	handlers  map[samplerIdentifier]*handler
}

func NewProcessor(logger logging.Logger, cpServer *server.Server, exporter exporter.LogsExporter) *Processor {
	return &Processor{
		ctx:      nil,
		logger:   logger,
		exporter: exporter,
		cpServer: cpServer,
		handlers: make(map[samplerIdentifier]*handler),
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

func (p *Processor) statsUpdater() {
	p.logger.Debug("Starting stats updater routine")

	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-p.ctx.Done():
			return
		case <-ticker.C:
			for samplerIdentifier, handler := range p.handlers {
				samplesCollected := handler.samplesCollected.Swap(0)
				if samplesCollected == 0 {
					continue
				}

				err := p.cpServer.UpdateSamplerStats(samplerIdentifier.resource, samplerIdentifier.name, samplesCollected)
				if err != nil {
					p.logger.Error("Error updating sampler stats", "error", err)
				}
			}
		}
	}
}

func (p *Processor) getHandler(resource string, sampler string) (*handler, bool) {
	samplerIdentifier := samplerIdentifier{
		resource: resource,
		name:     sampler,
	}
	handler, ok := p.handlers[samplerIdentifier]

	return handler, ok
}

func (p *Processor) setHandler(resource string, sampler string, handler *handler) {
	samplerIdentifier := samplerIdentifier{
		resource: resource,
		name:     sampler,
	}
	p.handlers[samplerIdentifier] = handler
}

func (p *Processor) updateStats(otlpLogs sample.OTLPLogs) {
	sample.Range(otlpLogs, func(resource, sampler string, _ sample.OTLPLog) {
		handler, ok := p.getHandler(resource, sampler)
		if !ok {
			// create handler
			handler = newHandler(p.logger, resource, sampler)
			p.setHandler(resource, sampler, handler)
		}
		handler.samplesCollected.Add(1)
	})
}

// computeDigests asynchronous computes the digests for the input samples, generated digests are exported
// in the background
func (p *Processor) computeDigests(otlpLogs sample.OTLPLogs) {
	sample.RangeWithType[sample.RawSampleOTLPLog](otlpLogs, func(resource, sampler string, log sample.RawSampleOTLPLog) {
		handler, ok := p.getHandler(resource, sampler)
		if !ok || handler.digester == nil {
			return
		}

		data, err := log.SampleData()
		if err != nil {
			// we use the handler logger here because it is rate limited and has context keys
			handler.logger.Error("Error getting sample data", "error", err)
			return
		}

		handler.digester.ProcessSample(log.StreamUIDs(), data)
	})
}

// computeEvents appends generated events to the provided otlpLogs structure
func (p *Processor) computeEvents(otlpLogs sample.OTLPLogs) {
	// TODO: append to the events information about the sampler that matched it
	sample.RangeSamplers(otlpLogs, func(resource, sampler string, samplerLogs sample.SamplerOTLPLogs) {
		handler, ok := p.getHandler(resource, sampler)
		if !ok || handler.eventor == nil {
			return
		}

		err := handler.eventor.ProcessSample(samplerLogs)
		if err != nil {
			// we use the handler logger here because it is rate limited and has context keys
			handler.logger.Error("Error processing sample", "error", err)
		}
	})
}

func (p *Processor) Start() error {
	if p.ctx != nil {
		return errors.New("processor already started")
	}

	p.ctx, p.ctxCancel = context.WithCancel(context.Background())
	go p.configUpdater()
	go p.statsUpdater()

	return nil
}

func (p *Processor) Stop() error {
	if p.ctx != nil {
		p.ctxCancel()
	}

	for _, handler := range p.handlers {
		if handler.digester != nil {
			handler.digester.Close()
		}
		if handler.eventor != nil {
			handler.eventor.Close()
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

func (p *Processor) newDigester(resource, sampler string, logger logging.Logger) *digest.Digester {
	// the digester performs async operations, so we need to make sure that errors are logged
	notifyErr := func(err error) { logger.Error("error digesting sample", "error", err) }

	return digest.NewDigester(
		digest.Settings{
			ResourceName:        resource,
			SamplerName:         sampler,
			ComputationLocation: control.ComputationLocationCollector,
			NotifyErr:           notifyErr,
			Exporter:            p.exporter,
			Logger:              logger,
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

	// Get handler
	samplerIdentifier := samplerIdentifier{
		resource: resource,
		name:     sampler,
	}

	tr, ok := p.handlers[samplerIdentifier]
	if !ok {
		tr = newHandler(p.logger, resource, sampler)
	}

	// Update stream uids map
	if config != nil && len(config.Streams) > 0 {
		tr.streamUIDs = make(map[string]control.SamplerStreamUID, len(config.Streams))
		for _, stream := range config.Streams {
			tr.streamUIDs[stream.Name] = stream.UID
		}
	} else {
		logger.Debug("No streams configuration found")
		tr.streamUIDs = map[string]control.SamplerStreamUID{}
	}

	// Update eventor
	if config != nil && len(config.Events) > 0 {
		logger.Debug("Setting event configuration", "config", config.Events)

		var err error
		if tr.eventor == nil {
			tr.eventor, err = p.newEventor(resource, sampler)
			if err != nil {
				logger.Error("Could not create eventor", "error", err)
				tr.eventor = nil
			}
		}
		if tr.eventor != nil {
			err := tr.eventor.SetEventsConfig(config.Events, config.Streams)
			if err != nil {
				logger.Error("Could not set event config", "error", err)
				tr.eventor = nil
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

	// Update handlers
	p.handlers[samplerIdentifier] = tr
}

func (p *Processor) Process(ctx context.Context, logs sample.OTLPLogs) error {
	// Translate sample names to uids
	sample.Range(logs, func(resource, sampler string, otlpLog sample.OTLPLog) {
		// If there is no configured handler, do nothing
		handler, ok := p.getHandler(resource, sampler)
		if !ok {
			return
		}

		// If stream names is empty or stream uids is already populated, do nothing.
		streamNames := otlpLog.StreamNames()
		if len(streamNames) == 0 || len(otlpLog.StreamUIDs()) > 0 {
			return
		}

		// Translate stream names to stream uids and clear stream names.
		streamUIDs := make([]control.SamplerStreamUID, 0, len(streamNames))
		for _, streamName := range streamNames {
			streamUID, ok := handler.streamUIDs[streamName]
			if !ok {
				handler.logger.Error("Stream name not found when translating from stream name to stream uid", "stream", streamName)
				continue
			}
			streamUIDs = append(streamUIDs, streamUID)
		}
		otlpLog.SetStreamNames([]string{})
		otlpLog.SetStreamUIDs(streamUIDs)

	})

	p.updateStats(logs)
	p.computeDigests(logs)
	p.computeEvents(logs)

	return p.exporter.Export(ctx, logs)
}
