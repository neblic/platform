package event

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/neblic/platform/controlplane/control"
	dsample "github.com/neblic/platform/dataplane/sample"
	"github.com/neblic/platform/internal/pkg/rule"
	"github.com/neblic/platform/sampler/sample"
	"golang.org/x/exp/slices"
	"golang.org/x/time/rate"
)

type Settings struct {
	ResourceName string
	SamplerName  string
}

type event struct {
	uid            control.SamplerEventUID
	streamUID      control.SamplerStreamUID
	rule           rule.Rule
	ruleExpression string
	limiter        rate.Limiter
}

type Eventor struct {
	resourceName string
	samplerName  string
	ruleBuilder  rule.Builder
	events       map[control.SamplerEventUID]*event
}

func NewEventor(settings Settings) (*Eventor, error) {
	ruleBuilder, err := rule.NewBuilder(sample.NewDynamicSchema(), rule.CheckFunctions)
	if err != nil {
		return nil, fmt.Errorf("cannot create rule builder: %w", err)
	}

	return &Eventor{
		resourceName: settings.ResourceName,
		samplerName:  settings.SamplerName,
		ruleBuilder:  *ruleBuilder,
		events:       make(map[control.SamplerEventUID]*event),
	}, nil
}

func (e *Eventor) newEventFrom(eventCfg control.Event) (*event, error) {
	rule, err := e.ruleBuilder.Build(eventCfg.Rule.Expression)
	if err != nil {
		return nil, fmt.Errorf("cannot create rule: %w", err)
	}

	return &event{
		rule:           *rule,
		uid:            eventCfg.UID,
		streamUID:      eventCfg.StreamUID,
		ruleExpression: eventCfg.Rule.Expression,
		limiter:        *rate.NewLimiter(rate.Limit(eventCfg.Limiter.Limit), int(eventCfg.Limiter.Limit)),
	}, nil
}

func (e *Eventor) SetEventsConfig(eventsCfgs map[control.SamplerEventUID]control.Event) error {
	var errs error

	// Add new events
	for eventUID, eventCfg := range eventsCfgs {
		if _, ok := e.events[eventUID]; !ok {
			newEvent, err := e.newEventFrom(eventCfg)
			if err != nil {
				errs = errors.Join(errs, err)
				continue
			}
			e.events[eventUID] = newEvent
		}
	}

	// Update existing events with different rule expression
	for eventUID, eventCfg := range eventsCfgs {
		existingEvent, ok := e.events[eventUID]
		if ok && existingEvent.ruleExpression != eventCfg.Rule.Expression {
			newEvent, err := e.newEventFrom(eventCfg)
			if err != nil {
				errs = errors.Join(errs, err)
				continue
			}
			e.events[eventUID] = newEvent
		}
	}

	// Remove old events
	for eventUID := range e.events {
		if _, ok := eventsCfgs[eventUID]; !ok {
			delete(e.events, eventUID)
		}
	}

	return errs
}

// ProessSample iterates over all the raw flows in the sampler logs and creates events when necessary.
// Generated events are appended to the provided sampler logs
func (e *Eventor) ProcessSample(samplerLogs dsample.SamplerOTLPLogs) error {
	var errs error

	// Iterate and append events in-place. As range function does not iterate over appended elements after
	// it's call, new events will not be visited.
	dsample.RangeSamplerLogsWithType[dsample.RawSampleOTLPLog](samplerLogs, func(rawSample dsample.RawSampleOTLPLog) {
		for _, event := range e.events {
			if slices.Contains(rawSample.Streams(), event.streamUID) {
				data, err := rawSample.SampleData()
				if err != nil {
					errs = errors.Join(errs, err)
					continue
				}

				ruleMatches, err := event.rule.Eval(context.Background(), data)
				if err != nil {
					errs = errors.Join(fmt.Errorf("eval(%s) -> %w", event.ruleExpression, err))
					continue
				}

				if ruleMatches {
					if event.limiter.Allow() {
						otlpLog := samplerLogs.AppendEventOTLPLog()
						otlpLog.SetUID(event.uid)
						otlpLog.SetTimestamp(time.Now())
						otlpLog.SetStreams([]control.SamplerStreamUID{event.streamUID})
						otlpLog.SetSampleKey(rawSample.SampleKey())
						otlpLog.SetSampleRawData(rawSample.SampleEncoding(), rawSample.SampleRawData())
						otlpLog.SetRuleExpression(event.ruleExpression)
					}
				}
			}
		}
	})

	return errs
}

func (e *Eventor) Close() {}
