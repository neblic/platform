package control

import (
	"fmt"
	"time"

	"github.com/neblic/platform/controlplane/protos"
	"google.golang.org/protobuf/types/known/durationpb"
)

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

func (ds *DigestSt) ToProto() *protos.Digest_St {
	return &protos.Digest_St{
		MaxProcessedFields: int32(ds.MaxProcessedFields),
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
