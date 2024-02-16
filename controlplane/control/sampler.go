package control

import (
	"errors"
	"regexp"

	"github.com/neblic/platform/controlplane/protos"
	"gopkg.in/yaml.v3"
)

var nameValidationRegex = regexp.MustCompile(`^[\.\/()\w-_]*$`)
var nameValidationErrTemplate = "invalid %s name %s, expected alphanumerical with ./()-_ characters"

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

type Streams map[SamplerStreamUID]Stream

// MarshalYAML implements the yaml.Marshaler interface. Streams contains a map, but the data is
// marshaled into a list.
func (s Streams) MarshalYAML() (interface{}, error) {
	streams := make([]Stream, 0, len(s))
	for _, stream := range s {
		streams = append(streams, stream)
	}

	node := yaml.Node{}
	err := node.Encode(streams)
	if err != nil {
		return nil, err
	}

	return node, err
}

// UnmarshalYAML implements the yaml.Unmarshaler interface. Streams contains a list, but the data is
// unmarshaled into a map.
func (s *Streams) UnmarshalYAML(node *yaml.Node) error {
	streams := []Stream{}
	err := node.Decode(&streams)
	if err != nil {
		return err
	}

	*s = map[SamplerStreamUID]Stream{}
	for _, stream := range streams {
		(*s)[stream.UID] = stream
	}

	return nil
}

type Digests map[SamplerDigestUID]Digest

// MarshalYAML implements the yaml.Marshaler interface. Digests contains a map, but the data is
// marshaled into a list.
func (d Digests) MarshalYAML() (interface{}, error) {
	digests := make([]Digest, 0, len(d))
	for _, digest := range d {
		digests = append(digests, digest)
	}

	node := yaml.Node{}
	err := node.Encode(digests)
	if err != nil {
		return nil, err
	}

	return node, err
}

// UnmarshalYAML implements the yaml.Unmarshaler interface. Digests contains a list, but the data is
// unmarshaled into a map.
func (d *Digests) UnmarshalYAML(node *yaml.Node) error {
	digests := []Digest{}
	err := node.Decode(&digests)
	if err != nil {
		return err
	}

	*d = map[SamplerDigestUID]Digest{}
	for _, digest := range digests {
		(*d)[digest.UID] = digest
	}

	return nil
}

type Events map[SamplerEventUID]Event

// MarshalYAML implements the yaml.Marshaler interface. Events contains a map, but the data is
// marshaled into a list.
func (d Events) MarshalYAML() (interface{}, error) {
	events := make([]Event, 0, len(d))
	for _, event := range d {
		events = append(events, event)
	}

	node := yaml.Node{}
	err := node.Encode(events)
	if err != nil {
		return nil, err
	}

	return node, err
}

// UnmarshalYAML implements the yaml.Unmarshaler interface. Events contains a list, but the data is
// unmarshaled into a map.
func (d *Events) UnmarshalYAML(node *yaml.Node) error {
	events := []Event{}
	err := node.Decode(&events)
	if err != nil {
		return err
	}

	*d = map[SamplerEventUID]Event{}
	for _, event := range events {
		(*d)[event.UID] = event
	}

	return nil
}

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

type SamplerSamplingStats struct {
	SamplesEvaluated uint64
	SamplesExported  uint64
	SamplesDigested  uint64
}

func NewSamplerSamplingStatsFromProto(stats *protos.SamplerSamplingStats) SamplerSamplingStats {
	if stats == nil {
		return SamplerSamplingStats{}
	}

	return SamplerSamplingStats{
		SamplesEvaluated: stats.GetSamplesEvaluated(),
		SamplesExported:  stats.GetSamplesExported(),
		SamplesDigested:  stats.GetSamplesDigested(),
	}
}

func (s SamplerSamplingStats) ToProto() *protos.SamplerSamplingStats {
	return &protos.SamplerSamplingStats{
		SamplesEvaluated: s.SamplesEvaluated,
		SamplesExported:  s.SamplesExported,
		SamplesDigested:  s.SamplesDigested,
	}
}

type SamplerUID string

// if updated, remember to update the exported tags in the public sampler API
const (
	ProducerTag = "producer"
	ConsumerTag = "consumer"
	RequestTag  = "request"
	ResponseTag = "response"
	DLQTag      = "dlq"
)

type Tag struct {
	Name  string
	Attrs map[string]string
}

func NewTagFromProto(tag *protos.Sampler_Tag) Tag {
	if tag == nil {
		return Tag{}
	}

	return Tag{
		Name:  tag.GetName(),
		Attrs: tag.GetAttrs(),
	}
}

func (t Tag) ToProto() *protos.Sampler_Tag {
	return &protos.Sampler_Tag{
		Name:  t.Name,
		Attrs: t.Attrs,
	}
}

type Tags []Tag

func NewTagsFromProto(protoTags []*protos.Sampler_Tag) Tags {
	if protoTags == nil {
		return nil
	}

	tags := []Tag{}
	for _, tag := range protoTags {
		tags = append(tags, NewTagFromProto(tag))
	}

	return tags
}

func (t Tags) ToProto() []*protos.Sampler_Tag {
	protoTags := []*protos.Sampler_Tag{}
	for _, tag := range t {
		protoTags = append(protoTags, tag.ToProto())
	}

	return protoTags
}

type Sampler struct {
	UID           SamplerUID
	Resource      string
	Name          string
	Tags          Tags
	Capabilities  Capabilities
	Config        SamplerConfig
	SamplingStats SamplerSamplingStats
}

func NewSampler(name, resource string, uid SamplerUID) *Sampler {
	return &Sampler{
		UID:      uid,
		Resource: resource,
		Name:     name,
	}
}

func NewSamplerFromProto(sampler *protos.Sampler) *Sampler {
	if sampler == nil {
		return nil
	}

	return &Sampler{
		UID:           SamplerUID(sampler.GetUid()),
		Resource:      sampler.GetResource(),
		Name:          sampler.GetName(),
		Tags:          NewTagsFromProto(sampler.GetTags()),
		Capabilities:  NewCapabilitiesFromProto(sampler.GetCapabilities()),
		Config:        NewSamplerConfigFromProto(sampler.Config),
		SamplingStats: NewSamplerSamplingStatsFromProto(sampler.GetSamplingStats()),
	}
}

func (p Sampler) ToProto() *protos.Sampler {
	return &protos.Sampler{
		Uid:           string(p.UID),
		Name:          p.Name,
		Resource:      p.Resource,
		Tags:          p.Tags.ToProto(),
		Capabilities:  p.Capabilities.ToProto(),
		Config:        p.Config.ToProto(),
		SamplingStats: p.SamplingStats.ToProto(),
	}
}
