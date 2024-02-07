package function

import (
	"reflect"
	"testing"
	"time"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types"
)

func TestCELFunctions(t *testing.T) {
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
		//abs
		{
			name: "abs_double",
			celEnvOpts: celEnvOpts{
				function:  AbsDouble,
				variables: []cel.EnvOption{},
			},
			args: args{
				expression: `abs(-1.23)`,
				variables:  map[string]any{},
			},
			want: types.Double(1.23),
		},
		{
			name: "abs_int",
			celEnvOpts: celEnvOpts{
				function:  AbsInt,
				variables: []cel.EnvOption{},
			},
			args: args{
				expression: `abs(-1)`,
				variables:  map[string]any{},
			},
			want: types.Int(1),
		},

		// now
		{
			name: "now",
			celEnvOpts: celEnvOpts{
				function:  Now,
				variables: []cel.EnvOption{},
			},
			args: args{
				expression: `now().getDayOfYear() + 1`,
				variables:  map[string]any{},
			},
			want: types.Int(time.Now().YearDay()),
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
