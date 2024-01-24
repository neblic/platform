package sampler

import (
	"slices"
	"time"

	"github.com/google/uuid"
	"github.com/neblic/platform/controlplane/control"
)

const allStreamName = "all"
const allStreamCelRule = "true"
const valueDigestName = "value"
const structDigestName = "struct"
const DLQEventName = "sample_sent_to_dlq"

// re-exported known tags
const (
	ProducerTag = control.ProducerTag
	ConsumerTag = control.ConsumerTag
	RequestTag  = control.RequestTag
	ResponseTag = control.ResponseTag
	// DLQTag is used to tag samples that are sent to a dead letter queue
	// Automatically enables an event that notifies when a sample is sent to a DLQ
	DLQTag = control.DLQTag
)

type options struct {
	initialConfig     control.SamplerConfigUpdate
	tags              []string
	updateStatsPeriod time.Duration
}

func newDefaultStreamUpdate(uid control.SamplerStreamUID) control.StreamUpdate {
	return control.StreamUpdate{
		Op: control.StreamUpsert,
		Stream: control.Stream{
			UID:  uid,
			Name: allStreamName,
			StreamRule: control.Rule{
				Lang:       control.SrlCel,
				Expression: allStreamCelRule,
			},
		},
	}
}

func newDefaultStructDigestUpdate(streamUID control.SamplerStreamUID, location control.ComputationLocation) control.DigestUpdate {
	return control.DigestUpdate{
		Op: control.DigestUpsert,
		Digest: control.Digest{
			UID:                 control.SamplerDigestUID(uuid.NewString()),
			Name:                structDigestName,
			StreamUID:           streamUID,
			FlushPeriod:         time.Second * time.Duration(60),
			ComputationLocation: location,
			Type:                control.DigestTypeSt,
			St: &control.DigestSt{
				MaxProcessedFields: int(100),
			},
		},
	}
}

func newDefaultValueDigestUpdate(streamUID control.SamplerStreamUID, location control.ComputationLocation) control.DigestUpdate {
	return control.DigestUpdate{
		Op: control.DigestUpsert,
		Digest: control.Digest{
			UID:                 control.SamplerDigestUID(uuid.NewString()),
			Name:                valueDigestName,
			StreamUID:           control.SamplerStreamUID(streamUID),
			FlushPeriod:         time.Second * time.Duration(60),
			ComputationLocation: location,
			Type:                control.DigestTypeValue,
			Value: &control.DigestValue{
				MaxProcessedFields: int(100),
			},
		},
	}
}

func newDefaultOptions() *options {
	initialConfig := control.NewSamplerConfigUpdate()
	initialConfig.LimiterIn = &control.LimiterConfig{Limit: 100}
	initialConfig.LimiterOut = &control.LimiterConfig{Limit: 10}
	initialStreamUID := control.SamplerStreamUID(uuid.NewString())
	initialConfig.StreamUpdates = []control.StreamUpdate{
		newDefaultStreamUpdate(initialStreamUID),
	}
	initialConfig.DigestUpdates = []control.DigestUpdate{
		newDefaultStructDigestUpdate(initialStreamUID, control.ComputationLocationSampler),
	}

	return &options{
		initialConfig:     initialConfig,
		updateStatsPeriod: time.Second * time.Duration(5),
	}
}

type Option interface {
	apply(*options)
}

type funcOption struct {
	f func(*options)
}

func (fco *funcOption) apply(co *options) {
	fco.f(co)
}

func newFuncOption(f func(*options)) *funcOption {
	return &funcOption{
		f: f,
	}
}

// WithInitialLimiterInLimit sets the initial limiter in rate limit. This configuration
// is only used the first time a sampler is registered with a server, posterior executions
// will use the configuration stored in the server and the provided configuration will be
// ignored.
func WithInitialLimiterInLimit(l int32) Option {
	return newFuncOption(func(o *options) {
		o.initialConfig.LimiterIn = &control.LimiterConfig{
			Limit: l,
		}
	})
}

// WithInitialDeterministicSamplingIn defines a deterministic sampling strategy which will be applied
// when a sample is received and before processing it in any way (e.g. before determining if a sample belongs
// to a stream which would require parsing it and evaluating the stream rules).
// Sampling is performed after the input limiter has been applied.
// This configuration is only used the first time a sampler is registered with a server, posterior executions
// will use the configuration stored in the server and the provided configuration will be
// ignored.
func WithInitialDeterministicSamplingIn(samplingRate int32) Option {
	return newFuncOption(func(o *options) {
		o.initialConfig.SamplingIn = &control.SamplingConfig{
			SamplingType: control.DeterministicSamplingType,
			DeterministicSampling: control.DeterministicSamplingConfig{
				SampleRate: samplingRate,
			},
		}
	})
}

// WithInitialLimiterInLimit sets the initial limiter in rate limit. This configuration
// is only used the first time a sampler is registered with a server, posterior executions
// will use the configuration stored in the server and the provided configuration will be
// ignored.
func WithInitialLimiterOutLimit(l int32) Option {
	return newFuncOption(func(o *options) {
		o.initialConfig.LimiterOut = &control.LimiterConfig{
			Limit: l,
		}
	})
}

// WithInitalStructDigest enables the computation of struct digest. Location defines where this computation
// must take place (in the sampler of in the collector). In case of computing the digest in the collector,
// the raw samples are exported (that can have an impact in the sampler performance)
func WithInitialStructDigest(location control.ComputationLocation) Option {
	return newFuncOption(func(o *options) {
		// Check if the default struct digest is present
		structDigestIdx := slices.IndexFunc(o.initialConfig.DigestUpdates, func(digestUpdate control.DigestUpdate) bool {
			return digestUpdate.Digest.Type == control.DigestTypeSt && digestUpdate.Digest.Name == structDigestName
		})

		// If the default struct digest is not present, add it
		if structDigestIdx == -1 {
			allStreamIdx := slices.IndexFunc(o.initialConfig.StreamUpdates, func(streamUpdate control.StreamUpdate) bool {
				return streamUpdate.Stream.Name == allStreamName && streamUpdate.Stream.StreamRule.Expression == allStreamCelRule
			})

			var initialStreamUID control.SamplerStreamUID
			if allStreamIdx == -1 {
				initialStreamUID = control.SamplerStreamUID(uuid.NewString())
				initialStreamUpdate := newDefaultStreamUpdate(initialStreamUID)

				// In case of computing the digest in the collector, we need to export the raw samples
				if location == control.ComputationLocationCollector {
					initialStreamUpdate.Stream.ExportRawSamples = true
				}

				o.initialConfig.StreamUpdates = append(o.initialConfig.StreamUpdates, initialStreamUpdate)
			} else {
				initialStreamUID = o.initialConfig.StreamUpdates[allStreamIdx].Stream.UID
			}

			o.initialConfig.DigestUpdates = append(o.initialConfig.DigestUpdates, newDefaultStructDigestUpdate(initialStreamUID, location))
		}
	})
}

// WithInitalValueDigest enables the computation of value digest. Location defines where this computation
// must take place (in the sampler of in the collector). In case of computing the digest in the collector,
// the raw samples are exported (that can have an impact in the sampler performance)
func WithInitalValueDigest(location control.ComputationLocation) Option {
	return newFuncOption(func(o *options) {
		// Check if the default value digest is present
		valueDigestIdx := slices.IndexFunc(o.initialConfig.DigestUpdates, func(digestUpdate control.DigestUpdate) bool {
			return digestUpdate.Digest.Type == control.DigestTypeValue && digestUpdate.Digest.Name == valueDigestName
		})

		// If the default struct value is not present, add it
		if valueDigestIdx == -1 {
			allStreamIdx := slices.IndexFunc(o.initialConfig.StreamUpdates, func(streamUpdate control.StreamUpdate) bool {
				return streamUpdate.Stream.Name == allStreamName && streamUpdate.Stream.StreamRule.Expression == allStreamCelRule
			})

			var initialStreamUID control.SamplerStreamUID
			if allStreamIdx == -1 {
				initialStreamUID = control.SamplerStreamUID(uuid.NewString())
				initialStreamUpdate := newDefaultStreamUpdate(initialStreamUID)

				// In case of computing the digest in the collector, we need to export the raw samples
				if location == control.ComputationLocationCollector {
					initialStreamUpdate.Stream.ExportRawSamples = true
				}

				o.initialConfig.StreamUpdates = append(o.initialConfig.StreamUpdates, initialStreamUpdate)
			} else {
				initialStreamUID = o.initialConfig.StreamUpdates[allStreamIdx].Stream.UID
			}

			o.initialConfig.DigestUpdates = append(o.initialConfig.DigestUpdates, newDefaultValueDigestUpdate(initialStreamUID, location))
		}
	})
}

// WithoutDefaultInitialConfig avoids setting the default 'all' stream and digest. This configuration
// is only used the first time a sampler is registered with a server, posterior executions
// will use the configuration stored in the server and the provided configuration will be
// ignored.
func WithoutDefaultInitialConfig() Option {
	return newFuncOption(func(o *options) {
		o.initialConfig.StreamUpdates = []control.StreamUpdate{}
		o.initialConfig.DigestUpdates = []control.DigestUpdate{}
	})
}

// WithUpdateStatsPeriod specifies the period to send sampler stats to server
// If the provided period is less than a second, it will be set to 1 second.
func WithUpdateStatsPeriod(p time.Duration) Option {
	return newFuncOption(func(o *options) {
		if p < time.Second {
			p = time.Second
		}
		o.updateStatsPeriod = p
	})
}

func applyDLQInitialConfig(o *options) {
	// Check if the default DLQ event is present
	dlqEventIdx := slices.IndexFunc(o.initialConfig.EventUpdates, func(eventUpdate control.EventUpdate) bool {
		return eventUpdate.Event.Name == DLQEventName
	})

	// If the default DLQ event is not present, add it
	if dlqEventIdx == -1 {
		o.initialConfig.EventUpdates = append(o.initialConfig.EventUpdates, control.EventUpdate{
			Op: control.EventUpsert,
			Event: control.Event{
				UID:        control.SamplerEventUID(uuid.NewString()),
				Name:       DLQEventName,
				StreamUID:  control.SamplerStreamUID(allStreamName),
				SampleType: control.RawSampleType,
				Rule: control.Rule{
					Lang:       control.SrlCel,
					Expression: "true",
				},
				Limiter: control.LimiterConfig{
					Limit: 1,
				},
			},
		})
	}
}

// WithTags can be used to specify tags to classify the Sampler or to specify its type.
// There are known tags that can be used to identify the Sampler type and may be used
// by the platform to provide additional functionality automnatically.
// See the tags constants defined in this file.
func WithTags(tags ...string) Option {
	return newFuncOption(func(o *options) {
		o.tags = tags
		for _, tag := range tags {
			switch tag {
			case DLQTag:
				applyDLQInitialConfig(o)
			}
		}
	})
}
