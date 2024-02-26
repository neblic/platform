package control

import (
	"regexp"

	"github.com/neblic/platform/controlplane/protos"
	"gopkg.in/yaml.v3"
)

var nameValidationRegex = regexp.MustCompile(`^[\.\/()\w-_]*$`)
var nameValidationErrTemplate = "invalid %s name %s, expected alphanumerical with ./()-_ characters"

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

type CollectorStats struct {
	SamplesCollected uint64
}

func NewCollectorStatsFromProto(stats *protos.Sampler_CollectorStats) CollectorStats {
	if stats == nil {
		return CollectorStats{}
	}

	return CollectorStats{
		SamplesCollected: stats.GetSamplesCollected(),
	}
}

func (s *CollectorStats) Add(SamplesCollected uint64) {
	s.SamplesCollected += SamplesCollected
}

func (s CollectorStats) ToProto() *protos.Sampler_CollectorStats {
	return &protos.Sampler_CollectorStats{
		SamplesCollected: s.SamplesCollected,
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
	UID            SamplerUID
	Resource       string
	Name           string
	Tags           Tags
	Capabilities   Capabilities
	Config         SamplerConfig
	SamplingStats  SamplerSamplingStats
	CollectorStats CollectorStats
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
		UID:            SamplerUID(sampler.GetUid()),
		Resource:       sampler.GetResource(),
		Name:           sampler.GetName(),
		Tags:           NewTagsFromProto(sampler.GetTags()),
		Capabilities:   NewCapabilitiesFromProto(sampler.GetCapabilities()),
		Config:         NewSamplerConfigFromProto(sampler.Config),
		SamplingStats:  NewSamplerSamplingStatsFromProto(sampler.GetSamplingStats()),
		CollectorStats: NewCollectorStatsFromProto(sampler.GetCollectorStats()),
	}
}

func (p Sampler) ToProto() *protos.Sampler {
	return &protos.Sampler{
		Uid:            string(p.UID),
		Name:           p.Name,
		Resource:       p.Resource,
		Tags:           p.Tags.ToProto(),
		Capabilities:   p.Capabilities.ToProto(),
		Config:         p.Config.ToProto(),
		SamplingStats:  p.SamplingStats.ToProto(),
		CollectorStats: p.CollectorStats.ToProto(),
	}
}
