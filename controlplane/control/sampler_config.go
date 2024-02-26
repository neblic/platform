package control

import (
	"time"

	"github.com/google/uuid"
	"github.com/neblic/platform/controlplane/protos"
)

// Used to get and update the sampler configuration.
//
// When sent by the server to update a sampler, only the fields that are present
// are updated. If a field is present, the previous value is replaced with the
// new one.
type SamplerConfig struct {
	Streams    Streams
	LimiterIn  *LimiterConfig
	SamplingIn *SamplingConfig
	LimiterOut *LimiterConfig
	Digests    Digests
	Events     Events
}

func NewSamplerConfig() *SamplerConfig {
	return &SamplerConfig{
		Streams: Streams{},
		Digests: Digests{},
		Events:  Events{},
	}
}

func NewImplicitSamplerConfig() *SamplerConfig {
	implicitConfig := NewSamplerConfig()
	streamUID := SamplerStreamUID(uuid.NewString())
	implicitConfig.Streams = Streams{
		streamUID: Stream{
			UID:  streamUID,
			Name: "all",
			StreamRule: Rule{
				Lang:       SrlUnknown,
				Expression: "",
			},
			ExportRawSamples: true,
			MaxSampleSize:    10240,
		},
	}
	structDigestUID := SamplerDigestUID(uuid.NewString())
	valueDigestUID := SamplerDigestUID(uuid.NewString())
	implicitConfig.Digests = Digests{
		structDigestUID: Digest{
			UID:                 structDigestUID,
			Name:                "struct",
			StreamUID:           streamUID,
			FlushPeriod:         time.Minute,
			ComputationLocation: ComputationLocationCollector,
			Type:                DigestTypeSt,
			St: &DigestSt{
				MaxProcessedFields: 100,
			},
		},
		valueDigestUID: Digest{
			UID:                 valueDigestUID,
			Name:                "value",
			StreamUID:           streamUID,
			FlushPeriod:         time.Minute,
			ComputationLocation: ComputationLocationCollector,
			Type:                DigestTypeValue,
			Value: &DigestValue{
				MaxProcessedFields: 100,
			},
		},
	}

	return implicitConfig
}

func NewSamplerConfigFromProto(config *protos.SamplerConfig) SamplerConfig {
	if config == nil {
		return SamplerConfig{}
	}

	var streams map[SamplerStreamUID]Stream
	if len(config.GetStreams()) > 0 {
		streams = make(map[SamplerStreamUID]Stream)
	}

	for _, protoSR := range config.GetStreams() {
		streams[SamplerStreamUID(protoSR.GetUid())] = NewStreamFromProto(protoSR)
	}
	var limiterIn *LimiterConfig
	if config.LimiterIn != nil {
		p := NewLimiterFromProto(config.GetLimiterIn())
		limiterIn = &p
	}

	var samplingIn *SamplingConfig
	if config.SamplingIn != nil {
		p := NewSamplingConfigFromProto(config.GetSamplingIn())
		samplingIn = &p
	}

	var limiterOut *LimiterConfig
	if config.LimiterOut != nil {
		p := NewLimiterFromProto(config.GetLimiterOut())
		limiterOut = &p
	}

	var digests map[SamplerDigestUID]Digest
	if len(config.GetDigests()) > 0 {
		digests = make(map[SamplerDigestUID]Digest)
	}
	for _, protoDigest := range config.GetDigests() {
		digests[SamplerDigestUID(protoDigest.GetUid())] = NewDigestFromProto(protoDigest)
	}

	var events map[SamplerEventUID]Event
	if len(config.GetEvents()) > 0 {
		events = make(map[SamplerEventUID]Event)
	}
	for _, protoEvent := range config.GetEvents() {
		events[SamplerEventUID(protoEvent.GetUid())] = NewEventFromProto(protoEvent)
	}

	return SamplerConfig{
		Streams:    streams,
		LimiterIn:  limiterIn,
		SamplingIn: samplingIn,
		LimiterOut: limiterOut,
		Digests:    digests,
		Events:     events,
	}
}

func (pc SamplerConfig) IsEmpty() bool {
	return (len(pc.Streams) == 0 &&
		pc.LimiterIn == nil &&
		pc.SamplingIn == nil &&
		pc.LimiterOut == nil &&
		len(pc.Digests) == 0) &&
		len(pc.Events) == 0
}

func (pc *SamplerConfig) DigestTypesByLocation(location ComputationLocation) []DigestType {

	digestTypesMap := make(map[DigestType]bool)
	for _, valueDigest := range pc.Digests {
		if valueDigest.ComputationLocation == location {
			digestTypesMap[valueDigest.Type] = true
		}
	}

	digestTypes := make([]DigestType, 0, len(digestTypesMap))
	for digestType := range digestTypesMap {
		digestTypes = append(digestTypes, digestType)
	}

	return digestTypes
}

func (pc *SamplerConfig) Merge(update SamplerConfigUpdate) {
	// Update Streams
	if update.Reset.Streams || pc.Streams == nil {
		pc.Streams = make(map[SamplerStreamUID]Stream)
	}
	for _, rule := range update.StreamUpdates {
		switch rule.Op {
		case StreamUpsert:
			pc.Streams[rule.Stream.UID] = rule.Stream
		case StreamDelete:
			delete(pc.Streams, rule.Stream.UID)
		default:
		}
	}

	// Update LimiterIn
	if update.Reset.LimiterIn {
		pc.LimiterIn = nil
	}
	if update.LimiterIn != nil {
		pc.LimiterIn = update.LimiterIn
	}

	// Update SamplingIn
	if update.Reset.SamplingIn {
		pc.SamplingIn = nil
	}
	if update.SamplingIn != nil {
		pc.SamplingIn = update.SamplingIn
	}

	// Update LimiterOut
	if update.Reset.LimiterOut {
		pc.LimiterOut = nil
	}
	if update.LimiterOut != nil {
		pc.LimiterOut = update.LimiterOut
	}

	// Update Digests
	if update.Reset.Digests || pc.Digests == nil {
		pc.Digests = make(map[SamplerDigestUID]Digest)
	}
	for _, rule := range update.DigestUpdates {
		switch rule.Op {
		case DigestUpsert:
			pc.Digests[rule.Digest.UID] = rule.Digest
		case DigestDelete:
			delete(pc.Digests, rule.Digest.UID)
		default:
		}
	}

	// Update Events
	if update.Reset.Events || pc.Events == nil {
		pc.Events = make(map[SamplerEventUID]Event)
	}
	for _, update := range update.EventUpdates {
		switch update.Op {
		case EventUpsert:
			// Update config
			pc.Events[update.Event.UID] = update.Event

		case EventDelete:
			// Delete from config
			delete(pc.Events, update.Event.UID)

		default:
		}
	}
}

func (pc SamplerConfig) ToProto() *protos.SamplerConfig {
	var protoStreams []*protos.Stream
	for _, stream := range pc.Streams {
		protoStreams = append(protoStreams, stream.ToProto())
	}

	var protoLimiterIn *protos.Limiter
	if pc.LimiterIn != nil {
		protoLimiterIn = pc.LimiterIn.ToProto()
	}

	var protoSamplingIn *protos.Sampling
	if pc.SamplingIn != nil {
		protoSamplingIn = pc.SamplingIn.ToProto()
	}

	var protoLimiterOut *protos.Limiter
	if pc.LimiterOut != nil {
		protoLimiterOut = pc.LimiterOut.ToProto()
	}

	var protoDigests []*protos.Digest
	for _, digest := range pc.Digests {
		protoDigests = append(protoDigests, digest.ToProto())
	}

	var protoEvents []*protos.Event
	for _, event := range pc.Events {
		protoEvents = append(protoEvents, event.ToProto())
	}

	return &protos.SamplerConfig{
		Streams:    protoStreams,
		LimiterIn:  protoLimiterIn,
		SamplingIn: protoSamplingIn,
		LimiterOut: protoLimiterOut,
		Digests:    protoDigests,
		Events:     protoEvents,
	}
}
