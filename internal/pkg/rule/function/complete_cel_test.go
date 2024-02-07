package function

import (
	"reflect"
	"testing"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types"
)

func TestCompleteCELFunctions(t *testing.T) {
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
		// complete
		{
			name: "complete with int and int step",
			celEnvOpts: celEnvOpts{
				function:  MakeCompleteInt(),
				variables: []cel.EnvOption{CompleteStatefulFunctionEnv, cel.Variable("state1", CompleteStatefulFunctionType)},
			},
			args: args{
				expression: `complete(1, 1, state1)`,
				variables:  map[string]any{"state1": &CompleteStatefulFunction{Parameters: &CompleteParameters{Step: 1}, State: &CompleteState{}}},
			},
			want: types.Bool(true),
		},
		{
			name: "complete with uint and int step",
			celEnvOpts: celEnvOpts{
				function:  MakeCompleteUint(),
				variables: []cel.EnvOption{CompleteStatefulFunctionEnv, cel.Variable("state1", CompleteStatefulFunctionType)},
			},
			args: args{
				expression: `complete(1u, 1, state1)`,
				variables:  map[string]any{"state1": &CompleteStatefulFunction{Parameters: &CompleteParameters{Step: 1}, State: &CompleteState{}}},
			},
			want: types.Bool(true),
		},
		{
			name: "complete with double, double step",
			celEnvOpts: celEnvOpts{
				function:  MakeCompleteFloat64(),
				variables: []cel.EnvOption{CompleteStatefulFunctionEnv, cel.Variable("state1", CompleteStatefulFunctionType)},
			},
			args: args{
				expression: `complete(1.0, 1.0, state1)`,
				variables:  map[string]any{"state1": &CompleteStatefulFunction{Parameters: &CompleteParameters{Step: 1}, State: &CompleteState{}}},
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
