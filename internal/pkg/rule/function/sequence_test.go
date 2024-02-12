package function

import (
	"testing"

	"golang.org/x/exp/constraints"
)

func TestCallSequence(t *testing.T) {
	zeroInt64 := int64(0)

	type args[T constraints.Ordered] struct {
		state *SequenceStateOf[T]
		value T
	}
	tests := []struct {
		name   string
		fields any
		args   any
		want   bool
	}{
		{
			name: "call function with same int64 value keeps ascendant order",
			args: args[int64]{
				state: &SequenceStateOf[int64]{
					Last:          &zeroInt64,
					ExpectedOrder: OrderTypeAsc,
					ResultOrder:   OrderTypeAsc,
				},
				value: 0,
			},
			want: true,
		},
		{
			name: "call function with increasing int64 value keeps ascendant order",
			args: args[int64]{
				state: &SequenceStateOf[int64]{
					Last:          &zeroInt64,
					ExpectedOrder: OrderTypeAsc,
					ResultOrder:   OrderTypeAsc,
				},
				value: 1,
			},
			want: true,
		},
		{
			name: "call function with decreasing int64 value violates ascendant order",
			args: args[int64]{
				state: &SequenceStateOf[int64]{
					Last:          &zeroInt64,
					ExpectedOrder: OrderTypeAsc,
					ResultOrder:   OrderTypeAsc,
				},
				value: -1,
			},
			want: false,
		},
		{
			name: "call function with same int64 value keeps descendant order",
			args: args[int64]{
				state: &SequenceStateOf[int64]{
					Last:          &zeroInt64,
					ExpectedOrder: OrderTypeDesc,
					ResultOrder:   OrderTypeDesc,
				},
				value: 0,
			},
			want: true,
		},
		{
			name: "call function with decreasing int64 value keeps descendant order",
			args: args[int64]{
				state: &SequenceStateOf[int64]{
					Last:          &zeroInt64,
					ExpectedOrder: OrderTypeDesc,
					ResultOrder:   OrderTypeDesc,
				},
				value: -1,
			},
			want: true,
		},
		{
			name: "call function with increasing int64 value violates descendant order",
			args: args[int64]{
				state: &SequenceStateOf[int64]{
					Last:          &zeroInt64,
					ExpectedOrder: OrderTypeDesc,
					ResultOrder:   OrderTypeDesc,
				},
				value: 1,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got bool
			switch v := tt.args.(type) {
			case args[int64]:
				got = CallSequence(v.state, v.value)
			case args[uint64]:
				got = CallSequence(v.state, v.value)
			case args[float64]:
				got = CallSequence(v.state, v.value)
			case args[string]:
				got = CallSequence(v.state, v.value)
			default:
				t.Error("unknown type")
				t.FailNow()
			}

			if got != tt.want {
				t.Errorf("CallSequence() = %v, want %v", got, tt.want)
			}
		})
	}
}
