package metric

import (
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/neblic/platform/controlplane/control"
	"go.opentelemetry.io/collector/pdata/pcommon"
)

type FieldType string

const (
	NumberType  FieldType = "number"
	StringType  FieldType = "string"
	BooleanType FieldType = "boolean"
	ArrayType   FieldType = "array"
	ObjectType  FieldType = "obj"
	AnyType     FieldType = "any"
)

type Path string

func NewPath() Path {
	return Path("")
}

func (p Path) AddPart(field string, fieldType FieldType) Path {
	field = `"` + strings.ReplaceAll(strings.ReplaceAll(field, `\`, `\\`), `"`, `\"`) + `"`
	p = p + Path(`[`+field+`:`+string(fieldType)+`]`)

	return p
}

func (p Path) IsEmpty() bool {
	return p == ""
}

func (p Path) String() string {
	return string(p)
}

// type MetricAttributes struct {
// 	attributes pcommon.Map
// }

// func NewMetricAttributes() MetricAttributes {
// 	return MetricAttributes{
// 		attributes: pcommon.NewMap(),
// 	}
// }

// func (ma MetricAttributes) WithPath(path Path) MetricAttributes {
// 	ma.attributes.PutStr(OTLPSampleFieldPathKey, path.String())
// 	return ma
// }

// func (ma MetricAttributes) Path() (string, bool) {
// 	pathValue, ok := ma.attributes.Get(OTLPSampleFieldPathKey)
// 	if !ok {
// 		return "", false
// 	}

// 	return pathValue.AsString(), true

// }

// func (ma MetricAttributes) WithName(name string) MetricAttributes {
// 	ma.attributes.PutStr(OTLPSampleNameKey, name)
// 	return ma
// }

// func (ma MetricAttributes) Name() (string, bool) {
// 	nameValue, ok := ma.attributes.Get(OTLPSampleNameKey)
// 	if !ok {
// 		return "", false
// 	}

// 	return nameValue.AsString(), true
// }

type DatapointAttributes struct {
	tsUnixNano int64
	attributes pcommon.Map
}

func NewDatapointAttributes() DatapointAttributes {
	return DatapointAttributes{
		tsUnixNano: 0,
		attributes: pcommon.NewMap(),
	}
}

func (da DatapointAttributes) WithTs(ts time.Time) DatapointAttributes {
	da.tsUnixNano = ts.UTC().UnixNano()
	return da
}

func (da DatapointAttributes) WithSampleType(sampleType control.SampleType) DatapointAttributes {
	da.attributes.PutStr(OTLPSampleSampleTypeKey, sampleType.String())
	return da
}

func (da DatapointAttributes) WithDigestUID(uid uuid.UUID) DatapointAttributes {
	da.attributes.PutStr(OTLPSampleDigestUIDKey, uid.String())
	return da
}

func (da DatapointAttributes) WithEventUID(uid uuid.UUID) DatapointAttributes {
	da.attributes.PutStr(OTLPSampleEventUIDKey, uid.String())
	return da
}

func (da DatapointAttributes) WithStreamUID(uid string) DatapointAttributes {
	da.attributes.PutStr(OTLPSampleStreamUIDKey, uid)
	return da
}
