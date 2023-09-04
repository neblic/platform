package data

import (
	"testing"

	"github.com/neblic/platform/sampler/protos"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestProtoValueToNative(t *testing.T) {
	testCases := []struct {
		desc    string
		proto   proto.Message
		wantMap map[string]interface{}
	}{
		{
			desc: "all fields are converted",
			proto: &protos.TestSample{
				Double:   1,
				Float:    1,
				Int32:    1,
				Int64:    1,
				Uint32:   1,
				Uint64:   1,
				Sint32:   1,
				Sint64:   1,
				Fixed32:  1,
				Fixed64:  1,
				Sfixed32: 1,
				Sfixed64: 1,
				Bool:     true,
				String_:  "1",
				Bytes:    []byte("12"),
				Array:    []int32{1, 2},
				Map: map[int32]int32{
					int32(1): int32(1),
				},
				NestedMsg: &protos.TestSample{
					Double: 1,
				},
				NestedMsgs: []*protos.TestSample{
					{
						Double: 1,
					},
					{
						Double: 2,
					},
				},
			},
			wantMap: map[string]interface{}{
				"double":   float64(1.0),
				"float":    float32(1.0),
				"int32":    int32(1),
				"int64":    int64(1),
				"uint32":   uint32(1),
				"uint64":   uint64(1),
				"sint32":   int32(1),
				"sint64":   int64(1),
				"fixed32":  uint32(1),
				"fixed64":  uint64(1),
				"sfixed32": int32(1),
				"sfixed64": int64(1),
				"bool":     true,
				"string":   "1",
				"bytes":    []byte("12"),
				"array":    []interface{}{int32(1), int32(2)},
				"map":      map[interface{}]interface{}{int32(1): int32(1)},
				"nested_msg": map[string]interface{}{
					"double": float64(1.0),
				},
				"nested_msgs": []interface{}{
					map[string]interface{}{
						"double": float64(1.0),
					},
					map[string]interface{}{
						"double": float64(2.0),
					},
				},
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			gotMap, err := protoObjectToMap(tC.proto.ProtoReflect())
			require.NoError(t, err)
			assert.Equal(t, tC.wantMap, gotMap)
		})
	}
}
