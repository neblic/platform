package control

import (
	"github.com/neblic/platform/controlplane/protos"
)

type SamplerEventUID string

type SampleType uint8

const (
	UnknownSampleType SampleType = iota
	RawSampleType
	StructDigestSampleType
	ValueDigestSampleType
	EventSampleType
	ConfigSampleType
)

var ValidSampleTypes = []SampleType{UnknownSampleType, RawSampleType, StructDigestSampleType, EventSampleType}

func NewSampleTypeFromProto(sampleType protos.SampleType) SampleType {
	return SampleType(sampleType)
}

func (s SampleType) String() string {
	switch s {
	case UnknownSampleType:
		return "unknown"
	case RawSampleType:
		return "raw"
	case StructDigestSampleType:
		return "struct-digest"
	case ValueDigestSampleType:
		return "value-digest"
	case EventSampleType:
		return "event"
	case ConfigSampleType:
		return "config"
	default:
		return "unknown"
	}
}

func (s SampleType) ToProto() protos.SampleType {
	return protos.SampleType(s)
}

func ParseSampleType(t string) SampleType {
	switch t {
	case "raw":
		return RawSampleType
	case "struct-digest":
		return StructDigestSampleType
	case "value-digest":
		return ValueDigestSampleType
	case "event":
		return EventSampleType
	case "config":
		return ConfigSampleType
	default:
		return UnknownSampleType
	}
}

type Event struct {
	UID        SamplerEventUID
	Name       string
	StreamUID  SamplerStreamUID
	SampleType SampleType
	Rule       Rule
}

func (e Event) GetName() string {
	return e.Name
}

func NewEventFromProto(protoEvent *protos.Event) Event {
	if protoEvent == nil {
		return Event{}
	}

	return Event{
		UID:        SamplerEventUID(protoEvent.GetUid()),
		Name:       protoEvent.Name,
		StreamUID:  SamplerStreamUID(protoEvent.GetStreamUid()),
		SampleType: NewSampleTypeFromProto(protoEvent.GetSampleType()),
		Rule:       NewRuleFromProto(protoEvent.GetRule()),
	}
}

func (e *Event) ToProto() *protos.Event {
	return &protos.Event{
		Uid:        string(e.UID),
		Name:       e.Name,
		StreamUid:  string(e.StreamUID),
		SampleType: e.SampleType.ToProto(),
		Rule:       e.Rule.ToProto(),
	}

}

type EventUpdateOp int

const (
	EventUpsert EventUpdateOp = iota + 1
	EventDelete
)

type EventUpdate struct {
	Op    EventUpdateOp
	Event Event
}

func NewEventUpdateFromProto(eventUpdate *protos.ClientEventUpdate) EventUpdate {
	if eventUpdate == nil {
		return EventUpdate{}
	}

	var op EventUpdateOp
	switch eventUpdate.GetOp() {
	case protos.ClientEventUpdate_UPSERT:
		op = EventUpsert
	case protos.ClientEventUpdate_DELETE:
		op = EventDelete
	}

	return EventUpdate{
		Op:    op,
		Event: NewEventFromProto(eventUpdate.GetEvent()),
	}
}
func (eu *EventUpdate) IsValid() error {
	isValid := uidValidationRegex.MatchString(string(eu.Event.UID))
	if !isValid {
		return fmt.Errorf(uidValidationErrTemplate, "event", eu.Event.UID)
	}
	return nil
}

func (eu *EventUpdate) ToProto() *protos.ClientEventUpdate {
	protoOp := protos.ClientEventUpdate_UNKNOWN
	switch eu.Op {
	case EventUpsert:
		protoOp = protos.ClientEventUpdate_UPSERT
	case EventDelete:
		protoOp = protos.ClientEventUpdate_DELETE
	}

	return &protos.ClientEventUpdate{
		Op:    protoOp,
		Event: eu.Event.ToProto(),
	}
}
