package rule

import (
	"reflect"
	"testing"
	"time"

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
		//abs
		{
			name: "abs_double",
			celEnvOpts: celEnvOpts{
				function:  absDoubleFunc,
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
				function:  absIntFunc,
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
				function:  nowFunc,
				variables: []cel.EnvOption{},
			},
			args: args{
				expression: `now().getDayOfYear() + 1`,
				variables:  map[string]any{},
			},
			want: types.Int(time.Now().YearDay()),
		},

		// sequence
		{
			name: "sequence with int and string order",
			celEnvOpts: celEnvOpts{
				function:  makeSequenceInt(NewStateProvider()),
				variables: []cel.EnvOption{},
			},
			args: args{
				expression: `sequence(1, "asc")`,
				variables:  map[string]any{},
			},
			want: types.Bool(true),
		},
		{
			name: "sequence with uint and string order",
			celEnvOpts: celEnvOpts{
				function:  makeSequenceUint(NewStateProvider()),
				variables: []cel.EnvOption{},
			},
			args: args{
				expression: `sequence(1u, "asc")`,
				variables:  map[string]any{},
			},
			want: types.Bool(true),
		},
		{
			name: "sequence with double and string order",
			celEnvOpts: celEnvOpts{
				function:  makeSequenceFloat64(NewStateProvider()),
				variables: []cel.EnvOption{},
			},
			args: args{
				expression: `sequence(1.0, "asc")`,
				variables:  map[string]any{},
			},
			want: types.Bool(true),
		},
		{
			name: "sequence with string and string order",
			celEnvOpts: celEnvOpts{
				function:  makeSequenceString(NewStateProvider()),
				variables: []cel.EnvOption{},
			},
			args: args{
				expression: `sequence("1", "asc")`,
				variables:  map[string]any{},
			},
			want: types.Bool(true),
		},
		// complete
		{
			name: "complete with int and int step",
			celEnvOpts: celEnvOpts{
				function:  makeCompleteInt(NewStateProvider()),
				variables: []cel.EnvOption{},
			},
			args: args{
				expression: `complete(1, 1)`,
				variables:  map[string]any{},
			},
			want: types.Bool(true),
		},
		{
			name: "complete with uint and int step",
			celEnvOpts: celEnvOpts{
				function:  makeCompleteUint(NewStateProvider()),
				variables: []cel.EnvOption{},
			},
			args: args{
				expression: `complete(1u, 1)`,
				variables:  map[string]any{},
			},
			want: types.Bool(true),
		},
		{
			name: "complete with double, double step",
			celEnvOpts: celEnvOpts{
				function:  makeCompleteFloat64(NewStateProvider()),
				variables: []cel.EnvOption{},
			},
			args: args{
				expression: `complete(1.0, 1.0)`,
				variables:  map[string]any{},
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
			if iss.Err() != nil {
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
