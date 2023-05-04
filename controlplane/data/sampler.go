package data

import (
	"fmt"

	"github.com/neblic/platform/controlplane/protos"
)

type Stream_RuleLang int

const (
	SrlCel = iota + 1
)

func (srl Stream_RuleLang) String() string {
	switch srl {
	case SrlCel:
		return "CEL"
	default:
		return "Unknown"
	}
}

func NewStreamRuleLangFromProto(lang protos.Stream_Rule_Language) Stream_RuleLang {
	var srl Stream_RuleLang
	switch lang {
	case protos.Stream_Rule_CEL:
		srl = SrlCel
	}

	return srl
}

func (srl Stream_RuleLang) ToProto() protos.Stream_Rule_Language {
	switch srl {
	case SrlCel:
		return protos.Stream_Rule_CEL
	default:
		return protos.Stream_Rule_UNKNOWN
	}
}

type SamplerStreamRuleUID string

type StreamRule struct {
	Lang Stream_RuleLang
	Rule string
}

func NewStreamRuleFromProto(sr *protos.Stream_Rule) StreamRule {
	if sr == nil {
		return StreamRule{}
	}

	return StreamRule{
		Lang: NewStreamRuleLangFromProto(sr.GetLanguage()),
		Rule: sr.Rule,
	}
}

func (sr StreamRule) ToProto() *protos.Stream_Rule {
	return &protos.Stream_Rule{
		Language: sr.Lang.ToProto(),
		Rule:     sr.Rule,
	}
}

type LimiterConfig struct {
	Limit int64
}

func NewLimiterFromProto(sr *protos.Limiter) LimiterConfig {
	if sr == nil {
		return LimiterConfig{}
	}

	return LimiterConfig{
		Limit: sr.Limit,
	}
}

func (sr LimiterConfig) ToProto() *protos.Limiter {
	return &protos.Limiter{
		Limit: sr.Limit,
	}
}

type SamplingType int

const (
	UnknownSamplingType SamplingType = iota
	DeterministicSamplingType
)

type DeterministicSamplingConfig struct {
	SampleRate             int32
	SampleEmptyDeterminant bool
}

type SamplingConfig struct {
	SamplingType          SamplingType
	DeterministicSampling DeterministicSamplingConfig
}

func (sc SamplingConfig) String() string {
	switch sc.SamplingType {
	case DeterministicSamplingType:
		return fmt.Sprintf("DeterministicSampling(SampleRate: %d, SampleEmptyDeterminant: %t)", sc.DeterministicSampling.SampleRate, sc.DeterministicSampling.SampleEmptyDeterminant)
	default:
		return "Unknown"
	}
}

func NewSamplingConfigFromProto(sr *protos.Sampling) SamplingConfig {
	if sr == nil {
		return SamplingConfig{}
	}

	var samplingType SamplingType
	switch sr.GetSampling().(type) {
	case *protos.Sampling_DeterministicSampling:
		samplingType = DeterministicSamplingType
		return SamplingConfig{
			SamplingType: samplingType,
			DeterministicSampling: DeterministicSamplingConfig{
				SampleRate: sr.GetDeterministicSampling().GetSampleRate(),
			},
		}
	default:
		return SamplingConfig{}
	}
}

func (sr SamplingConfig) ToProto() *protos.Sampling {
	switch sr.SamplingType {
	case DeterministicSamplingType:
		return &protos.Sampling{
			Sampling: &protos.Sampling_DeterministicSampling{
				DeterministicSampling: &protos.DeterministicSampling{
					SampleRate: sr.DeterministicSampling.SampleRate,
				},
			},
		}
	case UnknownSamplingType:
		return &protos.Sampling{}
	default:
		return nil
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
	StreamRuleUpsert StreamRuleUpdateOp = iota + 1
	StreamRuleDelete
)

type StreamUpdate struct {
	Op     StreamRuleUpdateOp
	Stream Stream
}

type SamplerConfigUpdateReset struct {
	LimiterIn  bool
	SamplingIn bool
	Streams    bool
	LimiterOut bool
}

func NewSamplerConfigUpdateResetFromProto(protoReset *protos.ClientSamplerConfigUpdate_Reset) SamplerConfigUpdateReset {
	if protoReset == nil {
		return SamplerConfigUpdateReset{}
	}

	return SamplerConfigUpdateReset{
		LimiterIn:  protoReset.GetLimiterIn(),
		SamplingIn: protoReset.GetSamplingIn(),
		Streams:    protoReset.GetStreams(),
		LimiterOut: protoReset.GetLimiterOut(),
	}
}

func (scr SamplerConfigUpdateReset) ToProto() *protos.ClientSamplerConfigUpdate_Reset {
	return &protos.ClientSamplerConfigUpdate_Reset{
		LimiterIn:  scr.LimiterIn,
		SamplingIn: scr.SamplingIn,
		Streams:    scr.Streams,
		LimiterOut: scr.LimiterOut,
	}
}

type SamplerConfigUpdate struct {
	// If a field is set to true, it means that the field is reset to its default.
	// If a configuration option is reset and set in the same request, it will be
	// first resetted and then set to its new value.
	Reset SamplerConfigUpdateReset

	// All fields are optional. If a field is nil, it means that the field is not updated.
	LimiterIn     *LimiterConfig
	SamplingIn    *SamplingConfig
	StreamUpdates []StreamUpdate
	LimiterOut    *LimiterConfig
}

func NewSamplerConfigUpdateFromProto(protoUpdate *protos.ClientSamplerConfigUpdate) SamplerConfigUpdate {
	if protoUpdate == nil {
		return SamplerConfigUpdate{}
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
	for _, rule := range protoUpdate.GetStreamUpdates() {
		var op StreamRuleUpdateOp
		switch rule.GetOp() {
		case protos.ClientStreamUpdate_UPSERT:
			op = StreamRuleUpsert
		case protos.ClientStreamUpdate_DELETE:
			op = StreamRuleDelete
		}

		streamUpdates = append(streamUpdates, StreamUpdate{
			Op:     op,
			Stream: NewStreamFromProto(rule.GetStream()),
		})
	}

	var limiterOut *LimiterConfig
	if protoUpdate.GetLimiterOut() != nil {
		newSrOut := NewLimiterFromProto(protoUpdate.GetLimiterOut())
		limiterOut = &newSrOut
	}

	return SamplerConfigUpdate{
		Reset: NewSamplerConfigUpdateResetFromProto(protoUpdate.GetReset_()),

		LimiterIn:     limiterIn,
		SamplingIn:    samplingConfigIn,
		StreamUpdates: streamUpdates,
		LimiterOut:    limiterOut,
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
	for _, ruleUpdate := range scu.StreamUpdates {
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

	var protoLimiterOut *protos.Limiter
	if scu.LimiterOut != nil {
		protoLimiterOut = scu.LimiterOut.ToProto()
	}

	return &protos.ClientSamplerConfigUpdate{
		Reset_: scu.Reset.ToProto(),

		LimiterIn:     protoLimiterIn,
		SamplingIn:    protoSamplingIn,
		StreamUpdates: protoUpdateStreams,
		LimiterOut:    protoLimiterOut,
	}
}

// Used to get and update the sampler configuration.
//
// When sent by the server to update a sampler, only the fields that are present
// are updated. If a field is present, the previous value is replaced with the
// new one.
type SamplerConfig struct {
	Streams    map[SamplerStreamUID]Stream
	LimiterIn  *LimiterConfig
	SamplingIn *SamplingConfig
	LimiterOut *LimiterConfig
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

	streams := map[SamplerStreamUID]Stream{}
	for _, protoSR := range config.GetStreams() {
		streams[SamplerStreamUID(protoSR.GetUid())] = NewStreamFromProto(protoSR)
	}

	var limiterOut *LimiterConfig
	if config.LimiterOut != nil {
		p := NewLimiterFromProto(config.GetLimiterOut())
		limiterOut = &p
	}

	return SamplerConfig{
		LimiterIn:  limiterIn,
		SamplingIn: samplingIn,
		Streams:    streams,
		LimiterOut: limiterOut,
	}
}

func (pc SamplerConfig) IsEmpty() bool {
	return pc.LimiterIn == nil && pc.SamplingIn == nil && len(pc.Streams) == 0 && pc.LimiterOut == nil
}

func (pc SamplerConfig) ToProto() *protos.SamplerConfig {
	var protoLimiterIn *protos.Limiter
	if pc.LimiterIn != nil {
		protoLimiterIn = pc.LimiterIn.ToProto()
	}

	var protoSamplingIn *protos.Sampling
	if pc.SamplingIn != nil {
		protoSamplingIn = pc.SamplingIn.ToProto()
	}

	var protoStream []*protos.Stream
	for _, stream := range pc.Streams {
		protoStream = append(protoStream, stream.ToProto())
	}

	var protoLimiterOut *protos.Limiter
	if pc.LimiterOut != nil {
		protoLimiterOut = pc.LimiterOut.ToProto()
	}

	return &protos.SamplerConfig{
		LimiterIn:  protoLimiterIn,
		SamplingIn: protoSamplingIn,
		Streams:    protoStream,
		LimiterOut: protoLimiterOut,
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
