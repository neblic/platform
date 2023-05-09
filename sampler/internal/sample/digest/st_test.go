package digest

import (
	"testing"

	"github.com/neblic/platform/sampler/internal/sample"
	"github.com/neblic/platform/sampler/protos"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

const maxProcessedFieldsDef = 10

var notifyErrDef = func(t *testing.T) func(error) {
	return func(err error) { assert.NoError(t, err) }
}

func TestUpdateValueNum(t *testing.T) {
	testCases := []struct {
		desc string

		prevDigest    *protos.Value
		values        []interface{}
		updatedDigest *protos.Value
	}{
		{
			desc:   "initialize digest",
			values: []interface{}{int64(1)},
			updatedDigest: &protos.Value{
				Number: &protos.NumberType{
					IntegerNum: &protos.IntNumType{
						Count: 1,
					},
				},
			},
		},
		{
			desc: "update integers",
			prevDigest: &protos.Value{
				Number: &protos.NumberType{
					IntegerNum: &protos.IntNumType{
						Count: 0,
					},
				},
			},
			values: []interface{}{int(-1), int8(-2), int16(-3), int32(-4), int64(-5)},
			updatedDigest: &protos.Value{
				Number: &protos.NumberType{
					IntegerNum: &protos.IntNumType{
						Count: 5,
					},
				},
			},
		},
		{
			desc: "update uintegers",
			prevDigest: &protos.Value{
				Number: &protos.NumberType{
					UintegerNum: &protos.UIntNumType{
						Count: 0,
					},
				},
			},
			values: []interface{}{uint(1), uint8(2), uint16(3), uint32(4), uint64(5)},
			updatedDigest: &protos.Value{
				Number: &protos.NumberType{
					UintegerNum: &protos.UIntNumType{
						Count: 5,
					},
				},
			},
		},
		{
			desc: "update floats",
			prevDigest: &protos.Value{
				Number: &protos.NumberType{
					FloatNum: &protos.FloatNumType{
						Count: 0,
					},
				},
			},
			values: []interface{}{float32(1.1), float64(2.2)},
			updatedDigest: &protos.Value{
				Number: &protos.NumberType{
					FloatNum: &protos.FloatNumType{
						Count: 2,
					},
				},
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			st := NewStDigest(maxProcessedFieldsDef, notifyErrDef(t))
			var (
				finalDigest *protos.Value
				err         error
			)
			for _, val := range tC.values {
				finalDigest, err = st.updateValue(tC.prevDigest, val)
				require.NoError(t, err)
			}
			assert.True(t, proto.Equal(finalDigest, tC.updatedDigest),
				"got: %v, want: %v", finalDigest, tC.updatedDigest)
		})
	}
}

func TestUpdateValueString(t *testing.T) {
	testCases := []struct {
		desc string

		prevDigest    *protos.Value
		values        []string
		updatedDigest *protos.Value
	}{
		{
			desc:   "initialize digest",
			values: []string{"a"},
			updatedDigest: &protos.Value{
				String_: &protos.StringType{
					Count: 1,
				},
			},
		},
		{
			desc: "update strings",
			prevDigest: &protos.Value{
				String_: &protos.StringType{
					Count: 0,
				},
			},
			values: []string{"a", "b"},
			updatedDigest: &protos.Value{
				String_: &protos.StringType{
					Count: 2,
				},
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			st := NewStDigest(maxProcessedFieldsDef, notifyErrDef(t))
			var (
				finalDigest *protos.Value
				err         error
			)
			for _, val := range tC.values {
				finalDigest, err = st.updateValue(tC.prevDigest, val)
				require.NoError(t, err)
			}
			assert.True(t, proto.Equal(finalDigest, tC.updatedDigest),
				"got: %v, want: %v", finalDigest, tC.updatedDigest)
		})
	}
}

func TestUpdateValueBoolean(t *testing.T) {
	testCases := []struct {
		desc string

		prevDigest    *protos.Value
		values        []bool
		updatedDigest *protos.Value
	}{
		{
			desc:   "initialize digest",
			values: []bool{true},
			updatedDigest: &protos.Value{
				Boolean: &protos.BooleanType{
					Count: 1,
				},
			},
		},
		{
			desc: "update booleans",
			prevDigest: &protos.Value{
				Boolean: &protos.BooleanType{
					Count: 0,
				},
			},
			values: []bool{true, false},
			updatedDigest: &protos.Value{
				Boolean: &protos.BooleanType{
					Count: 2,
				},
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			st := NewStDigest(maxProcessedFieldsDef, notifyErrDef(t))
			var (
				finalDigest *protos.Value
				err         error
			)
			for _, val := range tC.values {
				finalDigest, err = st.updateValue(tC.prevDigest, val)
				require.NoError(t, err)
			}
			assert.True(t, proto.Equal(finalDigest, tC.updatedDigest),
				"got: %v, want: %v", finalDigest, tC.updatedDigest)
		})
	}
}

func TestUpdateValueArray(t *testing.T) {
	valueUint := func(count int64) *protos.Value {
		return &protos.Value{Number: &protos.NumberType{UintegerNum: &protos.UIntNumType{Count: count}}}
	}
	valueString := func(count int64) *protos.Value {
		return &protos.Value{String_: &protos.StringType{Count: count}}
	}
	valueArrayUint := func(arrCount, UintCount int64) *protos.Value {
		return &protos.Value{Array: &protos.ArrayType{
			Count:                   arrCount,
			FixedLengthOrderedArray: &protos.FixedLengthOrderedArrayType{Fields: []*protos.Value{valueUint(UintCount)}}},
		}
	}

	testCases := []struct {
		desc string

		prevDigest    *protos.Value
		values        [][]interface{}
		updatedDigest *protos.Value
	}{
		{
			desc:   "initialize digest",
			values: [][]interface{}{{uint(1)}},
			updatedDigest: &protos.Value{
				Array: &protos.ArrayType{
					Count: 1,
					FixedLengthOrderedArray: &protos.FixedLengthOrderedArrayType{
						Fields: []*protos.Value{valueUint(1)},
					},
				},
			},
		},
		{
			desc:   "update fixed length order array, single type",
			values: [][]interface{}{{uint(1)}},
			prevDigest: &protos.Value{
				Array: &protos.ArrayType{
					Count: 1,
					FixedLengthOrderedArray: &protos.FixedLengthOrderedArrayType{
						Fields: []*protos.Value{valueUint(1)},
					},
				},
			},
			updatedDigest: &protos.Value{
				Array: &protos.ArrayType{
					Count: 2,
					FixedLengthOrderedArray: &protos.FixedLengthOrderedArrayType{
						Fields: []*protos.Value{valueUint(2)},
					},
				},
			},
		},
		{
			desc:   "update fixed length order array, mixed array types",
			values: [][]interface{}{{uint(1), "a"}},
			prevDigest: &protos.Value{
				Array: &protos.ArrayType{
					Count: 1,
					FixedLengthOrderedArray: &protos.FixedLengthOrderedArrayType{
						Fields: []*protos.Value{valueUint(1), valueString(1)},
					},
				},
			},
			updatedDigest: &protos.Value{
				Array: &protos.ArrayType{
					Count: 2,
					FixedLengthOrderedArray: &protos.FixedLengthOrderedArrayType{
						Fields: []*protos.Value{valueUint(2), valueString(2)},
					},
				},
			},
		},
		{
			desc:   "update fixed length order array, nested arrays",
			values: [][]interface{}{{[]interface{}{uint(1)}}},
			prevDigest: &protos.Value{
				Array: &protos.ArrayType{
					Count: 1,
					FixedLengthOrderedArray: &protos.FixedLengthOrderedArrayType{
						Fields: []*protos.Value{valueArrayUint(1, 1)},
					},
				},
			},
			updatedDigest: &protos.Value{
				Array: &protos.ArrayType{
					Count: 2,
					FixedLengthOrderedArray: &protos.FixedLengthOrderedArrayType{
						Fields: []*protos.Value{valueArrayUint(2, 2)},
					},
				},
			},
		},
		{
			desc:   "update variable length order array",
			values: [][]interface{}{{uint(1), "a"}},
			prevDigest: &protos.Value{
				Array: &protos.ArrayType{
					Count: 2,
					VariableLengthArray: &protos.VariableLengthArrayType{
						MinLength: 1,
						MaxLength: 3,
						SumLength: 4,
					},
				},
			},
			updatedDigest: &protos.Value{
				Array: &protos.ArrayType{
					Count: 3,
					VariableLengthArray: &protos.VariableLengthArrayType{
						MinLength: 1,
						MaxLength: 3,
						SumLength: 6,
					},
				},
			},
		},
		{
			desc:   "update fixed length order array to variable length order array",
			values: [][]interface{}{{uint(1), "a", true}},
			prevDigest: &protos.Value{
				Array: &protos.ArrayType{
					Count: 1,
					FixedLengthOrderedArray: &protos.FixedLengthOrderedArrayType{
						Fields: []*protos.Value{valueUint(1), valueString(1)},
					},
				},
			},
			updatedDigest: &protos.Value{
				Array: &protos.ArrayType{
					Count: 2,
					VariableLengthArray: &protos.VariableLengthArrayType{
						MinLength: 2,
						MaxLength: 3,
						SumLength: 5,
					},
				},
			},
		},
	}

	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			st := NewStDigest(maxProcessedFieldsDef, notifyErrDef(t))
			var (
				finalDigest *protos.Value
				err         error
			)
			for _, val := range tC.values {
				finalDigest, err = st.updateValue(tC.prevDigest, val)
				require.NoError(t, err)
			}
			assert.True(t, proto.Equal(finalDigest, tC.updatedDigest),
				"got: %v, want: %v", finalDigest, tC.updatedDigest)
		})
	}
}

func TestUpdateValueObj(t *testing.T) {
	valueUint := func(count int64) *protos.Value {
		return &protos.Value{Number: &protos.NumberType{UintegerNum: &protos.UIntNumType{Count: count}}}
	}
	valueString := func(count int64) *protos.Value {
		return &protos.Value{String_: &protos.StringType{Count: count}}
	}
	valueObjUint := func(objCount, uintCount int64) *protos.Value {
		return &protos.Value{Obj: &protos.ObjType{Count: objCount, Fields: map[string]*protos.Value{"nested_key": valueUint(uintCount)}}}
	}

	testCases := []struct {
		desc string

		prevDigest    *protos.Value
		values        []map[string]interface{}
		updatedDigest *protos.Value
	}{
		{
			desc:   "initialize digest",
			values: []map[string]interface{}{{"key": uint(1)}},
			updatedDigest: &protos.Value{
				Obj: &protos.ObjType{
					Count: 1,
					Fields: map[string]*protos.Value{
						"key": valueUint(1),
					},
				},
			},
		},
		{
			desc:   "update object, single type",
			values: []map[string]interface{}{{"key": uint(1)}},
			prevDigest: &protos.Value{
				Obj: &protos.ObjType{
					Count: 1,
					Fields: map[string]*protos.Value{
						"key": valueUint(1),
					},
				},
			},
			updatedDigest: &protos.Value{
				Obj: &protos.ObjType{
					Count: 2,
					Fields: map[string]*protos.Value{
						"key": valueUint(2),
					},
				},
			},
		},
		{
			desc:   "update object, mixed types",
			values: []map[string]interface{}{{"key": uint(1), "nested_key": "a"}},
			prevDigest: &protos.Value{
				Obj: &protos.ObjType{
					Count: 1,
					Fields: map[string]*protos.Value{
						"key":        valueUint(1),
						"nested_key": valueString(1),
					},
				},
			},
			updatedDigest: &protos.Value{
				Obj: &protos.ObjType{
					Count: 2,
					Fields: map[string]*protos.Value{
						"key":        valueUint(2),
						"nested_key": valueString(2),
					},
				},
			},
		},
		{
			desc:   "update object, nested objects",
			values: []map[string]interface{}{{"key": map[string]interface{}{"nested_key": uint(1)}}},
			prevDigest: &protos.Value{
				Obj: &protos.ObjType{
					Count: 1,
					Fields: map[string]*protos.Value{
						"key": valueObjUint(1, 1),
					},
				},
			},
			updatedDigest: &protos.Value{
				Obj: &protos.ObjType{
					Count: 2,
					Fields: map[string]*protos.Value{
						"key": valueObjUint(2, 2),
					},
				},
			},
		},
	}

	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			st := NewStDigest(maxProcessedFieldsDef, notifyErrDef(t))
			var (
				finalDigest *protos.Value
				err         error
			)
			for _, val := range tC.values {
				finalDigest, err = st.updateValue(tC.prevDigest, val)
				require.NoError(t, err)
			}
			assert.True(t, proto.Equal(finalDigest, tC.updatedDigest),
				"got: %v, want: %v", finalDigest, tC.updatedDigest)
		})
	}
}

func TestUpdateValueMixed(t *testing.T) {
	testCases := []struct {
		desc string

		prevDigest    *protos.Value
		values        []interface{}
		updatedDigest *protos.Value
	}{
		{
			desc:   "update value, mixed types",
			values: []interface{}{"a"},
			prevDigest: &protos.Value{
				Number: &protos.NumberType{UintegerNum: &protos.UIntNumType{Count: 1}},
			},
			updatedDigest: &protos.Value{
				Number:  &protos.NumberType{UintegerNum: &protos.UIntNumType{Count: 1}},
				String_: &protos.StringType{Count: 1},
			},
		},
	}

	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			st := NewStDigest(maxProcessedFieldsDef, notifyErrDef(t))
			var (
				finalDigest *protos.Value
				err         error
			)
			for _, val := range tC.values {
				finalDigest, err = st.updateValue(tC.prevDigest, val)
				require.NoError(t, err)
			}
			assert.True(t, proto.Equal(finalDigest, tC.updatedDigest),
				"got: %v, want: %v", finalDigest, tC.updatedDigest)
		})
	}
}

type sampleStruct struct {
	Key int
}

func TestBuildDigest(t *testing.T) {
	valueFloat := func(count int64) *protos.Value {
		return &protos.Value{Number: &protos.NumberType{FloatNum: &protos.FloatNumType{Count: count}}}
	}
	valueInt := func(count int64) *protos.Value {
		return &protos.Value{Number: &protos.NumberType{IntegerNum: &protos.IntNumType{Count: count}}}
	}

	testCases := []struct {
		desc       string
		sample     *sample.Data
		wantDigest *protos.StructureDigest
	}{
		{
			desc:   "add sample from JSON",
			sample: sample.NewSampleDataFromJSON(`{"key": 1}`),
			wantDigest: &protos.StructureDigest{
				Obj: &protos.ObjType{
					Count: 1,
					Fields: map[string]*protos.Value{
						"key": valueFloat(1), // all numbers in JSON are represented as float
					},
				},
			},
		},
		{
			desc: "add sample from proto",
			sample: sample.NewSampleDataFromProto(&protos.TestSample{
				Int32: 1,
			}),
			wantDigest: &protos.StructureDigest{
				Obj: &protos.ObjType{
					Count: 1,
					Fields: map[string]*protos.Value{
						"int32": valueInt(1),
					},
				},
			},
		},
		{
			desc:   "add sample from native",
			sample: sample.NewSampleDataFromNative(sampleStruct{Key: 1}),
			wantDigest: &protos.StructureDigest{
				Obj: &protos.ObjType{
					Count: 1,
					Fields: map[string]*protos.Value{
						"Key": valueInt(1),
					},
				},
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			st := NewStDigest(maxProcessedFieldsDef, notifyErrDef(t))
			err := st.AddSampleData(tC.sample)
			require.NoError(t, err)

			JSONDigest, err := st.JSON()
			require.NoError(t, err)

			var protoDigest protos.StructureDigest
			err = protojson.Unmarshal([]byte(JSONDigest), &protoDigest)
			require.NoError(t, err)

			assert.True(t, proto.Equal(&protoDigest, tC.wantDigest),
				"got: %v, want: %v", &protoDigest, tC.wantDigest)
		})
	}
}
