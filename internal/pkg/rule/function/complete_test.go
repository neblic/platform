package function

import (
	"testing"
)

func Test_CallComplete(t *testing.T) {
	oneInt64 := int64(1)

	type args[T Number] struct {
		state      *CompleteStateOf[T]
		parameters *CompleteParameters
		value      T
	}
	tests := []struct {
		name string
		args any
		want bool
	}{
		{
			name: "call function with expected value keeps completness",
			args: args[int64]{
				state: &CompleteStateOf[int64]{
					Next:        &oneInt64,
					Step:        1,
					AllComplete: true,
				},
				parameters: nil,
				value:      1,
			},
			want: true,
		},
		{
			name: "call function with unexpected value violates completness",
			args: args[int64]{
				state: &CompleteStateOf[int64]{
					Next:        &oneInt64,
					Step:        1,
					AllComplete: true,
				},
				parameters: nil,
				value:      1000000,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got bool
			switch v := tt.args.(type) {
			case args[int64]:
				got = CallComplete(v.state, v.parameters, v.value)
			case args[uint64]:
				got = CallComplete(v.state, v.parameters, v.value)
			case args[float64]:
				got = CallComplete(v.state, v.parameters, v.value)
			default:
				t.Error("unknown type")
				t.FailNow()
			}

			if got != tt.want {
				t.Errorf("CallComplete() = %v, want %v", got, tt.want)
			}
		})
	}
}
