package protos

import "math"

func NewMinStat() *MinStat {
	return &MinStat{
		Value: math.Inf(+1),
	}
}

func NewAvgStat() *AvgStat {
	return &AvgStat{
		Sum:   0,
		Count: 0,
	}
}

func NewMaxStat() *MaxStat {
	return &MaxStat{
		Value: math.Inf(-1),
	}
}

func NewNumberStat() *NumberStat {
	return &NumberStat{
		Min: NewMinStat(),
		Avg: NewAvgStat(),
		Max: NewMaxStat(),
	}
}

func NewNumberValue() *NumberValue {
	return &NumberValue{
		TotalCount:   0,
		DefaultCount: 0,
		NullCount:    0,

		Min: NewMinStat(),
		Avg: NewAvgStat(),
		Max: NewMaxStat(),
	}
}

func NewStringValue() *StringValue {
	return &StringValue{
		TotalCount:   0,
		DefaultCount: 0,
		NullCount:    0,

		Length: NewNumberStat(),
	}
}

func NewBooleanValue() *BooleanValue {
	return &BooleanValue{
		TotalCount:   0,
		DefaultCount: 0,
		NullCount:    0,

		TotalFalse: 0,
		TotalTrue:  0,
	}
}

func NewArrayValue() *ArrayValue {
	return &ArrayValue{
		TotalCount:   0,
		DefaultCount: 0,
		NullCount:    0,

		Values: NewValueValue(),
	}
}

func NewObjValue() *ObjValue {
	return &ObjValue{
		TotalCount:   0,
		DefaultCount: 0,
		NullCount:    0,

		Fields: map[string]*ValueValue{},
	}
}

func NewValueValue() *ValueValue {
	return &ValueValue{
		TotalCount: 0,
		NullCount:  0,

		Number:  nil,
		String_: nil,
		Boolean: nil,
		Array:   nil,
		Obj:     nil,
	}
}
