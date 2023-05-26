package data

import (
	"fmt"
	"time"

	"github.com/neblic/platform/controlplane/protos"
	"google.golang.org/protobuf/types/known/durationpb"
)

type StreamRuleLang int

const (
	SrlCel = iota + 1
)

func (srl StreamRuleLang) String() string {
	switch srl {
	case SrlCel:
		return "CEL"
	default:
		return "Unknown"
	}
}

func NewStreamRuleLangFromProto(lang protos.Stream_Rule_Language) StreamRuleLang {
	var srl StreamRuleLang
	switch lang {
	case protos.Stream_Rule_CEL:
		srl = SrlCel
	}

	return srl
}

func (srl StreamRuleLang) ToProto() protos.Stream_Rule_Language {
	switch srl {
	case SrlCel:
		return protos.Stream_Rule_CEL
	default:
		return protos.Stream_Rule_UNKNOWN
	}
}

type SamplerStreamRuleUID string

type StreamRule struct {
	Lang StreamRuleLang
	Rule string
}

func (s StreamRule) String() string {
	// Lang intentionally not shown since it is implicit it is always a CEL type for now
	return fmt.Sprintf("%s", s.Rule)
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
	Limit int32
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

func (sc SamplingConfig) CLIInfo() string {
	switch sc.SamplingType {
	case DeterministicSamplingType:
		return fmt.Sprintf("Type: Deterministic, SampleRate: %d, SampleEmptyDeterminant: %t", sc.DeterministicSampling.SampleRate, sc.DeterministicSampling.SampleEmptyDeterminant)
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

func (sc SamplingConfig) ToProto() *protos.Sampling {
	switch sc.SamplingType {
	case DeterministicSamplingType:
		return &protos.Sampling{
			Sampling: &protos.Sampling_DeterministicSampling{
				DeterministicSampling: &protos.DeterministicSampling{
					SampleRate: sc.DeterministicSampling.SampleRate,
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
	UID              SamplerStreamUID
	StreamRule       StreamRule
	ExportRawSamples bool
}

func (s Stream) CLIInfo() string {
	return fmt.Sprintf("UID: %s, Rule: %s", s.UID, s.StreamRule)
}

func NewStreamFromProto(s *protos.Stream) Stream {
	if s == nil {
		return Stream{}
	}

	return Stream{
		UID:              SamplerStreamUID(s.GetUid()),
		StreamRule:       NewStreamRuleFromProto(s.GetRule()),
		ExportRawSamples: s.ExportRawSamples,
	}
}

func (s Stream) ToProto() *protos.Stream {
	return &protos.Stream{
		Uid:              string(s.UID),
		Rule:             s.StreamRule.ToProto(),
		ExportRawSamples: s.ExportRawSamples,
	}
}

type StreamUpdateOp int

const (
	StreamUpsert StreamUpdateOp = iota + 1
	StreamDelete
)

type StreamUpdate struct {
	Op     StreamUpdateOp
	Stream Stream
}

func NewStreamUpdateFromProto(streamUpdate *protos.ClientStreamUpdate) StreamUpdate {
	if streamUpdate == nil {
		return StreamUpdate{}
	}

	var op StreamUpdateOp
	switch streamUpdate.GetOp() {
	case protos.ClientStreamUpdate_UPSERT:
		op = StreamUpsert
	case protos.ClientStreamUpdate_DELETE:
		op = StreamDelete
	}

	return StreamUpdate{
		Op:     op,
		Stream: NewStreamFromProto(streamUpdate.GetStream()),
	}
}

func (s *StreamUpdate) ToProto() *protos.ClientStreamUpdate {
	protoOp := protos.ClientStreamUpdate_UNKNOWN
	switch s.Op {
	case StreamUpsert:
		protoOp = protos.ClientStreamUpdate_UPSERT
	case StreamDelete:
		protoOp = protos.ClientStreamUpdate_DELETE
	}

	return &protos.ClientStreamUpdate{
		Op:     protoOp,
		Stream: s.Stream.ToProto(),
	}
}

type DigestType uint8

const (
	DigestTypeUnknown DigestType = iota
	DigestTypeSt
)

type SamplerDigestUID string

type DigestSt struct {
	MaxProcessedFields int
}

func (ds DigestSt) CLIInfo() string {
	return fmt.Sprintf("MaxProcessedFields: %d", ds.MaxProcessedFields)
}

func NewDigestStFromProto(protoDigestSt *protos.Digest_St) DigestSt {
	if protoDigestSt == nil {
		return DigestSt{}
	}

	return DigestSt{
		MaxProcessedFields: int(protoDigestSt.MaxProcessedFields),
	}
}

func (d *DigestSt) ToProto() *protos.Digest_St {
	return &protos.Digest_St{
		MaxProcessedFields: int32(d.MaxProcessedFields),
	}
}

type Digest struct {
	UID         SamplerDigestUID
	StreamUID   SamplerStreamUID
	FlushPeriod time.Duration
	BufferSize  int

	// digest specific config
	Type DigestType
	St   DigestSt
}

func (d Digest) CLIInfo() string {
	var t string
	switch d.Type {
	case DigestTypeSt:
		t = fmt.Sprintf("Type: Structure, %s", d.St.CLIInfo())
	default:
		t = "Type: Unknown"
	}

	// flush period intentionally not shown given that for now, it is an internal configuration that will configured
	// with a default value by the server
	return fmt.Sprintf("UID: %s, StreamUID: %s, FlushPeriod: %s, %s", d.UID, d.StreamUID, d.FlushPeriod, t)
}

func NewDigestFromProto(protoDigest *protos.Digest) Digest {
	if protoDigest == nil {
		return Digest{}
	}

	digest := Digest{
		UID:         SamplerDigestUID(protoDigest.GetUid()),
		StreamUID:   SamplerStreamUID(protoDigest.GetStreamUid()),
		FlushPeriod: protoDigest.GetFlushPeriod().AsDuration(),
		BufferSize:  int(protoDigest.GetBufferSize()),
	}

	switch t := protoDigest.GetType().(type) {
	case *protos.Digest_St_:
		digest.Type = DigestTypeSt
		digest.St = NewDigestStFromProto(t.St)
	default:
		digest.Type = DigestTypeUnknown
	}

	return digest
}

func (d *Digest) ToProto() *protos.Digest {
	protoDigest := &protos.Digest{
		Uid:         string(d.UID),
		StreamUid:   string(d.StreamUID),
		FlushPeriod: durationpb.New(d.FlushPeriod),
		BufferSize:  int32(d.BufferSize),
	}

	switch d.Type {
	case DigestTypeSt:
		protoDigest.Type = &protos.Digest_St_{
			St: d.St.ToProto(),
		}
	}

	return protoDigest
}

type DigestUpdateOp int

const (
	DigestUpsert DigestUpdateOp = iota + 1
	DigestDelete
)

type DigestUpdate struct {
	Op     DigestUpdateOp
	Digest Digest
}

func NewDigestUpdateFromProto(digestUpdate *protos.ClientDigestUpdate) DigestUpdate {
	if digestUpdate == nil {
		return DigestUpdate{}
	}

	var op DigestUpdateOp
	switch digestUpdate.GetOp() {
	case protos.ClientDigestUpdate_UPSERT:
		op = DigestUpsert
	case protos.ClientDigestUpdate_DELETE:
		op = DigestDelete
	}

	return DigestUpdate{
		Op:     op,
		Digest: NewDigestFromProto(digestUpdate.GetDigest()),
	}
}

func (s *DigestUpdate) ToProto() *protos.ClientDigestUpdate {
	protoOp := protos.ClientDigestUpdate_UNKNOWN
	switch s.Op {
	case DigestUpsert:
		protoOp = protos.ClientDigestUpdate_UPSERT
	case DigestDelete:
		protoOp = protos.ClientDigestUpdate_DELETE
	}

	return &protos.ClientDigestUpdate{
		Op:     protoOp,
		Digest: s.Digest.ToProto(),
	}
}

type SamplerConfigUpdateReset struct {
	LimiterIn  bool
	SamplingIn bool
	Streams    bool
	LimiterOut bool
	Digests    bool
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
	}
}

func (scr SamplerConfigUpdateReset) ToProto() *protos.ClientSamplerConfigUpdate_Reset {
	return &protos.ClientSamplerConfigUpdate_Reset{
		Streams:    scr.Streams,
		LimiterIn:  scr.LimiterIn,
		SamplingIn: scr.SamplingIn,
		LimiterOut: scr.LimiterOut,
		Digests:    scr.Digests,
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

	return SamplerConfigUpdate{
		Reset: NewSamplerConfigUpdateResetFromProto(protoUpdate.GetReset_()),

		LimiterIn:     limiterIn,
		SamplingIn:    samplingConfigIn,
		StreamUpdates: streamUpdates,
		LimiterOut:    limiterOut,
		DigestUpdates: digestUpdates,
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

	return &protos.ClientSamplerConfigUpdate{
		Reset_: scu.Reset.ToProto(),

		StreamUpdates: protoUpdateStreams,
		LimiterIn:     protoLimiterIn,
		SamplingIn:    protoSamplingIn,
		LimiterOut:    protoLimiterOut,
		DigestUpdates: protoUpdateDigests,
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
	Digests    map[SamplerDigestUID]Digest
}

func NewSamplerConfig() *SamplerConfig {
	return &SamplerConfig{
		Streams: make(map[SamplerStreamUID]Stream),
		Digests: make(map[SamplerDigestUID]Digest),
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

	return SamplerConfig{
		Streams:    streams,
		LimiterIn:  limiterIn,
		SamplingIn: samplingIn,
		LimiterOut: limiterOut,
		Digests:    digests,
	}
}

func (pc SamplerConfig) IsEmpty() bool {
	return (len(pc.Streams) == 0 &&
		pc.LimiterIn == nil &&
		pc.SamplingIn == nil &&
		pc.LimiterOut == nil &&
		len(pc.Digests) == 0)
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

	return &protos.SamplerConfig{
		Streams:    protoStreams,
		LimiterIn:  protoLimiterIn,
		SamplingIn: protoSamplingIn,
		LimiterOut: protoLimiterOut,
		Digests:    protoDigests,
	}
}

type SamplerSamplingStats struct {
	SamplesEvaluated uint64
	SamplesExported  uint64
}

func (s SamplerSamplingStats) CLIInfo() string {
	return fmt.Sprintf("Evaluated: %d, Exported: %d", s.SamplesEvaluated, s.SamplesExported)
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
