package control

import (
	"fmt"
	"time"

	"github.com/neblic/platform/controlplane/protos"
	"google.golang.org/protobuf/types/known/durationpb"
)

type SamplerStreamUID string

type Keyed struct {
	TTL     time.Duration
	MaxKeys int32
}

func NewKeyedFromProto(k *protos.Stream_Keyed) *Keyed {
	if k == nil {
		return nil
	}

	return &Keyed{
		TTL:     k.GetTtl().AsDuration(),
		MaxKeys: k.GetMaxKeys(),
	}
}

func (k *Keyed) ToProto() *protos.Stream_Keyed {
	if k == nil {
		return nil
	}

	return &protos.Stream_Keyed{
		Ttl:     durationpb.New(k.TTL),
		MaxKeys: k.MaxKeys,
	}
}

type Stream struct {
	UID              SamplerStreamUID
	Name             string
	StreamRule       Rule
	ExportRawSamples bool
	Keyed            *Keyed
}

func (s Stream) GetName() string {
	return s.Name
}

func NewStreamFromProto(s *protos.Stream) Stream {
	if s == nil {
		return Stream{}
	}

	return Stream{
		UID:              SamplerStreamUID(s.GetUid()),
		Name:             s.Name,
		StreamRule:       NewRuleFromProto(s.GetRule()),
		ExportRawSamples: s.ExportRawSamples,
		Keyed:            NewKeyedFromProto(s.GetKeyed()),
	}
}

func (s Stream) ToProto() *protos.Stream {
	return &protos.Stream{
		Uid:              string(s.UID),
		Name:             s.Name,
		Rule:             s.StreamRule.ToProto(),
		ExportRawSamples: s.ExportRawSamples,
		Keyed:            s.Keyed.ToProto(),
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

func (su *StreamUpdate) IsValid() error {
	isValid := nameValidationRegex.MatchString(string(su.Stream.Name))
	if !isValid {
		return fmt.Errorf(nameValidationErrTemplate, "stream", su.Stream.Name)
	}
	return nil
}

func (su *StreamUpdate) ToProto() *protos.ClientStreamUpdate {
	protoOp := protos.ClientStreamUpdate_UNKNOWN
	switch su.Op {
	case StreamUpsert:
		protoOp = protos.ClientStreamUpdate_UPSERT
	case StreamDelete:
		protoOp = protos.ClientStreamUpdate_DELETE
	}

	return &protos.ClientStreamUpdate{
		Op:     protoOp,
		Stream: su.Stream.ToProto(),
	}
}
