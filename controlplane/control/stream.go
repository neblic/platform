package control

import (
	"github.com/neblic/platform/controlplane/protos"
)

type SamplerStreamUID string

type Stream struct {
	UID              SamplerStreamUID
	Name             string
	StreamRule       Rule
	ExportRawSamples bool
}

func (s Stream) GetUID() SamplerStreamUID {
	return s.UID
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
	}
}

func (s Stream) ToProto() *protos.Stream {
	return &protos.Stream{
		Uid:              string(s.UID),
		Name:             s.Name,
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
