package digest

import (
	"testing"

	"github.com/neblic/platform/dataplane/protos"
	"github.com/neblic/platform/dataplane/protos/test"
	"github.com/neblic/platform/internal/pkg/data"
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

		prevDigest    *protos.ValueSt
		values        []interface{}
		updatedDigest *protos.ValueSt
	}{
		{
			desc:   "initialize digest",
			values: []interface{}{int64(1)},
			updatedDigest: &protos.ValueSt{
				Number: &protos.NumberSt{
					IntegerNum: &protos.IntNumSt{
						Count: 1,
					},
				},
			},
		},
		{
			desc: "update integers",
			prevDigest: &protos.ValueSt{
				Number: &protos.NumberSt{
					IntegerNum: &protos.IntNumSt{
						Count: 0,
					},
				},
			},
			values: []interface{}{int(-1), int8(-2), int16(-3), int32(-4), int64(-5)},
			updatedDigest: &protos.ValueSt{
				Number: &protos.NumberSt{
					IntegerNum: &protos.IntNumSt{
						Count: 5,
					},
				},
			},
		},
		{
			desc: "update uintegers",
			prevDigest: &protos.ValueSt{
				Number: &protos.NumberSt{
					UintegerNum: &protos.UIntNumSt{
						Count: 0,
					},
				},
			},
			values: []interface{}{uint(1), uint8(2), uint16(3), uint32(4), uint64(5)},
			updatedDigest: &protos.ValueSt{
				Number: &protos.NumberSt{
					UintegerNum: &protos.UIntNumSt{
						Count: 5,
					},
				},
			},
		},
		{
			desc: "update floats",
			prevDigest: &protos.ValueSt{
				Number: &protos.NumberSt{
					FloatNum: &protos.FloatNumSt{
						Count: 0,
					},
				},
			},
			values: []interface{}{float32(1.1), float64(2.2)},
			updatedDigest: &protos.ValueSt{
				Number: &protos.NumberSt{
					FloatNum: &protos.FloatNumSt{
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
				finalDigest *protos.ValueSt
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

		prevDigest    *protos.ValueSt
		values        []string
		updatedDigest *protos.ValueSt
	}{
		{
			desc:   "initialize digest",
			values: []string{"a"},
			updatedDigest: &protos.ValueSt{
				String_: &protos.StringSt{
					Count: 1,
				},
			},
		},
		{
			desc: "update strings",
			prevDigest: &protos.ValueSt{
				String_: &protos.StringSt{
					Count: 0,
				},
			},
			values: []string{"a", "b"},
			updatedDigest: &protos.ValueSt{
				String_: &protos.StringSt{
					Count: 2,
				},
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			st := NewStDigest(maxProcessedFieldsDef, notifyErrDef(t))
			var (
				finalDigest *protos.ValueSt
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

		prevDigest    *protos.ValueSt
		values        []bool
		updatedDigest *protos.ValueSt
	}{
		{
			desc:   "initialize digest",
			values: []bool{true},
			updatedDigest: &protos.ValueSt{
				Boolean: &protos.BooleanSt{
					Count: 1,
				},
			},
		},
		{
			desc: "update booleans",
			prevDigest: &protos.ValueSt{
				Boolean: &protos.BooleanSt{
					Count: 0,
				},
			},
			values: []bool{true, false},
			updatedDigest: &protos.ValueSt{
				Boolean: &protos.BooleanSt{
					Count: 2,
				},
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			st := NewStDigest(maxProcessedFieldsDef, notifyErrDef(t))
			var (
				finalDigest *protos.ValueSt
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
	valueUint := func(count int64) *protos.ValueSt {
		return &protos.ValueSt{Number: &protos.NumberSt{UintegerNum: &protos.UIntNumSt{Count: count}}}
	}
	valueString := func(count int64) *protos.ValueSt {
		return &protos.ValueSt{String_: &protos.StringSt{Count: count}}
	}
	valueUintString := func(uintCount int64, stringCount int64) *protos.ValueSt {
		return &protos.ValueSt{
			Number:  valueUint(uintCount).Number,
			String_: valueString(stringCount).String_,
		}
	}

	testCases := []struct {
		desc string

		prevDigest    *protos.ValueSt
		values        [][]interface{}
		updatedDigest *protos.ValueSt
	}{
		{
			desc:   "initialize digest",
			values: [][]interface{}{{uint(1)}},
			updatedDigest: &protos.ValueSt{
				Array: &protos.ArraySt{
					Count:     1,
					Values:    valueUint(1),
					MinLength: 1,
					MaxLength: 1,
					SumLength: 1,
				},
			},
		},
		{
			desc:   "update array, single type",
			values: [][]interface{}{{uint(1)}},
			prevDigest: &protos.ValueSt{
				Array: &protos.ArraySt{
					Count:     1,
					Values:    valueUint(1),
					MinLength: 1,
					MaxLength: 1,
					SumLength: 1,
				},
			},
			updatedDigest: &protos.ValueSt{
				Array: &protos.ArraySt{
					Count:     2,
					Values:    valueUint(2),
					MinLength: 1,
					MaxLength: 1,
					SumLength: 2,
				},
			},
		},
		{
			desc:   "update array, mixed array types",
			values: [][]interface{}{{uint(1), "a"}},
			prevDigest: &protos.ValueSt{
				Array: &protos.ArraySt{
					Count:     1,
					Values:    valueUintString(1, 1),
					MinLength: 2,
					MaxLength: 2,
					SumLength: 2,
				},
			},
			updatedDigest: &protos.ValueSt{
				Array: &protos.ArraySt{
					Count:     2,
					Values:    valueUintString(2, 2),
					MinLength: 2,
					MaxLength: 2,
					SumLength: 4,
				},
			},
		},
		{
			desc:   "update array, nested",
			values: [][]interface{}{{[]interface{}{uint(1)}}},
			prevDigest: &protos.ValueSt{
				Array: &protos.ArraySt{
					Count: 1,
					Values: &protos.ValueSt{
						Array: &protos.ArraySt{
							Count:     2,
							Values:    valueUint(2),
							MinLength: 2,
							MaxLength: 2,
							SumLength: 4,
						},
					},
					MinLength: 1,
					MaxLength: 1,
					SumLength: 1,
				},
			},
			updatedDigest: &protos.ValueSt{
				Array: &protos.ArraySt{
					Count: 2,
					Values: &protos.ValueSt{
						Array: &protos.ArraySt{
							Count:     3,
							Values:    valueUint(3),
							MinLength: 1,
							MaxLength: 2,
							SumLength: 5,
						},
					},
					MinLength: 1,
					MaxLength: 1,
					SumLength: 2,
				},
			},
		},
	}

	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			st := NewStDigest(maxProcessedFieldsDef, notifyErrDef(t))
			var (
				finalDigest *protos.ValueSt
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
	valueUint := func(count int64) *protos.ValueSt {
		return &protos.ValueSt{Number: &protos.NumberSt{UintegerNum: &protos.UIntNumSt{Count: count}}}
	}
	valueString := func(count int64) *protos.ValueSt {
		return &protos.ValueSt{String_: &protos.StringSt{Count: count}}
	}
	valueObjUint := func(objCount, uintCount int64) *protos.ValueSt {
		return &protos.ValueSt{Obj: &protos.ObjSt{Count: objCount, Fields: map[string]*protos.ValueSt{"nested_key": valueUint(uintCount)}}}
	}

	testCases := []struct {
		desc string

		prevDigest    *protos.ValueSt
		values        []map[string]interface{}
		updatedDigest *protos.ValueSt
	}{
		{
			desc:   "initialize digest",
			values: []map[string]interface{}{{"key": uint(1)}},
			updatedDigest: &protos.ValueSt{
				Obj: &protos.ObjSt{
					Count: 1,
					Fields: map[string]*protos.ValueSt{
						"key": valueUint(1),
					},
				},
			},
		},
		{
			desc:   "update object, single type",
			values: []map[string]interface{}{{"key": uint(1)}},
			prevDigest: &protos.ValueSt{
				Obj: &protos.ObjSt{
					Count: 1,
					Fields: map[string]*protos.ValueSt{
						"key": valueUint(1),
					},
				},
			},
			updatedDigest: &protos.ValueSt{
				Obj: &protos.ObjSt{
					Count: 2,
					Fields: map[string]*protos.ValueSt{
						"key": valueUint(2),
					},
				},
			},
		},
		{
			desc:   "update object, mixed types",
			values: []map[string]interface{}{{"key": uint(1), "nested_key": "a"}},
			prevDigest: &protos.ValueSt{
				Obj: &protos.ObjSt{
					Count: 1,
					Fields: map[string]*protos.ValueSt{
						"key":        valueUint(1),
						"nested_key": valueString(1),
					},
				},
			},
			updatedDigest: &protos.ValueSt{
				Obj: &protos.ObjSt{
					Count: 2,
					Fields: map[string]*protos.ValueSt{
						"key":        valueUint(2),
						"nested_key": valueString(2),
					},
				},
			},
		},
		{
			desc:   "update object, nested objects",
			values: []map[string]interface{}{{"key": map[string]interface{}{"nested_key": uint(1)}}},
			prevDigest: &protos.ValueSt{
				Obj: &protos.ObjSt{
					Count: 1,
					Fields: map[string]*protos.ValueSt{
						"key": valueObjUint(1, 1),
					},
				},
			},
			updatedDigest: &protos.ValueSt{
				Obj: &protos.ObjSt{
					Count: 2,
					Fields: map[string]*protos.ValueSt{
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
				finalDigest *protos.ValueSt
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

		prevDigest    *protos.ValueSt
		values        []interface{}
		updatedDigest *protos.ValueSt
	}{
		{
			desc:   "update value, mixed types",
			values: []interface{}{"a"},
			prevDigest: &protos.ValueSt{
				Number: &protos.NumberSt{UintegerNum: &protos.UIntNumSt{Count: 1}},
			},
			updatedDigest: &protos.ValueSt{
				Number:  &protos.NumberSt{UintegerNum: &protos.UIntNumSt{Count: 1}},
				String_: &protos.StringSt{Count: 1},
			},
		},
	}

	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			st := NewStDigest(maxProcessedFieldsDef, notifyErrDef(t))
			var (
				finalDigest *protos.ValueSt
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
	valueFloat := func(count int64) *protos.ValueSt {
		return &protos.ValueSt{Number: &protos.NumberSt{FloatNum: &protos.FloatNumSt{Count: count}}}
	}
	valueInt := func(count int64) *protos.ValueSt {
		return &protos.ValueSt{Number: &protos.NumberSt{IntegerNum: &protos.IntNumSt{Count: count}}}
	}

	testCases := []struct {
		desc       string
		sample     *data.Data
		wantDigest *protos.StructureDigest
	}{
		{
			desc:   "add sample from JSON",
			sample: data.NewSampleDataFromJSON(`{"key": 1}`),
			wantDigest: &protos.StructureDigest{
				Obj: &protos.ObjSt{
					Count: 1,
					Fields: map[string]*protos.ValueSt{
						"key": valueFloat(1), // all numbers in JSON are represented as float
					},
				},
			},
		},
		{
			desc: "add sample from proto",
			sample: data.NewSampleDataFromProto(&test.TestSample{
				Int32: 1,
			}),
			wantDigest: &protos.StructureDigest{
				Obj: &protos.ObjSt{
					Count: 1,
					Fields: map[string]*protos.ValueSt{
						"int32": valueInt(1),
					},
				},
			},
		},
		{
			desc:   "add sample from native",
			sample: data.NewSampleDataFromNative(sampleStruct{Key: 1}),
			wantDigest: &protos.StructureDigest{
				Obj: &protos.ObjSt{
					Count: 1,
					Fields: map[string]*protos.ValueSt{
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
