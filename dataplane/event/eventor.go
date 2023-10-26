package event

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/neblic/platform/controlplane/control"
	"github.com/neblic/platform/dataplane/sample"
	"github.com/neblic/platform/internal/pkg/rule"
	"github.com/neblic/platform/sampler/defs"
	"golang.org/x/exp/slices"
)

type Settings struct {
	ResourceName string
	SamplerName  string
}

type event struct {
	rule           rule.Rule
	uid            control.SamplerEventUID
	streamUID      control.SamplerStreamUID
	ruleExpression string
}

type Eventor struct {
	resourceName string
	samplerName  string
	ruleBuilder  rule.Builder
	events       map[control.SamplerEventUID]*event
}

func NewEventor(settings Settings) (*Eventor, error) {
	ruleBuilder, err := rule.NewBuilder(defs.NewDynamicSchema())
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

func (e *Eventor) SetEventsConfig(eventsCfgs map[control.SamplerEventUID]control.Event) error {
	var errs error

	for eventUID, eventCfg := range eventsCfgs {
		rule, err := e.ruleBuilder.Build(eventCfg.Rule.Expression)
		if err != nil {
			errs = errors.Join(errs, err)
			continue
		}

		e.events[eventUID] = &event{
			rule:           *rule,
			uid:            eventCfg.UID,
			streamUID:      eventCfg.StreamUID,
			ruleExpression: eventCfg.Rule.Expression,
		}
	}

	return errs
}

// ProessSample iterates over all the raw flows in the sampler logs and creates events when necessary.
// Generated events are appended to the provided sampler logs
func (e *Eventor) ProcessSample(samplerLogs sample.SamplerOTLPLogs) error {
	// List of computed events
	var errs error

	// Iterate and append events inplace. As range function does not iterate over appended elements after
	// it's call, new events will not be visited.
	sample.RangeSamplerLogsWithType[sample.RawSampleOTLPLog](samplerLogs, func(rawSample sample.RawSampleOTLPLog) {
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
					otlpLog := samplerLogs.AppendEventOTLPLog()
					otlpLog.SetUID(event.uid)
					otlpLog.SetTimestamp(time.Now())
					otlpLog.SetStreams([]control.SamplerStreamUID{event.streamUID})
					otlpLog.SetSampleRawData(rawSample.SampleEncoding(), rawSample.SampleRawData())
					otlpLog.SetRuleExpression(event.ruleExpression)
				}
			}
		}
	})

	return errs
}
