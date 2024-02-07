package function

import (
	"reflect"
	"testing"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types"
)

func TestFunctions(t *testing.T) {
	type celEnvOpts struct {
		function  cel.EnvOption
		variables []cel.EnvOption
	}
	type args struct {
		expression string
		variables  map[string]any
	}
	tests := []struct {
		name       string
		celEnvOpts celEnvOpts
		args       args
		want       interface{}
	}{
		{
			name: "sequence with int and string order",
			celEnvOpts: celEnvOpts{
				function:  MakeSequenceInt(),
				variables: []cel.EnvOption{SequenceStatefulFunctionEnv, cel.Variable("state1", SequenceStatefulFunctionType)},
			},
			args: args{
				expression: `sequence(1, "asc", state1)`,
				variables:  map[string]any{"state1": &SequenceStatefulFunction{Parameters: &SequenceParameters{Order: OrderTypeAsc}, State: &SequenceState{}}},
			},
			want: types.Bool(true),
		},
		{
			name: "sequence with uint and string order",
			celEnvOpts: celEnvOpts{
				function:  MakeSequenceUint(),
				variables: []cel.EnvOption{SequenceStatefulFunctionEnv, cel.Variable("state1", SequenceStatefulFunctionType)},
			},
			args: args{
				expression: `sequence(1u, "asc", state1)`,
				variables:  map[string]any{"state1": &SequenceStatefulFunction{Parameters: &SequenceParameters{Order: OrderTypeAsc}, State: &SequenceState{}}},
			},
			want: types.Bool(true),
		},
		{
			name: "sequence with double and string order returns false",
			celEnvOpts: celEnvOpts{
				function:  MakeSequenceFloat64(),
				variables: []cel.EnvOption{SequenceStatefulFunctionEnv, cel.Variable("state1", SequenceStatefulFunctionType)},
			},
			args: args{
				expression: `sequence(1.0, "asc", state1)`,
				variables:  map[string]any{"state1": &SequenceStatefulFunction{Parameters: &SequenceParameters{Order: OrderTypeAsc}, State: &SequenceState{ofFloat64: &SequenceStateOf[float64]{Last: func() *float64 { f := 1000.0; return &f }(), ExpectedOrder: OrderTypeAsc}}}},
			},
			want: types.Bool(false),
		},
		{
			name: "sequence with string and string order",
			celEnvOpts: celEnvOpts{
				function:  MakeSequenceString(),
				variables: []cel.EnvOption{SequenceStatefulFunctionEnv, cel.Variable("state1", SequenceStatefulFunctionType)},
			},
			args: args{
				expression: `sequence("1", "asc", state1)`,
				variables:  map[string]any{"state1": &SequenceStatefulFunction{Parameters: &SequenceParameters{Order: OrderTypeAsc}, State: &SequenceState{}}},
			},
			want: types.Bool(true),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			// Create environment
			envOptions := append(tt.celEnvOpts.variables, tt.celEnvOpts.function)
			celEnv, err := cel.NewEnv(envOptions...)
			if err != nil {
				t.Errorf("environment creation error: %v", err)
			}

			// Check iss for error in both Parse and Check.
			ast, iss := celEnv.Compile(tt.args.expression)
			if iss != nil && iss.Err() != nil {
				t.Errorf("Expression compilation produced some issues; %v", iss.Err())
			}

			prg, err := celEnv.Program(ast)
			if err != nil {
				t.Errorf("Program creation error: %v", err)
			}

			// Evaluate expression
			got, _, err := prg.Eval(tt.args.variables)
			if err != nil {
				t.Errorf("Expression evaulation error: %v", err)
			}

			// Check output
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("TestFunctions() = %v, want %v", got, tt.want)
			}
		})
	}
}
