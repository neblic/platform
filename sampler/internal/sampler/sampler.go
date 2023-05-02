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
	"golang.org/x/time/rate"
	"google.golang.org/protobuf/proto"
)

var _ defs.Sampler = (*Sampler)(nil)

type Sampler struct {
	name         string
	resourceName string

	limiterIn *rate.Limiter
	// TODO: do not use a mutex and instead atomically upadte the list of configured streams
	streamsM   sync.Mutex
	streams    map[data.SamplerStreamUID]*rule.Rule
	limiterOut *rate.Limiter

	controlPlaneClient *csampler.Sampler
	exporter           exporter.Exporter
	ruleBuilder        *rule.Builder
	samplingStats      data.SamplerSamplingStats

	logger logging.Logger
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

	p := &Sampler{
		name:         opts.Name,
		resourceName: opts.Resource,

		ruleBuilder: ruleBuilder,
		streams:     make(map[data.SamplerStreamUID]*rule.Rule),

		controlPlaneClient: controlPlaneClient,
		limiterIn:          rate.NewLimiter(rate.Limit(opts.LimiterInLimit), int(opts.LimiterInLimit)),
		exporter:           opts.Exporter,
		limiterOut:         rate.NewLimiter(rate.Limit(opts.LimiterOutLimit), int(opts.LimiterOutLimit)),

		logger: logger.With("sampler_name", opts.Name, "sampler_uid", controlPlaneClient.UID()),
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

func (p *Sampler) SampleJSON(ctx context.Context, jsonSample string) (bool, error) {
	s, err := rule.NewEvalSampleFromJSON(jsonSample)
	if err != nil {
		return false, err
	}

	return p.sample(ctx, s)
}

func (p *Sampler) SampleNative(ctx context.Context, nativeSample any) (bool, error) {
	s, err := rule.NewEvalSampleFromNative(nativeSample)
	if err != nil {
		return false, err
	}

	return p.sample(ctx, s)
}

func (p *Sampler) SampleProto(ctx context.Context, protoSample proto.Message) (bool, error) {
	s, err := rule.NewEvalSampleFromProto(protoSample)
	if err != nil {
		return false, err
	}

	return p.sample(ctx, s)
}

func (p *Sampler) sample(ctx context.Context, evalSample *rule.EvalSample) (bool, error) {
	p.samplingStats.SamplesEvaluated++

	if !p.limiterIn.Allow() {
		return false, nil
	}

	//TODO: Apply sampler_in

	// assign sample to all matching streams based on their rules
	p.streamsM.Lock()
	defer p.streamsM.Unlock()
	var matches []sample.Match
	for streamUID, streamRule := range p.streams {
		if match, err := streamRule.Eval(ctx, evalSample); err != nil {
			p.logger.Debug(fmt.Sprintf("error evaluating sample: %s", err))
		} else if match {
			matches = append(matches, sample.Match{
				StreamUID: streamUID,
			})
		}
	}

	if len(matches) > 0 {
		if p.limiterOut.Allow() {
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
	}

	return false, nil
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
