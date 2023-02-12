package data

import (
	"github.com/neblic/platform/controlplane/protos"
)

type SamplingRuleLang int

const (
	SrlUnknown SamplingRuleLang = iota
	SrlCel
)

func (srl SamplingRuleLang) String() string {
	switch srl {
	case SrlCel:
		return "CEL"
	case SrlUnknown:
		fallthrough
	default:
		return "Unknown"
	}
}

func NewSamplingRuleLangFromProto(lang protos.SamplingRule_Language) SamplingRuleLang {
	switch lang {
	case protos.SamplingRule_UNKNOWN:
		return SrlUnknown
	case protos.SamplingRule_CEL:
		return SrlCel
	default:
		return SrlUnknown
	}
}

func (srl SamplingRuleLang) ToProto() protos.SamplingRule_Language {
	switch srl {
	case SrlCel:
		return protos.SamplingRule_CEL
	case SrlUnknown:
		fallthrough
	default:
		return protos.SamplingRule_UNKNOWN
	}
}

type SamplerSamplingRuleUID string
type SamplerUID string

type SamplingRule struct {
	UID  SamplerSamplingRuleUID
	Lang SamplingRuleLang
	Rule string
}

func NewSamplingRuleFromProto(sr *protos.SamplingRule) SamplingRule {
	if sr == nil {
		return SamplingRule{}
	}

	return SamplingRule{
		UID:  SamplerSamplingRuleUID(sr.GetUid()),
		Lang: NewSamplingRuleLangFromProto(sr.GetLanguage()),
		Rule: sr.Rule,
	}
}

func (sr SamplingRule) ToProto() *protos.SamplingRule {
	return &protos.SamplingRule{
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

type SamplingRuleUpdateOp int

const (
	SamplingRuleUpsert SamplingRuleUpdateOp = iota
	SamplingRuleDelete
)

type SamplingRuleUpdate struct {
	Op           SamplingRuleUpdateOp
	SamplingRule SamplingRule
}

type SamplerConfigUpdate struct {
	SamplingRate        *SamplingRate
	SamplingRuleUpdates []SamplingRuleUpdate
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

	var samplingRuleUpdates []SamplingRuleUpdate
	for _, rule := range protoUpdate.GetSamplingRuleUpdates() {
		var op SamplingRuleUpdateOp
		switch rule.GetOp() {
		case protos.ClientSamplingRuleUpdate_UPSERT:
			op = SamplingRuleUpsert
		case protos.ClientSamplingRuleUpdate_DELETE:
			op = SamplingRuleDelete
		default:
		}

		samplingRuleUpdates = append(samplingRuleUpdates, SamplingRuleUpdate{
			Op:           op,
			SamplingRule: NewSamplingRuleFromProto(rule.GetSamplingRule()),
		})
	}

	return SamplerConfigUpdate{
		SamplingRate:        sr,
		SamplingRuleUpdates: samplingRuleUpdates,
	}
}

func (pcu SamplerConfigUpdate) ToProto() *protos.ClientSamplerConfigUpdate {
	var protoSamplingRate *protos.SamplingRate
	if pcu.SamplingRate != nil {
		protoSamplingRate = pcu.SamplingRate.ToProto()
	}

	var protoUpdateSamplingRules []*protos.ClientSamplingRuleUpdate
	for _, ruleUpdate := range pcu.SamplingRuleUpdates {
		protoOp := protos.ClientSamplingRuleUpdate_UNKNOWN
		switch ruleUpdate.Op {
		case SamplingRuleUpsert:
			protoOp = protos.ClientSamplingRuleUpdate_UPSERT
		case SamplingRuleDelete:
			protoOp = protos.ClientSamplingRuleUpdate_DELETE
		}

		protoUpdateSamplingRules = append(protoUpdateSamplingRules, &protos.ClientSamplingRuleUpdate{
			Op:           protoOp,
			SamplingRule: ruleUpdate.SamplingRule.ToProto(),
		})
	}

	return &protos.ClientSamplerConfigUpdate{
		SamplingRate:        protoSamplingRate,
		SamplingRuleUpdates: protoUpdateSamplingRules,
	}
}

// The fields are pointers because the sampler uses this type to update its config.
// When a field is not set, it is considered to not have changed.
//
// This struct gets serialized to disk to persist sampler configurations,
// all fields need to be uppercase to be properly persisted.
type SamplerConfig struct {
	SamplingRules map[SamplerSamplingRuleUID]SamplingRule
	SamplingRate  *SamplingRate
}

func NewSamplerConfig() *SamplerConfig {
	return &SamplerConfig{
		SamplingRules: make(map[SamplerSamplingRuleUID]SamplingRule),
	}
}

func NewSamplerConfigFromProto(config *protos.SamplerConfig) SamplerConfig {
	if config == nil {
		return SamplerConfig{}
	}

	samplingRules := map[SamplerSamplingRuleUID]SamplingRule{}
	for _, protoSR := range config.GetSamplingRules() {
		samplingRules[SamplerSamplingRuleUID(protoSR.GetUid())] = NewSamplingRuleFromProto(protoSR)
	}

	var samplingRate *SamplingRate
	if config.SamplingRate != nil {
		samplingRate = &SamplingRate{
			Limit: config.GetSamplingRate().GetLimit(),
			Burst: config.GetSamplingRate().GetBurst(),
		}
	}

	return SamplerConfig{
		SamplingRules: samplingRules,
		SamplingRate:  samplingRate,
	}
}

func (pc SamplerConfig) IsEmpty() bool {
	return pc.SamplingRate == nil && len(pc.SamplingRules) == 0
}

func (pc SamplerConfig) ToProto() *protos.SamplerConfig {
	var protoSRules []*protos.SamplingRule
	for _, samplingRule := range pc.SamplingRules {
		protoSRules = append(protoSRules, samplingRule.ToProto())
	}

	var protoSamplingRate *protos.SamplingRate
	if pc.SamplingRate != nil {
		protoSamplingRate = &protos.SamplingRate{
			Limit: pc.SamplingRate.Limit,
			Burst: pc.SamplingRate.Burst,
		}
	}

	return &protos.SamplerConfig{
		SamplingRules: protoSRules,
		SamplingRate:  protoSamplingRate,
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
