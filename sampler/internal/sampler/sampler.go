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
	"github.com/neblic/platform/sampler/defs"
	"github.com/neblic/platform/sampler/internal/sample/sampling"
	"golang.org/x/time/rate"
)

var _ defs.Sampler = (*Sampler)(nil)

type streamConfig struct {
	rule            *rule.Rule
	exportRawSample bool
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
	exporter           exporter.Exporter
	ruleBuilder        *rule.Builder

	forwardError func(error)
	logger       logging.Logger
}

func New(
	settings *Settings,
	logger logging.Logger,
) (*Sampler, error) {
	ruleBuilder, err := rule.NewBuilder(settings.Schema)
	if err != nil {
		return nil, fmt.Errorf("couldn't build CEL rule builder: %w", err)
	}

	var clientOpts []csampler.Option
	clientOpts = append(clientOpts, csampler.WithLogger(logger))
	clientOpts = append(clientOpts, csampler.WithInitialConfig(settings.InitialConfig))
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

	initialConfig := control.NewSamplerConfig()
	initialConfig.Merge(settings.InitialConfig)

	digesterSettings := digest.Settings{
		ResourceName:   settings.Resource,
		SamplerName:    settings.Name,
		EnabledDigests: initialConfig.DigestTypesByLocation(control.ComputationLocationSampler),
		NotifyErr:      forwardError,
		Exporter:       settings.Exporter,
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
		exporter:           settings.Exporter,
		ruleBuilder:        ruleBuilder,

		forwardError: forwardError,
		logger:       logger.With("sampler_name", settings.Name, "sampler_uid", controlPlaneClient.UID()),
	}

	go p.listenControlEvents()

	err = controlPlaneClient.Connect(settings.ControlPlaneAddr)
	if err != nil {
		return nil, fmt.Errorf("couldn't connect to the control plane: %w", err)
	}

	go p.updateStats(settings.UpdateStatsPeriod)

	return p, nil
}

func (p *Sampler) listenControlEvents() {
loop:
	for {
		event, more := <-p.controlPlaneClient.Events()
		if !more {
			p.logger.Info("Closed control plane events channel")
			break loop
		}

		switch ev := event.(type) {
		case csampler.ConfigUpdate:
			p.updateConfig(ev.Config)
			p.configUpdates++
		case csampler.StateUpdate:
			switch ev.State {
			case csampler.Registered:
				p.logger.Debug("Sampler registered with server")
			case csampler.Unregistered:
				p.logger.Debug("Sampler deregistered from server")
			}
		default:
			p.logger.Info(fmt.Sprintf("Received unknown event type %T", ev))
		}
	}
}

func (p *Sampler) updateStats(period time.Duration) {
	ticker := time.NewTicker(period)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := p.controlPlaneClient.UpdateStats(context.Background(), p.samplingStats); err != nil {
				p.logger.Error(fmt.Sprintf("Error updating stats: %s", err))
			}
		}
	}
}

func (p *Sampler) updateConfig(config control.SamplerConfig) {
	// configure limiter in
	if config.LimiterIn != nil {
		limit := rate.Limit(config.LimiterIn.Limit)
		if limit == -1 {
			limit = rate.Inf
		}
		p.limiterIn = rate.NewLimiter(limit, int(config.LimiterIn.Limit))
	}

	// configure sampler in
	if config.SamplingIn != nil {
		switch config.SamplingIn.SamplingType {
		case control.DeterministicSamplingType:
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
			builtRule, err := p.buildSamplingRule(stream.StreamRule)
			if err != nil {
				p.logger.Error(fmt.Sprintf("couldn't build sampling rule %+v: %s", stream, err))
				continue
			}

			newStreams[stream.UID] = streamConfig{
				rule:            builtRule,
				exportRawSample: stream.ExportRawSamples,
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
		p.limiterOut = rate.NewLimiter(limit, int(config.LimiterOut.Limit))
	}

	if config.Digests != nil {
		p.digester.SetDigestsConfig(config.Digests)
	}
}

func (p *Sampler) buildSamplingRule(streamRule control.Rule) (*rule.Rule, error) {
	switch streamRule.Lang {
	case control.SrlCel:
		builtRule, err := p.ruleBuilder.Build(streamRule.Expression)
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
	rawSample.SetStreams(streams)
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

func (p *Sampler) sample(ctx context.Context, key string, sampleData *data.Data) (bool, error) {
	p.samplingStats.SamplesEvaluated++

	if p.limiterIn != nil && !p.limiterIn.Allow() {
		return false, nil
	}

	// key acts as the deterministic sampler determinant
	if p.samplerIn != nil && !p.samplerIn.Sample(key) {
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
			err := p.exportRawSample(ctx, streams, key, sampleData)
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

func (p *Sampler) Sample(ctx context.Context, smpl defs.Sample) bool {
	if p.configUpdates == 0 {
		return false
	}

	var (
		sampleData *data.Data
	)
	switch smpl.Type {
	case defs.JSONSampleType:
		sampleData = data.NewSampleDataFromJSON(smpl.JSON)
	case defs.NativeSampleType:
		sampleData = data.NewSampleDataFromNative(smpl.Native)
	case defs.ProtoSampleType:
		sampleData = data.NewSampleDataFromProto(smpl.Proto)
	default:
		return false
	}

	sampled, err := p.sample(ctx, smpl.Key, sampleData)
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

	if err := p.exporter.Close(context.Background()); err != nil {
		return fmt.Errorf("error closing samples dp. %w", err)
	}

	return nil
}
