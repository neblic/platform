package sampler

import (
	"context"
	"fmt"
	"time"

	"github.com/neblic/platform/controlplane/data"
	csampler "github.com/neblic/platform/controlplane/sampler"
	"github.com/neblic/platform/logging"
	"github.com/neblic/platform/sampler/defs"
	"github.com/neblic/platform/sampler/internal/rule"
	"github.com/neblic/platform/sampler/internal/sample"
	"github.com/neblic/platform/sampler/internal/sample/exporter"
	"github.com/neblic/platform/sampler/internal/sample/sampling"
	"golang.org/x/time/rate"
)

var _ defs.Sampler = (*Sampler)(nil)

type Sampler struct {
	name         string
	resourceName string

	limiterIn  *rate.Limiter
	samplerIn  sampling.Sampler
	streams    map[data.SamplerStreamUID]*rule.Rule
	limiterOut *rate.Limiter

	controlPlaneClient *csampler.Sampler
	exporter           exporter.Exporter
	ruleBuilder        *rule.Builder
	samplingStats      data.SamplerSamplingStats

	errFwrder chan error
	logger    logging.Logger
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

	var samplerIn sampling.Sampler
	switch settings.SamplingIn.SamplingType {
	case data.DeterministicSamplingType:
		deterministicSampler, err := sampling.NewDeterministicSampler(
			uint(settings.SamplingIn.DeterministicSampling.SampleRate),
			settings.SamplingIn.DeterministicSampling.SampleEmptyDeterminant)
		if err != nil {
			return nil, fmt.Errorf("couldn't initialize the deterministic sampler: %w", err)
		}
		samplerIn = deterministicSampler
	}

	p := &Sampler{
		name:         settings.Name,
		resourceName: settings.Resource,

		ruleBuilder: ruleBuilder,
		streams:     make(map[data.SamplerStreamUID]*rule.Rule),

		controlPlaneClient: controlPlaneClient,
		limiterIn:          rate.NewLimiter(rate.Limit(settings.LimiterIn.Limit), int(settings.LimiterIn.Limit)),
		samplerIn:          samplerIn,
		exporter:           settings.Exporter,
		limiterOut:         rate.NewLimiter(rate.Limit(settings.LimiterOut.Limit), int(settings.LimiterOut.Limit)),

		errFwrder: settings.ErrFwrder,
		logger:    logger.With("sampler_name", settings.Name, "sampler_uid", controlPlaneClient.UID()),
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

func (p *Sampler) updateConfig(config data.SamplerConfig) {
	// configure limiter in
	if config.LimiterIn != nil {
		limit := rate.Limit(config.LimiterIn.Limit)
		if limit == -1 {
			limit = rate.Inf
		}
		p.limiterIn = rate.NewLimiter(limit, int(config.LimiterIn.Limit))
	}

	// replace all existing streams
	if config.Streams != nil {
		newStreams := make(map[data.SamplerStreamUID]*rule.Rule)
		for _, stream := range config.Streams {
			builtRule, err := p.buildSamplingRule(stream.StreamRule)
			if err != nil {
				p.logger.Error(fmt.Sprintf("couldn't build sampling rule %+v: %s", stream, err))
				continue
			}

			newStreams[stream.UID] = builtRule
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

	// TODO: configure digesters
}

func (p *Sampler) buildSamplingRule(streamRule data.StreamRule) (*rule.Rule, error) {
	switch streamRule.Lang {
	case data.SrlCel:
		builtRule, err := p.ruleBuilder.Build(streamRule.Rule)
		if err != nil {
			return nil, fmt.Errorf("couldn't build CEL rule %s: %s", streamRule.Rule, err)
		}

		return builtRule, nil
	default:
		return nil, fmt.Errorf("couldn't build rule with unknown type %s", streamRule.Lang.String())
	}
}

func (p *Sampler) buildSample(sampleData *sample.Data, streams []data.SamplerStreamUID) (exporter.SamplerSamples, error) {
	dataJSON, err := sampleData.JSON()
	if err != nil {
		return exporter.SamplerSamples{}, fmt.Errorf("couldn't get sampler body: %w", err)
	}

	return exporter.SamplerSamples{
		ResourceName: p.resourceName,
		SamplerName:  p.name,
		Samples: []exporter.Sample{{
			Ts:       time.Now(),
			Type:     exporter.RawSampleType,
			Streams:  streams,
			Encoding: exporter.JSONSampleEncoding,
			Data:     []byte(dataJSON),
		}},
	}, nil
}

func (p *Sampler) sample(ctx context.Context, sampleData *sample.Data, determinant string) (bool, error) {
	p.samplingStats.SamplesEvaluated++

	if p.limiterIn != nil && !p.limiterIn.Allow() {
		return false, nil
	}

	if p.samplerIn != nil && !p.samplerIn.Sample(determinant) {
		return false, nil
	}

	// optimization: if there are no output tokens available, no need to do anything since it won't be sampled
	if p.limiterOut != nil && p.limiterOut.Tokens() == 0 {
		return false, nil
	}

	// assign sample to all matching streams based on their rules
	var streams []data.SamplerStreamUID
	for streamUID, streamRule := range p.streams {
		if match, err := streamRule.Eval(ctx, sampleData); err != nil {
			p.forwardError(err)
		} else if match {
			streams = append(streams, streamUID)
		}
	}

	if len(streams) > 0 {
		if p.limiterOut != nil && !p.limiterOut.Allow() {
			return false, nil
		}

		// TODO: hand over sample to digester
		// * it will handle what digests generate and when to export them, asynchronously

		// TODO: make sure the exporter follows a similar async handling logic of samples as the digester
		resourceSample, err := p.buildSample(sampleData, streams)
		if err != nil {
			return false, err
		}

		if err := p.exporter.Export(ctx, []exporter.SamplerSamples{resourceSample}); err != nil {
			return false, fmt.Errorf("failure to export samples: %w", err)
		}

		p.samplingStats.SamplesExported++

		return true, nil
	}

	return false, nil
}

func (p *Sampler) forwardError(err error) {
	if p.errFwrder != nil {
		select {
		case p.errFwrder <- err:
		default:
		}
	}
}

func (p *Sampler) Sample(ctx context.Context, smpl defs.Sample) bool {
	var (
		sampleData *sample.Data
	)
	switch smpl.Type {
	case defs.JSONSampleType:
		sampleData = sample.NewSampleDataFromJSON(smpl.JSON)
	case defs.NativeSampleType:
		sampleData = sample.NewSampleDataFromNative(smpl.Native)
	case defs.ProtoSampleType:
		sampleData = sample.NewSampleDataFromProto(smpl.Proto)
	default:
		return false
	}

	sampled, err := p.sample(ctx, sampleData, smpl.Determinant)
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
		return fmt.Errorf("error closing samples exporter: %w", err)
	}

	return nil
}
