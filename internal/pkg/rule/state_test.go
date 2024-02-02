package rule

import (
	"testing"

	"golang.org/x/exp/constraints"
)

func TestSequenceState_Add(t *testing.T) {
	var zeroInt64 int64 = 0
	var zeroUint64 uint64 = 0
	var float64Zero float64 = 0
	var stringZero string = "0"

	type fields[T constraints.Ordered] struct {
		last          *T
		expectedOrder OrderType
		Order         OrderType
	}
	type args[T constraints.Ordered] struct {
		value T
	}
	tests := []struct {
		name   string
		fields any
		args   any
		want   bool
	}{
		{
			name: "add first int64 keeps ascendant order",
			fields: fields[int64]{
				last:          nil,
				expectedOrder: OrderTypeAsc,
				Order:         OrderTypeAsc,
			},
			args: args[int64]{
				value: 0,
			},
			want: true,
		},
		{
			name: "add non increasing int64 keeps ascendant order",
			fields: fields[int64]{
				last:          &zeroInt64,
				expectedOrder: OrderTypeAsc,
				Order:         OrderTypeAsc,
			},
			args: args[int64]{
				value: 0,
			},
			want: true,
		},
		{
			name: "add increasing int64 keeps ascendant order",
			fields: fields[int64]{
				last:          &zeroInt64,
				expectedOrder: OrderTypeAsc,
				Order:         OrderTypeAsc,
			},
			args: args[int64]{
				value: 1,
			},
			want: true,
		},
		{
			name: "add non decreasing int64 violates ascendant order",
			fields: fields[int64]{
				last:          &zeroInt64,
				expectedOrder: OrderTypeAsc,
				Order:         OrderTypeAsc,
			},
			args: args[int64]{
				value: -1,
			},
			want: false,
		},
		{
			name: "add non decreasing int64 keeps descendant order",
			fields: fields[int64]{
				last:          &zeroInt64,
				expectedOrder: OrderTypeDesc,
				Order:         OrderTypeDesc,
			},
			args: args[int64]{
				value: 0,
			},
			want: true,
		},
		{
			name: "add decreasing int64 keeps descendant order",
			fields: fields[int64]{
				last:          &zeroInt64,
				expectedOrder: OrderTypeDesc,
				Order:         OrderTypeDesc,
			},
			args: args[int64]{
				value: -1,
			},
			want: true,
		},
		{
			name: "add increasing int64 violates descendant order",
			fields: fields[int64]{
				last:          &zeroInt64,
				expectedOrder: OrderTypeDesc,
				Order:         OrderTypeDesc,
			},
			args: args[int64]{
				value: 1,
			},
			want: false,
		},
		{
			name: "add increasing uint64 keeps ascendant order",
			fields: fields[uint64]{
				last:          &zeroUint64,
				expectedOrder: OrderTypeAsc,
				Order:         OrderTypeAsc,
			},
			args: args[uint64]{
				value: 1,
			},
			want: true,
		},
		{
			name: "add increasing float64 keeps ascendant order",
			fields: fields[float64]{
				last:          &float64Zero,
				expectedOrder: OrderTypeAsc,
				Order:         OrderTypeAsc,
			},
			args: args[float64]{
				value: 1,
			},
			want: true,
		},
		{
			name: "add increasing string keeps ascendant order",
			fields: fields[string]{
				last:          &stringZero,
				expectedOrder: OrderTypeAsc,
				Order:         OrderTypeAsc,
			},
			args: args[string]{
				value: "1",
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got bool
			switch tt.fields.(type) {
			case fields[int64]:
				ss := &SequenceState[int64]{
					last:          tt.fields.(fields[int64]).last,
					expectedOrder: tt.fields.(fields[int64]).expectedOrder,
					Order:         tt.fields.(fields[int64]).Order,
				}
				got = ss.Add(tt.args.(args[int64]).value)
			case fields[uint64]:
				ss := &SequenceState[uint64]{
					last:          tt.fields.(fields[uint64]).last,
					expectedOrder: tt.fields.(fields[uint64]).expectedOrder,
					Order:         tt.fields.(fields[uint64]).Order,
				}
				got = ss.Add(tt.args.(args[uint64]).value)
			case fields[float64]:
				ss := &SequenceState[float64]{
					last:          tt.fields.(fields[float64]).last,
					expectedOrder: tt.fields.(fields[float64]).expectedOrder,
					Order:         tt.fields.(fields[float64]).Order,
				}
				got = ss.Add(tt.args.(args[float64]).value)
			case fields[string]:
				ss := &SequenceState[string]{
					last:          tt.fields.(fields[string]).last,
					expectedOrder: tt.fields.(fields[string]).expectedOrder,
					Order:         tt.fields.(fields[string]).Order,
				}
				got = ss.Add(tt.args.(args[string]).value)
			default:
				t.Error("unknown type")
				t.FailNow()
			}

			if got != tt.want {
				t.Errorf("SequenceState.Add() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCompleteState_Add(t *testing.T) {
	var oneInt64 int64 = 1

	type fields[T constraints.Ordered] struct {
		next        *T
		step        T
		AllComplete bool
	}
	type args[T constraints.Ordered] struct {
		value T
	}
	tests := []struct {
		name   string
		fields any
		args   any
		want   bool
	}{
		{
			name: "add first int64 keeps completness",
			fields: fields[int64]{
				next:        nil,
				step:        1,
				AllComplete: true,
			},
			args: args[int64]{
				value: 0,
			},
			want: true,
		},
		{
			name: "add expected value keeps completness",
			fields: fields[int64]{
				next:        &oneInt64,
				step:        1,
				AllComplete: true,
			},
			args: args[int64]{
				value: 1,
			},
			want: true,
		},
		{
			name: "add increasing int64 keeps ascendant order",
			fields: fields[int64]{
				next:        &oneInt64,
				step:        1,
				AllComplete: true,
			},
			args: args[int64]{
				value: 1000000,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got bool
			switch tt.fields.(type) {
			case fields[int64]:
				ss := &CompleteState[int64]{
					next:        tt.fields.(fields[int64]).next,
					step:        tt.fields.(fields[int64]).step,
					AllComplete: tt.fields.(fields[int64]).AllComplete,
				}
				got = ss.Add(tt.args.(args[int64]).value)
			case fields[uint64]:
				ss := &CompleteState[uint64]{
					next:        tt.fields.(fields[uint64]).next,
					step:        tt.fields.(fields[uint64]).step,
					AllComplete: tt.fields.(fields[uint64]).AllComplete,
				}
				got = ss.Add(tt.args.(args[uint64]).value)
			case fields[float64]:
				ss := &CompleteState[float64]{
					next:        tt.fields.(fields[float64]).next,
					step:        tt.fields.(fields[float64]).step,
					AllComplete: tt.fields.(fields[float64]).AllComplete,
				}
				got = ss.Add(tt.args.(args[float64]).value)
			default:
				t.Error("unknown type")
				t.FailNow()
			}

			if got != tt.want {
				t.Errorf("CompleteState.Add() = %v, want %v", got, tt.want)
			}
		})
	}
}
