package types

import (
	"errors"
	"math"
	"strconv"

	"github.com/axiomhq/hyperloglog"
	"github.com/neblic/platform/dataplane/protos"
)

type MinStat struct {
	Value float64
}

func NewMinStat() *MinStat {
	return &MinStat{
		Value: math.Inf(+1),
	}
}

func NewMinStatFromProto(minStat *protos.MinStat) *MinStat {
	return &MinStat{
		Value: minStat.Value,
	}
}

func (ms *MinStat) ToProto() *protos.MinStat {
	return &protos.MinStat{
		Value: ms.Value,
	}
}

type AvgStat struct {
	Sum   float64
	Count uint64
}

func NewAvgStat() *AvgStat {
	return &AvgStat{
		Sum:   0,
		Count: 0,
	}
}

func NewAvgStatFromProto(avgStat *protos.AvgStat) *AvgStat {
	return &AvgStat{
		Sum:   avgStat.Sum,
		Count: avgStat.Count,
	}
}

func (as *AvgStat) ToProto() *protos.AvgStat {
	return &protos.AvgStat{
		Sum:   as.Sum,
		Count: as.Count,
	}
}

type MaxStat struct {
	Value float64
}

func NewMaxStat() *MaxStat {
	return &MaxStat{
		Value: math.Inf(-1),
	}
}

type HyperLogLog struct {
	Sketch      *hyperloglog.Sketch
	Cardinality uint64
}

func NewHyperLogLog() *HyperLogLog {
	return &HyperLogLog{
		Sketch:      hyperloglog.New14(),
		Cardinality: 0,
	}
}

func NewHyperLogLogFromProto(hyperLogLog *protos.HyperLogLog) (*HyperLogLog, error) {
	sketch := hyperloglog.New14()
	err := sketch.UnmarshalBinary(hyperLogLog.Data)

	return &HyperLogLog{
		Sketch:      sketch,
		Cardinality: hyperLogLog.Cardinality,
	}, err
}

func (hll *HyperLogLog) InsertBytes(data []byte) *HyperLogLog {
	hll.Sketch.Insert(data)

	return hll
}

func (hll *HyperLogLog) InsertInt64(data int64) *HyperLogLog {
	hll.Sketch.Insert([]byte(strconv.FormatInt(data, 10)))

	return hll
}

func (hll *HyperLogLog) InsertFloat64(data float64) *HyperLogLog {
	hll.Sketch.Insert([]byte(strconv.FormatFloat(data, 'E', -1, 64)))

	return hll
}

func (hll *HyperLogLog) ToProto() *protos.HyperLogLog {
	// Sketch MarshalBinary returns an error to follow the encoding.BinaryMarshaler interface,
	// but it never returns an error.
	data, _ := hll.Sketch.MarshalBinary()

	return &protos.HyperLogLog{
		Data:        data,
		Cardinality: hll.Sketch.Estimate(),
	}
}

func NewMaxStatFromProto(maxStat *protos.MaxStat) *MaxStat {
	return &MaxStat{
		Value: maxStat.Value,
	}
}

func (ms *MaxStat) ToProto() *protos.MaxStat {
	return &protos.MaxStat{
		Value: ms.Value,
	}
}

type NumberStat struct {
	Min         *MinStat
	Avg         *AvgStat
	Max         *MaxStat
	HyperLogLog *HyperLogLog
}

func NewNumberStat() *NumberStat {
	return &NumberStat{
		Min:         NewMinStat(),
		Avg:         NewAvgStat(),
		Max:         NewMaxStat(),
		HyperLogLog: NewHyperLogLog(),
	}
}

func NewNumberStatFromProto(numberStat *protos.NumberStat) (*NumberStat, error) {
	hyperLogLog, err := NewHyperLogLogFromProto(numberStat.HyperLogLog)

	return &NumberStat{
		Min:         NewMinStatFromProto(numberStat.Min),
		Avg:         NewAvgStatFromProto(numberStat.Avg),
		Max:         NewMaxStatFromProto(numberStat.Max),
		HyperLogLog: hyperLogLog,
	}, err
}

func (ns *NumberStat) ToProto() *protos.NumberStat {
	return &protos.NumberStat{
		Min:         ns.Min.ToProto(),
		Avg:         ns.Avg.ToProto(),
		Max:         ns.Max.ToProto(),
		HyperLogLog: ns.HyperLogLog.ToProto(),
	}
}

type NumberValue struct {
	TotalCount   uint64
	DefaultCount uint64
	NullCount    uint64
	Min          *MinStat
	Avg          *AvgStat
	Max          *MaxStat
	HyperLogLog  *HyperLogLog
}

func NewNumberValue() *NumberValue {
	return &NumberValue{
		TotalCount:   0,
		DefaultCount: 0,
		NullCount:    0,

		Min:         NewMinStat(),
		Avg:         NewAvgStat(),
		Max:         NewMaxStat(),
		HyperLogLog: NewHyperLogLog(),
	}
}

func NewNumberValueFromProto(numberValue *protos.NumberValue) (*NumberValue, error) {
	hyperLogLog, err := NewHyperLogLogFromProto(numberValue.HyperLogLog)

	return &NumberValue{
		TotalCount:   numberValue.TotalCount,
		DefaultCount: numberValue.DefaultCount,
		NullCount:    numberValue.NullCount,

		Min:         NewMinStatFromProto(numberValue.Min),
		Avg:         NewAvgStatFromProto(numberValue.Avg),
		Max:         NewMaxStatFromProto(numberValue.Max),
		HyperLogLog: hyperLogLog,
	}, err
}

func (nv *NumberValue) ToProto() *protos.NumberValue {
	return &protos.NumberValue{
		TotalCount:   nv.TotalCount,
		DefaultCount: nv.DefaultCount,
		NullCount:    nv.NullCount,

		Min:         nv.Min.ToProto(),
		Avg:         nv.Avg.ToProto(),
		Max:         nv.Max.ToProto(),
		HyperLogLog: nv.HyperLogLog.ToProto(),
	}
}

type StringValue struct {
	TotalCount   uint64
	DefaultCount uint64
	NullCount    uint64
	HyperLogLog  *HyperLogLog
	Length       *NumberStat
}

func NewStringValue() *StringValue {
	return &StringValue{
		TotalCount:   0,
		DefaultCount: 0,
		NullCount:    0,
		HyperLogLog:  NewHyperLogLog(),
		Length:       NewNumberStat(),
	}
}

func NewStringValueFromProto(stringValue *protos.StringValue) (*StringValue, error) {
	hyperLogLog, err := NewHyperLogLogFromProto(stringValue.HyperLogLog)
	length, errLength := NewNumberStatFromProto(stringValue.Length)

	return &StringValue{
		TotalCount:   stringValue.TotalCount,
		DefaultCount: stringValue.DefaultCount,
		NullCount:    stringValue.NullCount,

		HyperLogLog: hyperLogLog,

		Length: length,
	}, errors.Join(err, errLength)
}

func (sv *StringValue) ToProto() *protos.StringValue {
	return &protos.StringValue{
		TotalCount:   sv.TotalCount,
		DefaultCount: sv.DefaultCount,
		NullCount:    sv.NullCount,

		HyperLogLog: sv.HyperLogLog.ToProto(),

		Length: sv.Length.ToProto(),
	}
}

type BooleanValue struct {
	TotalCount   uint64
	DefaultCount uint64
	NullCount    uint64
	FalseCount   uint64
	TrueCount    uint64
}

func NewBooleanValue() *BooleanValue {
	return &BooleanValue{
		TotalCount:   0,
		DefaultCount: 0,
		NullCount:    0,

		FalseCount: 0,
		TrueCount:  0,
	}
}

func NewBooleanValueFromProto(booleanValue *protos.BooleanValue) *BooleanValue {
	return &BooleanValue{
		TotalCount:   booleanValue.TotalCount,
		DefaultCount: booleanValue.DefaultCount,
		NullCount:    booleanValue.NullCount,

		FalseCount: booleanValue.FalseCount,
		TrueCount:  booleanValue.TrueCount,
	}
}

func (bv *BooleanValue) ToProto() *protos.BooleanValue {
	return &protos.BooleanValue{
		TotalCount:   bv.TotalCount,
		DefaultCount: bv.DefaultCount,
		NullCount:    bv.NullCount,

		FalseCount: bv.FalseCount,
		TrueCount:  bv.TrueCount,
	}
}

type ArrayValue struct {
	TotalCount   uint64
	DefaultCount uint64
	NullCount    uint64
	Values       *ValueValue
}

func NewArrayValue() *ArrayValue {
	return &ArrayValue{
		TotalCount:   0,
		DefaultCount: 0,
		NullCount:    0,

		Values: NewValueValue(),
	}
}

func NewArrayValueFromProto(arrayValue *protos.ArrayValue) (*ArrayValue, error) {
	values, err := NewValueValueFromProto(arrayValue.Values)

	return &ArrayValue{
		TotalCount:   arrayValue.TotalCount,
		DefaultCount: arrayValue.DefaultCount,
		NullCount:    arrayValue.NullCount,

		Values: values,
	}, err
}

func (av *ArrayValue) ToProto() *protos.ArrayValue {
	return &protos.ArrayValue{
		TotalCount:   av.TotalCount,
		DefaultCount: av.DefaultCount,
		NullCount:    av.NullCount,

		Values: av.Values.ToProto(),
	}
}

type ObjValue struct {
	TotalCount   uint64
	DefaultCount uint64
	NullCount    uint64
	Fields       map[string]*ValueValue
}

func NewObjValue() *ObjValue {
	return &ObjValue{
		TotalCount:   0,
		DefaultCount: 0,
		NullCount:    0,

		Fields: map[string]*ValueValue{},
	}
}

func NewObjValueFromProto(objValue *protos.ObjValue) (*ObjValue, error) {
	fields := map[string]*ValueValue{}
	var errs error
	for key, value := range objValue.Fields {
		var err error
		fields[key], err = NewValueValueFromProto(value)
		errs = errors.Join(errs, err)
	}

	return &ObjValue{
		TotalCount:   objValue.TotalCount,
		DefaultCount: objValue.DefaultCount,
		NullCount:    objValue.NullCount,

		Fields: fields,
	}, errs
}

func (ov *ObjValue) ToProto() *protos.ObjValue {
	fields := map[string]*protos.ValueValue{}
	for key, value := range ov.Fields {
		fields[key] = value.ToProto()
	}

	return &protos.ObjValue{
		TotalCount:   ov.TotalCount,
		DefaultCount: ov.DefaultCount,
		NullCount:    ov.NullCount,

		Fields: fields,
	}
}

type ValueValue struct {
	TotalCount uint64
	NullCount  uint64
	Number     *NumberValue
	String     *StringValue
	Boolean    *BooleanValue
	Array      *ArrayValue
	Obj        *ObjValue
}

func NewValueValue() *ValueValue {
	return &ValueValue{
		TotalCount: 0,
		NullCount:  0,

		Number:  nil,
		String:  nil,
		Boolean: nil,
		Array:   nil,
		Obj:     nil,
	}
}

func NewValueValueFromProto(valueValue *protos.ValueValue) (*ValueValue, error) {
	var errs error

	var number *NumberValue
	if valueValue.Number != nil {
		var err error
		number, err = NewNumberValueFromProto(valueValue.Number)
		errs = errors.Join(errs, err)
	}

	var str *StringValue
	if valueValue.String_ != nil {
		var err error
		str, err = NewStringValueFromProto(valueValue.String_)
		errs = errors.Join(errs, err)
	}

	var array *ArrayValue
	if valueValue.Array != nil {
		var err error
		array, err = NewArrayValueFromProto(valueValue.Array)
		errs = errors.Join(errs, err)
	}

	var obj *ObjValue
	if valueValue.Obj != nil {
		var err error
		obj, err = NewObjValueFromProto(valueValue.Obj)
		errs = errors.Join(errs, err)
	}

	return &ValueValue{
		TotalCount: valueValue.TotalCount,
		NullCount:  valueValue.NullCount,

		Number:  number,
		String:  str,
		Boolean: NewBooleanValueFromProto(valueValue.Boolean),
		Array:   array,
		Obj:     obj,
	}, errs
}

func (vv *ValueValue) ToProto() *protos.ValueValue {
	var number *protos.NumberValue
	if vv.Number != nil {
		number = vv.Number.ToProto()
	}
	var str *protos.StringValue
	if vv.String != nil {
		str = vv.String.ToProto()
	}
	var boolean *protos.BooleanValue
	if vv.Boolean != nil {
		boolean = vv.Boolean.ToProto()
	}
	var array *protos.ArrayValue
	if vv.Array != nil {
		array = vv.Array.ToProto()
	}
	var obj *protos.ObjValue
	if vv.Obj != nil {
		obj = vv.Obj.ToProto()
	}

	return &protos.ValueValue{
		TotalCount: vv.TotalCount,
		NullCount:  vv.NullCount,

		Number:  number,
		String_: str,
		Boolean: boolean,
		Array:   array,
		Obj:     obj,
	}
}
