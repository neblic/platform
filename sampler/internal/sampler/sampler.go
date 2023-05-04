package sampler

import (
	"context"
	"fmt"
	"sync"
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

	limiterIn *rate.Limiter
	samplerIn sampling.Sampler
	// TODO: do not use a mutex and instead atomically upadte the list of configured streams
	streamsM   sync.Mutex
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
	opts *Options,
	logger logging.Logger,
) (*Sampler, error) {
	ruleBuilder, err := rule.NewBuilder(opts.Schema)
	if err != nil {
		return nil, fmt.Errorf("couldn't build CEL rule builder: %w", err)
	}

	var clientOpts []csampler.Option
	clientOpts = append(clientOpts, csampler.WithLogger(logger))
	if opts.EnableTLS {
		clientOpts = append(clientOpts, csampler.WithTLS())
	}

	// provider already checked that it is a valid type
	switch opts.Auth.Type {
	case "bearer":
		clientOpts = append(clientOpts, csampler.WithAuthBearer(opts.Auth.Bearer.Token))
	}

	// TODO: Once we have refactored the sampler client to reuse the same grpc connection internally,
	// move control plane connection to provider
	controlPlaneClient := csampler.New(opts.Name, opts.Resource, clientOpts...)

	var samplerIn sampling.Sampler
	switch opts.SamplingIn.SamplingType {
	case data.DeterministicSamplingType:
		deterministicSampler, err := sampling.NewDeterministicSampler(
			uint(opts.SamplingIn.DeterministicSampling.SampleRate),
			opts.SamplingIn.DeterministicSampling.SampleEmptyDeterminant)
		if err != nil {
			return nil, fmt.Errorf("couldn't initialize the deterministic sampler: %w", err)
		}
		samplerIn = deterministicSampler
	}

	p := &Sampler{
		name:         opts.Name,
		resourceName: opts.Resource,

		ruleBuilder: ruleBuilder,
		streams:     make(map[data.SamplerStreamUID]*rule.Rule),

		controlPlaneClient: controlPlaneClient,
		limiterIn:          rate.NewLimiter(rate.Limit(opts.LimiterIn.Limit), int(opts.LimiterIn.Limit)),
		samplerIn:          samplerIn,
		exporter:           opts.Exporter,
		limiterOut:         rate.NewLimiter(rate.Limit(opts.LimiterOut.Limit), int(opts.LimiterOut.Limit)),

		errFwrder: opts.ErrFwrder,
		logger:    logger.With("sampler_name", opts.Name, "sampler_uid", controlPlaneClient.UID()),
	}

	go p.listenControlEvents()

	err = controlPlaneClient.Connect(opts.ControlPlaneAddr)
	if err != nil {
		return nil, fmt.Errorf("couldn't connect to the control plane: %w", err)
	}

	go p.updateStats(opts.UpdateStatsPeriod)

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
		p.streamsM.Lock()
		defer p.streamsM.Unlock()

		p.streams = make(map[data.SamplerStreamUID]*rule.Rule)
		for _, stream := range config.Streams {
			builtRule, err := p.buildSamplingRule(stream.StreamRule)
			if err != nil {
				p.logger.Error(fmt.Sprintf("couldn't build sampling rule %+v: %s", stream, err))
				continue
			}

			p.streams[stream.UID] = builtRule
		}
	}

	// configure limiter out
	if config.LimiterOut != nil {
		limit := rate.Limit(config.LimiterOut.Limit)
		if limit == -1 {
			limit = rate.Inf
		}
		p.limiterOut = rate.NewLimiter(limit, int(config.LimiterOut.Limit))
	}
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

func (p *Sampler) Sample(ctx context.Context, sample defs.Sample) bool {
	var (
		evalSample *rule.EvalSample
		err        error
	)
	switch sample.Type {
	case defs.JsonSampleType:
		evalSample, err = rule.NewEvalSampleFromJSON(sample.Json)
	case defs.NativeSampleType:
		evalSample, err = rule.NewEvalSampleFromNative(sample.Native)
	case defs.ProtoSampleType:
		evalSample, err = rule.NewEvalSampleFromProto(sample.Proto)
	default:
		return false
	}
	if err != nil {
		p.forwardError(err)
		return false
	}

	sampled, err := p.sample(ctx, evalSample, sample.Determinant)
	if err != nil {
		p.forwardError(err)
		return false
	}

	return sampled
}

func (p *Sampler) sample(ctx context.Context, evalSample *rule.EvalSample, determinant string) (bool, error) {
	p.samplingStats.SamplesEvaluated++

	if p.limiterIn != nil && !p.limiterIn.Allow() {
		return false, nil
	}

	if p.samplerIn != nil && !p.samplerIn.Sample(determinant) {
		return false, nil
	}

	// Optimization: if there are no output tokens available, no need to do anything since it won't be sampled
	if p.limiterOut != nil && p.limiterOut.Tokens() == 0 {
		return false, nil
	}

	// assign sample to all matching streams based on their rules
	p.streamsM.Lock()
	defer p.streamsM.Unlock()
	var matches []sample.Match
	for streamUID, streamRule := range p.streams {
		if match, err := streamRule.Eval(ctx, evalSample); err != nil {
			p.forwardError(err)
		} else if match {
			matches = append(matches, sample.Match{
				StreamUID: streamUID,
			})
		}
	}

	if len(matches) > 0 {
		if p.limiterOut != nil && !p.limiterOut.Allow() {
			return false, nil
		}

		if err := p.exporter.Export(ctx, []sample.ResourceSamples{{
			ResourceName: p.resourceName,
			SamplerName:  p.name,
			SamplersSamples: []sample.SamplerSamples{{
				Samples: []sample.Sample{{
					Ts:      time.Now(),
					Data:    evalSample.AsMap(),
					Matches: matches,
				}},
			}},
		}}); err != nil {
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

func (p *Sampler) Close() error {
	if err := p.controlPlaneClient.Close(closeTimeout); err != nil {
		return fmt.Errorf("error closing control plane client: %w", err)
	}

	if err := p.exporter.Close(context.Background()); err != nil {
		return fmt.Errorf("error closing samples exporter: %w", err)
	}

	return nil
}
