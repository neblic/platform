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

	ruleBuilder *rule.Builder

	rules         map[data.SamplerSamplingRuleUID]*rule.Rule
	rulesM        sync.Mutex
	samplingStats data.SamplerSamplingStats

	controlPlaneClient *csampler.Sampler
	exporter           exporter.Exporter
	limiter            *rate.Limiter

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
		rules:       make(map[data.SamplerSamplingRuleUID]*rule.Rule),

		controlPlaneClient: controlPlaneClient,
		exporter:           opts.Exporter,
		limiter:            rate.NewLimiter(rate.Limit(opts.RateLimit), int(opts.RateBurst)),

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
	// replace all existing rules
	if config.SamplingRules != nil {
		p.rulesM.Lock()
		defer p.rulesM.Unlock()

		rules := make(map[data.SamplerSamplingRuleUID]*rule.Rule)
		p.rules = rules

		for _, sRule := range config.SamplingRules {
			builtRule, err := p.buildSamplingRule(sRule)
			if err != nil {
				p.logger.Error(fmt.Sprintf("couldn't build sampling rule %+v: %s", sRule, err))
				continue
			}

			rules[sRule.UID] = builtRule
		}
	}

	// configure limiter
	if config.SamplingRate != nil {
		limit := rate.Limit(config.SamplingRate.Limit)
		if limit == -1 {
			limit = rate.Inf
		}
		p.limiter = rate.NewLimiter(limit, int(config.SamplingRate.Burst))
	}
}

func (p *Sampler) buildSamplingRule(samplingRule data.SamplingRule) (*rule.Rule, error) {
	switch samplingRule.Lang {
	case data.SrlCel:
		builtRule, err := p.ruleBuilder.Build(samplingRule.Rule)
		if err != nil {
			return nil, fmt.Errorf("couldn't build CEL rule %s: %s", samplingRule.Rule, err)
		}

		return builtRule, nil
	case data.SrlUnknown:
		fallthrough
	default:
		return nil, fmt.Errorf("couldn't build rule with unknown type %s", samplingRule.Lang.String())
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

	p.rulesM.Lock()
	defer p.rulesM.Unlock()

	var matches []sample.Match
	for samplingRuleUID, samplingRule := range p.rules {
		if match, err := samplingRule.Eval(ctx, evalSample); err != nil {
			p.logger.Debug(fmt.Sprintf("error evaluating sample: %s", err))
		} else if match {
			matches = append(matches, sample.Match{
				SamplingRuleUID: samplingRuleUID,
			})
		}
	}

	if len(matches) > 0 {
		if p.limiter.Allow() {
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
