package sampler

import (
	"context"
	"fmt"
	"time"

	"github.com/neblic/platform/controlplane/control"
	csampler "github.com/neblic/platform/controlplane/sampler"
	"github.com/neblic/platform/dataplane/digest"
	dpsample "github.com/neblic/platform/dataplane/sample"
	"github.com/neblic/platform/internal/pkg/data"
	"github.com/neblic/platform/internal/pkg/exporter"
	"github.com/neblic/platform/internal/pkg/rule"
	"github.com/neblic/platform/logging"
	"github.com/neblic/platform/sampler/internal/sample/sampling"
	"github.com/neblic/platform/sampler/sample"
	"golang.org/x/time/rate"
)

var (
	capabilities = control.Capabilities{
		Stream: control.StreamCapabilities{
			Enabled: true,
		},
		LimiterIn: control.LimiterCapabilities{
			Enabled: true,
		},
		SamplingIn: control.SamplingCapabilities{
			Enabled: true,
			Types: []control.SamplingType{
				control.DeterministicSamplingType,
			},
		},
		LimiterOut: control.LimiterCapabilities{
			Enabled: true,
		},
		Digest: control.DigestCapabilities{
			Enabled: true,
			Types: []control.DigestType{
				control.DigestTypeSt,
				control.DigestTypeValue,
			},
		},
	}
)

type streamConfig struct {
	rule            *rule.Rule
	exportRawSample bool
	maxSampleSize   int32
}

type Sampler struct {
	name          string
	resourceName  string
	samplingStats control.SamplerSamplingStats

	configUpdates uint64
	streams       map[control.SamplerStreamUID]streamConfig
	limiterIn     *rate.Limiter
	samplerIn     sampling.Sampler
	limiterOut    *rate.Limiter

	controlPlaneClient *csampler.Sampler
	digester           *digest.Digester
	exporter           exporter.LogsExporter
	ruleBuilder        *rule.Builder

	forwardError func(error)
	logger       logging.Logger
}

func New(
	settings *Settings,
	logger logging.Logger,
) (*Sampler, error) {
	logger.Debug("Initializing sampler", "settings", settings.String())

	ruleBuilder, err := rule.NewBuilder(settings.Schema, rule.StreamFunctions)
	if err != nil {
		return nil, fmt.Errorf("couldn't build CEL rule builder: %w", err)
	}

	var clientOpts []csampler.Option
	clientOpts = append(clientOpts, csampler.WithLogger(logger))
	clientOpts = append(clientOpts, csampler.WithInitialConfig(settings.InitialConfig))
	clientOpts = append(clientOpts, csampler.WithCapabilities(capabilities))
	clientOpts = append(clientOpts, csampler.WithTags(settings.Tags...))
	if settings.EnableTLS {
		clientOpts = append(clientOpts, csampler.WithTLS())
	}

	// provider already checked that it is a valid type
	switch settings.Auth.Type {
	case "bearer":
		clientOpts = append(clientOpts, csampler.WithAuthBearer(settings.Auth.Bearer.Token))
	}

	// TODO: Once we have refactored the sampler client to reuse the same grpc connection internally,
	// move control plane connection to provider
	controlPlaneClient := csampler.New(settings.Name, settings.Resource, clientOpts...)

	forwardError := func(err error) {
		if settings.ErrFwrder != nil {
			select {
			case settings.ErrFwrder <- err:
			default:
			}
		}
	}

	digesterSettings := digest.Settings{
		ResourceName:        settings.Resource,
		SamplerName:         settings.Name,
		ComputationLocation: control.ComputationLocationSampler,
		NotifyErr:           forwardError,
		Exporter:            settings.LogsExporter,
		Logger:              logger,
	}

	digester := digest.NewDigester(digesterSettings)

	p := &Sampler{
		name:         settings.Name,
		resourceName: settings.Resource,

		configUpdates: 0,
		streams:       make(map[control.SamplerStreamUID]streamConfig),
		limiterIn:     nil,
		samplerIn:     nil,
		limiterOut:    nil,

		controlPlaneClient: controlPlaneClient,
		digester:           digester,
		exporter:           settings.LogsExporter,
		ruleBuilder:        ruleBuilder,

		forwardError: forwardError,
		logger:       logger.With("sampler_name", settings.Name, "sampler_uid", controlPlaneClient.UID()),
	}

	go p.listenControlEvents()

	p.logger.Info("Connecting to control plane", "addr", settings.ControlPlaneAddr)
	err = controlPlaneClient.Connect(settings.ControlPlaneAddr)
	if err != nil {
		return nil, fmt.Errorf("couldn't connect to the control plane: %w", err)
	}

	go p.updateStats(settings.UpdateStatsPeriod)

	return p, nil
}

func (p *Sampler) listenControlEvents() {
	p.logger.Debug("Listening for control plane events")

loop:
	for {
		event, more := <-p.controlPlaneClient.Events()
		if !more {
			p.logger.Info("Closed control plane events channel")
			break loop
		}

		switch ev := event.(type) {
		case csampler.ConfigUpdate:
			p.logger.Debug("Received config update", "config", ev.Config)
			p.updateConfig(ev.Config)
		case csampler.StateUpdate:
			switch ev.State {
			case csampler.Registered:
				p.logger.Info("Sampler registered with server")
			case csampler.Unregistered:
				p.logger.Info("Sampler deregistered from server")
			}
		default:
			p.logger.Warn(fmt.Sprintf("Received unknown event type %T", ev))
		}
	}
}

func (p *Sampler) updateStats(period time.Duration) {
	p.logger.Debug("Starting periodic sampler stats update", "period", period)

	if period == 0 {
		p.logger.Warn("Sampler stats update period is 0, stats won't be sent to server")
		return
	}

	ticker := time.NewTicker(period)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := p.controlPlaneClient.UpdateStats(context.Background(), p.samplingStats); err != nil {
				// TODO: use custom errors
				if p.controlPlaneClient.State() != csampler.Unregistered {
					p.logger.Error(fmt.Sprintf("Error updating stats: %s", err))
				}
			}
		}
	}
}

func (p *Sampler) updateConfig(config control.SamplerConfig) {
	p.configUpdates++

	// configure limiter in
	if config.LimiterIn != nil {
		limit := rate.Limit(config.LimiterIn.Limit)
		if limit == -1 {
			limit = rate.Inf
		}

		p.logger.Debug("Configuring limiter in", "limit", limit)
		p.limiterIn = rate.NewLimiter(limit, int(config.LimiterIn.Limit))
	}

	// configure sampler in
	if config.SamplingIn != nil {
		switch config.SamplingIn.SamplingType {
		case control.DeterministicSamplingType:
			p.logger.Debug("Configuring deterministic sampler in", "sample_rate", config.SamplingIn.DeterministicSampling.SampleRate)
			deterministicSampler, err := sampling.NewDeterministicSampler(
				uint(config.SamplingIn.DeterministicSampling.SampleRate),
				config.SamplingIn.DeterministicSampling.SampleEmptyDeterminant)
			if err != nil {
				p.logger.Error(fmt.Sprintf("couldn't initialize the deterministic sampler: %v", err))
			} else {
				p.samplerIn = deterministicSampler
			}
		}
	}

	// replace all existing streams
	if config.Streams != nil {
		newStreams := make(map[control.SamplerStreamUID]streamConfig)
		for _, stream := range config.Streams {
			p.logger.Debug("Configuring stream", "stream", stream)

			builtRule, err := p.buildSamplingRule(stream.StreamRule, stream)
			if err != nil {
				p.logger.Error(fmt.Sprintf("couldn't build sampling rule %+v: %s", stream, err))
				continue
			}

			newStreams[stream.UID] = streamConfig{
				rule:            builtRule,
				exportRawSample: stream.ExportRawSamples,
				maxSampleSize:   stream.MaxSampleSize,
			}
		}

		p.streams = newStreams
	}

	// configure limiter out
	if config.LimiterOut != nil {
		limit := rate.Limit(config.LimiterOut.Limit)
		if limit == -1 {
			limit = rate.Inf
		}

		p.logger.Debug("Configuring limiter out", "limit", limit)
		p.limiterOut = rate.NewLimiter(limit, int(config.LimiterOut.Limit))
	}

	if config.Digests != nil {
		p.logger.Debug("Configuring digests", "digests", config.Digests)
		p.digester.SetDigestsConfig(config.Digests)
	}
}

func (p *Sampler) buildSamplingRule(streamRule control.Rule, stream control.Stream) (*rule.Rule, error) {
	switch streamRule.Lang {
	case control.SrlCel:
		builtRule, err := p.ruleBuilder.Build(streamRule.Expression, stream.Keyed)
		if err != nil {
			return nil, fmt.Errorf("couldn't build CEL rule %s: %s", streamRule.Expression, err)
		}

		return builtRule, nil
	default:
		return nil, fmt.Errorf("couldn't build rule with unknown type %s", streamRule.Lang.String())
	}
}

func (p *Sampler) buildRawSample(streams []control.SamplerStreamUID, key string, sampleData *data.Data) (dpsample.OTLPLogs, error) {
	dataJSON, err := sampleData.JSON()
	if err != nil {
		return dpsample.OTLPLogs{}, fmt.Errorf("couldn't get sampler body: %w", err)
	}

	otlpLogs := dpsample.NewOTLPLogs()
	samplerOtlpLogs := otlpLogs.AppendSamplerOTLPLogs(p.resourceName, p.name)
	rawSample := samplerOtlpLogs.AppendRawSampleOTLPLog()
	rawSample.SetTimestamp(time.Now())
	rawSample.SetStreamUIDs(streams)
	rawSample.SetSampleKey(key)
	rawSample.SetSampleRawData(dpsample.JSONEncoding, []byte(dataJSON))

	return otlpLogs, nil
}

func (p *Sampler) exportRawSample(ctx context.Context, streams []control.SamplerStreamUID, key string, sampleData *data.Data) error {
	otlpLogs, err := p.buildRawSample(streams, key, sampleData)
	if err != nil {
		return err
	}

	if err := p.exporter.Export(ctx, otlpLogs); err != nil {
		return fmt.Errorf("failure to export samples: %w", err)
	}

	p.samplingStats.SamplesExported++

	return nil
}

func (p *Sampler) sample(ctx context.Context, sampleOpts sample.Options, sampleData *data.Data) (bool, error) {
	p.samplingStats.SamplesEvaluated++

	if p.limiterIn != nil && !p.limiterIn.Allow() {
		return false, nil
	}

	// key acts as the deterministic sampler determinant
	if p.samplerIn != nil && !p.samplerIn.Sample(sampleOpts.Key) {
		return false, nil
	}

	// optimization: if there are no output tokens available, no need to do anything since it won't be sampled
	if p.limiterOut != nil && p.limiterOut.Tokens() == 0 {
		return false, nil
	}

	// assign sample to all matching streams based on their rules
	var streams []control.SamplerStreamUID
	var exportRawSample bool
	for streamUID, stream := range p.streams {
		if int(stream.maxSampleSize) > 0 && sampleOpts.Size > int(stream.maxSampleSize) {
			p.forwardError(fmt.Errorf("sample dropepd due to be over the maximum allowed size %d>%d", sampleOpts.Size, stream.maxSampleSize))
			continue
		}

		if match, err := stream.rule.Eval(ctx, sampleData); err != nil {
			p.forwardError(err)
		} else if match {
			streams = append(streams, streamUID)

			if stream.exportRawSample {
				exportRawSample = true
			}
		}
	}

	if len(streams) > 0 {
		if p.limiterOut != nil && !p.limiterOut.Allow() {
			return false, nil
		}

		// forward sample to digester
		if p.digester.ProcessSample(streams, sampleData) {
			p.samplingStats.SamplesDigested++
		}

		// export raw sample
		if exportRawSample {
			err := p.exportRawSample(ctx, streams, sampleOpts.Key, sampleData)
			if err != nil {
				return false, err
			}
		}

		return true, nil
	}

	return false, nil
}

func (p *Sampler) ConfigUpdates() uint64 {
	return p.configUpdates
}

func (p *Sampler) Sample(ctx context.Context, smpl sample.Sample) bool {
	if p.configUpdates == 0 {
		return false
	}

	var (
		sampleData *data.Data
	)
	switch smpl.Type {
	case sample.JSONSampleType:
		sampleData = data.NewSampleDataFromJSON(smpl.JSON)
	case sample.NativeSampleType:
		sampleData = data.NewSampleDataFromNative(smpl.Native)
	case sample.ProtoSampleType:
		sampleData = data.NewSampleDataFromProto(smpl.Proto)
	default:
		return false
	}

	sampled, err := p.sample(ctx, smpl.Options, sampleData)
	if err != nil {
		p.forwardError(err)
		return false
	}

	return sampled
}

func (p *Sampler) Close() error {
	if err := p.controlPlaneClient.Close(closeTimeout); err != nil {
		return fmt.Errorf("error closing control plane client: %w", err)
	}

	return nil
}
