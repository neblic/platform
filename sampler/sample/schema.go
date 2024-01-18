package sample

import "google.golang.org/protobuf/proto"

// Schema defines a sampler schema.
type Schema interface {
	isSchema()
}

// DynamicSchema defines a schema-free sample.
type DynamicSchema struct{}

func (DynamicSchema) isSchema() {}

// NewDynamicSchema creates a new DynamicSchema
func NewDynamicSchema() DynamicSchema {
	return DynamicSchema{}
}

// ProtoSchema defines a proto-based sample.
type ProtoSchema struct {
	Proto proto.Message
}

// NewProtoSchema creates a new ProtoSchema. The proto argument defines the sample schema.
func NewProtoSchema(proto proto.Message) ProtoSchema {
	return ProtoSchema{Proto: proto}
}

func (ProtoSchema) isSchema() {}
