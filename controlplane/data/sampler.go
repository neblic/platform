package data

import (
	"github.com/neblic/platform/controlplane/protos"
)

type StreamRuleLang int

const (
	SrlUnknown StreamRuleLang = iota
	SrlCel
)

func (srl StreamRuleLang) String() string {
	switch srl {
	case SrlCel:
		return "CEL"
	case SrlUnknown:
		fallthrough
	default:
		return "Unknown"
	}
}

func NewStreamRuleLangFromProto(lang protos.StreamRule_Language) StreamRuleLang {
	switch lang {
	case protos.StreamRule_UNKNOWN:
		return SrlUnknown
	case protos.StreamRule_CEL:
		return SrlCel
	default:
		return SrlUnknown
	}
}

func (srl StreamRuleLang) ToProto() protos.StreamRule_Language {
	switch srl {
	case SrlCel:
		return protos.StreamRule_CEL
	case SrlUnknown:
		fallthrough
	default:
		return protos.StreamRule_UNKNOWN
	}
}

type SamplerStreamRuleUID string

type StreamRule struct {
	UID  SamplerStreamRuleUID
	Lang StreamRuleLang
	Rule string
}

func NewStreamRuleFromProto(sr *protos.StreamRule) StreamRule {
	if sr == nil {
		return StreamRule{}
	}

	return StreamRule{
		UID:  SamplerStreamRuleUID(sr.GetUid()),
		Lang: NewStreamRuleLangFromProto(sr.GetLanguage()),
		Rule: sr.Rule,
	}
}

func (sr StreamRule) ToProto() *protos.StreamRule {
	return &protos.StreamRule{
		Uid:      string(sr.UID),
		Language: sr.Lang.ToProto(),
		Rule:     sr.Rule,
	}
}

type SamplingRate struct {
	Limit int64
	Burst int64
}

func NewSamplingRateFromProto(sr *protos.SamplingRate) SamplingRate {
	if sr == nil {
		return SamplingRate{}
	}

	return SamplingRate{
		Limit: sr.Limit,
		Burst: sr.Burst,
	}
}

func (sr SamplingRate) ToProto() *protos.SamplingRate {
	return &protos.SamplingRate{
		Limit: sr.Limit,
		Burst: sr.Burst,
	}
}

type SamplerStreamUID string

type Stream struct {
	UID        SamplerStreamUID
	StreamRule StreamRule
}

func NewStreamFromProto(s *protos.Stream) Stream {
	if s == nil {
		return Stream{}
	}

	return Stream{
		UID:        SamplerStreamUID(s.GetUid()),
		StreamRule: NewStreamRuleFromProto(s.GetRule()),
	}
}

func (s Stream) ToProto() *protos.Stream {
	return &protos.Stream{
		Uid:  string(s.UID),
		Rule: s.StreamRule.ToProto(),
	}
}

type StreamRuleUpdateOp int

const (
	StreamRuleUpsert StreamRuleUpdateOp = iota
	StreamRuleDelete
)

type StreamUpdate struct {
	Op     StreamRuleUpdateOp
	Stream Stream
}

type SamplerConfigUpdate struct {
	SamplingRate  *SamplingRate
	StreamUpdates []StreamUpdate
}

func NewSamplerConfigUpdateFromProto(protoUpdate *protos.ClientSamplerConfigUpdate) SamplerConfigUpdate {
	if protoUpdate == nil {
		return SamplerConfigUpdate{}
	}

	var sr *SamplingRate
	if protoUpdate.GetSamplingRate() != nil {
		newSr := NewSamplingRateFromProto(protoUpdate.GetSamplingRate())
		sr = &newSr
	}

	var streamUpdates []StreamUpdate
	for _, rule := range protoUpdate.GetStreamUpdates() {
		var op StreamRuleUpdateOp
		switch rule.GetOp() {
		case protos.ClientStreamUpdate_UPSERT:
			op = StreamRuleUpsert
		case protos.ClientStreamUpdate_DELETE:
			op = StreamRuleDelete
		default:
		}

		streamUpdates = append(streamUpdates, StreamUpdate{
			Op:     op,
			Stream: NewStreamFromProto(rule.GetStream()),
		})
	}

	return SamplerConfigUpdate{
		SamplingRate:  sr,
		StreamUpdates: streamUpdates,
	}
}

func (pcu SamplerConfigUpdate) ToProto() *protos.ClientSamplerConfigUpdate {
	var protoSamplingRate *protos.SamplingRate
	if pcu.SamplingRate != nil {
		protoSamplingRate = pcu.SamplingRate.ToProto()
	}

	var protoUpdateStreams []*protos.ClientStreamUpdate
	for _, ruleUpdate := range pcu.StreamUpdates {
		protoOp := protos.ClientStreamUpdate_UNKNOWN
		switch ruleUpdate.Op {
		case StreamRuleUpsert:
			protoOp = protos.ClientStreamUpdate_UPSERT
		case StreamRuleDelete:
			protoOp = protos.ClientStreamUpdate_DELETE
		}

		protoUpdateStreams = append(protoUpdateStreams, &protos.ClientStreamUpdate{
			Op:     protoOp,
			Stream: ruleUpdate.Stream.ToProto(),
		})
	}

	return &protos.ClientSamplerConfigUpdate{
		SamplingRate:  protoSamplingRate,
		StreamUpdates: protoUpdateStreams,
	}
}

// The fields are pointers because the sampler uses this type to update its config.
// When a field is not set, it is considered to not have changed.
//
// This struct gets serialized to disk to persist sampler configurations,
// all fields need to be uppercase to be properly persisted.
type SamplerConfig struct {
	Streams      map[SamplerStreamUID]Stream
	SamplingRate *SamplingRate
}

func NewSamplerConfig() *SamplerConfig {
	return &SamplerConfig{
		Streams: make(map[SamplerStreamUID]Stream),
	}
}

func NewSamplerConfigFromProto(config *protos.SamplerConfig) SamplerConfig {
	if config == nil {
		return SamplerConfig{}
	}

	streams := map[SamplerStreamUID]Stream{}
	for _, protoSR := range config.GetStreams() {
		streams[SamplerStreamUID(protoSR.GetUid())] = NewStreamFromProto(protoSR)
	}

	var samplingRate *SamplingRate
	if config.SamplingRate != nil {
		samplingRate = &SamplingRate{
			Limit: config.GetSamplingRate().GetLimit(),
			Burst: config.GetSamplingRate().GetBurst(),
		}
	}

	return SamplerConfig{
		Streams:      streams,
		SamplingRate: samplingRate,
	}
}

func (pc SamplerConfig) IsEmpty() bool {
	return pc.SamplingRate == nil && len(pc.Streams) == 0
}

func (pc SamplerConfig) ToProto() *protos.SamplerConfig {
	var protoStream []*protos.Stream
	for _, stream := range pc.Streams {
		protoStream = append(protoStream, stream.ToProto())
	}

	var protoSamplingRate *protos.SamplingRate
	if pc.SamplingRate != nil {
		protoSamplingRate = &protos.SamplingRate{
			Limit: pc.SamplingRate.Limit,
			Burst: pc.SamplingRate.Burst,
		}
	}

	return &protos.SamplerConfig{
		Streams:      protoStream,
		SamplingRate: protoSamplingRate,
	}
}

type SamplerSamplingStats struct {
	SamplesEvaluated uint64
	SamplesExported  uint64
}

func NewSamplerSamplingStatsFromProto(stats *protos.SamplerSamplingStats) SamplerSamplingStats {
	if stats == nil {
		return SamplerSamplingStats{}
	}

	return SamplerSamplingStats{
		SamplesEvaluated: stats.GetSamplesEvaluated(),
		SamplesExported:  stats.GetSamplesExported(),
	}
}

func (s SamplerSamplingStats) ToProto() *protos.SamplerSamplingStats {
	return &protos.SamplerSamplingStats{
		SamplesEvaluated: s.SamplesEvaluated,
		SamplesExported:  s.SamplesExported,
	}
}

type SamplerUID string

type Sampler struct {
	Name          string
	Resource      string
	UID           SamplerUID
	Config        SamplerConfig
	SamplingStats SamplerSamplingStats
}

func NewSampler(name, resource string, uid SamplerUID) *Sampler {
	return &Sampler{
		Name:     name,
		Resource: resource,
		UID:      uid,
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
		Config:        NewSamplerConfigFromProto(sampler.Config),
		SamplingStats: NewSamplerSamplingStatsFromProto(sampler.GetSamplingStats()),
	}
}

func (p Sampler) ToProto() *protos.Sampler {
	return &protos.Sampler{
		Uid:           string(p.UID),
		Name:          p.Name,
		Resource:      p.Resource,
		Config:        p.Config.ToProto(),
		SamplingStats: p.SamplingStats.ToProto(),
	}
}
