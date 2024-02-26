package control

import (
	"errors"

	"github.com/neblic/platform/controlplane/protos"
)

type SamplerConfigUpdateReset struct {
	LimiterIn  bool
	SamplingIn bool
	Streams    bool
	LimiterOut bool
	Digests    bool
	Events     bool
}

func NewSamplerConfigUpdateResetFromProto(protoReset *protos.ClientSamplerConfigUpdate_Reset) SamplerConfigUpdateReset {
	if protoReset == nil {
		return SamplerConfigUpdateReset{}
	}

	return SamplerConfigUpdateReset{
		Streams:    protoReset.GetStreams(),
		LimiterIn:  protoReset.GetLimiterIn(),
		SamplingIn: protoReset.GetSamplingIn(),
		LimiterOut: protoReset.GetLimiterOut(),
		Digests:    protoReset.GetDigests(),
		Events:     protoReset.GetEvents(),
	}
}

func (scr SamplerConfigUpdateReset) ToProto() *protos.ClientSamplerConfigUpdate_Reset {
	return &protos.ClientSamplerConfigUpdate_Reset{
		Streams:    scr.Streams,
		LimiterIn:  scr.LimiterIn,
		SamplingIn: scr.SamplingIn,
		LimiterOut: scr.LimiterOut,
		Digests:    scr.Digests,
		Events:     scr.Events,
	}
}

type SamplerConfigUpdate struct {
	// If a field is set to true, it means that the field is reset to its default.
	// If a configuration option is reset and set in the same request, it will be
	// first resetted and then set to its new value.
	Reset SamplerConfigUpdateReset

	// All fields are optional. If a field is nil, it means that the field is not updated.
	StreamUpdates []StreamUpdate
	LimiterIn     *LimiterConfig
	SamplingIn    *SamplingConfig
	LimiterOut    *LimiterConfig
	DigestUpdates []DigestUpdate
	EventUpdates  []EventUpdate
}

func NewSamplerConfigUpdate() SamplerConfigUpdate {
	return SamplerConfigUpdate{}
}

func NewSamplerConfigUpdateFromProto(protoUpdate *protos.ClientSamplerConfigUpdate) SamplerConfigUpdate {
	if protoUpdate == nil {
		return NewSamplerConfigUpdate()
	}

	var limiterIn *LimiterConfig
	if protoUpdate.GetLimiterIn() != nil {
		newSrIn := NewLimiterFromProto(protoUpdate.GetLimiterIn())
		limiterIn = &newSrIn
	}

	var samplingConfigIn *SamplingConfig
	if protoUpdate.GetSamplingIn() != nil {
		newScIn := NewSamplingConfigFromProto(protoUpdate.GetSamplingIn())
		samplingConfigIn = &newScIn
	}

	var streamUpdates []StreamUpdate
	for _, streamUpdate := range protoUpdate.GetStreamUpdates() {
		streamUpdates = append(streamUpdates, NewStreamUpdateFromProto(streamUpdate))
	}

	var limiterOut *LimiterConfig
	if protoUpdate.GetLimiterOut() != nil {
		newSrOut := NewLimiterFromProto(protoUpdate.GetLimiterOut())
		limiterOut = &newSrOut
	}

	var digestUpdates []DigestUpdate
	for _, digestUpdate := range protoUpdate.GetDigestUpdates() {
		digestUpdates = append(digestUpdates, NewDigestUpdateFromProto(digestUpdate))
	}

	var eventUpdates []EventUpdate
	for _, eventUpdate := range protoUpdate.GetEventUpdates() {
		eventUpdates = append(eventUpdates, NewEventUpdateFromProto(eventUpdate))
	}

	return SamplerConfigUpdate{
		Reset: NewSamplerConfigUpdateResetFromProto(protoUpdate.GetReset_()),

		LimiterIn:     limiterIn,
		SamplingIn:    samplingConfigIn,
		StreamUpdates: streamUpdates,
		LimiterOut:    limiterOut,
		DigestUpdates: digestUpdates,
		EventUpdates:  eventUpdates,
	}
}

func (scu SamplerConfigUpdate) ToProto() *protos.ClientSamplerConfigUpdate {
	var protoLimiterIn *protos.Limiter
	if scu.LimiterIn != nil {
		protoLimiterIn = scu.LimiterIn.ToProto()
	}

	var protoSamplingIn *protos.Sampling
	if scu.SamplingIn != nil {
		protoSamplingIn = scu.SamplingIn.ToProto()
	}

	var protoUpdateStreams []*protos.ClientStreamUpdate
	for _, streamUpdate := range scu.StreamUpdates {
		protoUpdateStreams = append(protoUpdateStreams, streamUpdate.ToProto())
	}

	var protoLimiterOut *protos.Limiter
	if scu.LimiterOut != nil {
		protoLimiterOut = scu.LimiterOut.ToProto()
	}

	var protoUpdateDigests []*protos.ClientDigestUpdate
	for _, digestUpdate := range scu.DigestUpdates {
		protoUpdateDigests = append(protoUpdateDigests, digestUpdate.ToProto())
	}

	var protoUpdateEvents []*protos.ClientEventUpdate
	for _, eventUpdate := range scu.EventUpdates {
		protoUpdateEvents = append(protoUpdateEvents, eventUpdate.ToProto())
	}

	return &protos.ClientSamplerConfigUpdate{
		Reset_: scu.Reset.ToProto(),

		StreamUpdates: protoUpdateStreams,
		LimiterIn:     protoLimiterIn,
		SamplingIn:    protoSamplingIn,
		LimiterOut:    protoLimiterOut,
		DigestUpdates: protoUpdateDigests,
		EventUpdates:  protoUpdateEvents,
	}
}

func (scu SamplerConfigUpdate) IsValid() error {
	var errs error

	for _, digestUpdate := range scu.DigestUpdates {
		err := digestUpdate.IsValid()
		errs = errors.Join(errs, err)
	}

	for _, eventUpdate := range scu.EventUpdates {
		err := eventUpdate.IsValid()
		errs = errors.Join(errs, err)
	}

	for _, streamUpdate := range scu.StreamUpdates {
		err := streamUpdate.IsValid()
		errs = errors.Join(errs, err)
	}

	return errs
}
