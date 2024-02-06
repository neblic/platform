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
				function:  makeSequenceInt(),
				variables: []cel.EnvOption{sequenceStatefulFunctionEnv, cel.Variable("state1", sequenceStatefulFunctionType)},
			},
			args: args{
				expression: `sequence(1, "asc", state1)`,
				variables:  map[string]any{"state1": &SequenceStatefulFunction{}},
			},
			want: types.Bool(true),
		},
		{
			name: "sequence with uint and string order",
			celEnvOpts: celEnvOpts{
				function:  makeSequenceUint(),
				variables: []cel.EnvOption{sequenceStatefulFunctionEnv, cel.Variable("state1", sequenceStatefulFunctionType)},
			},
			args: args{
				expression: `sequence(1u, "asc", state1)`,
				variables:  map[string]any{"state1": &SequenceStatefulFunction{}},
			},
			want: types.Bool(true),
		},
		{
			name: "sequence with double and string order returns false",
			celEnvOpts: celEnvOpts{
				function:  makeSequenceFloat64(),
				variables: []cel.EnvOption{sequenceStatefulFunctionEnv, cel.Variable("state1", sequenceStatefulFunctionType)},
			},
			args: args{
				expression: `sequence(1.0, "asc", state1)`,
				variables:  map[string]any{"state1": &SequenceStatefulFunction{ofFloat64: &SequenceStatefulFunctionOf[float64]{last: func() *float64 { f := 1000.0; return &f }(), expectedOrder: OrderTypeAsc}}},
			},
			want: types.Bool(false),
		},
		{
			name: "sequence with string and string order",
			celEnvOpts: celEnvOpts{
				function:  makeSequenceString(),
				variables: []cel.EnvOption{sequenceStatefulFunctionEnv, cel.Variable("state1", sequenceStatefulFunctionType)},
			},
			args: args{
				expression: `sequence("1", "asc", state1)`,
				variables:  map[string]any{"state1": &SequenceStatefulFunction{}},
			},
			want: types.Bool(true),
		},
		// complete
		{
			name: "complete with int and int step",
			celEnvOpts: celEnvOpts{
				function:  makeCompleteInt(),
				variables: []cel.EnvOption{completeStatefulFunctionEnv, cel.Variable("state1", completeStatefulFunctionType)},
			},
			args: args{
				expression: `complete(1, 1, state1)`,
				variables:  map[string]any{"state1": &CompleteStatefulFunction{}},
			},
			want: types.Bool(true),
		},
		{
			name: "complete with uint and int step",
			celEnvOpts: celEnvOpts{
				function:  makeCompleteUint(),
				variables: []cel.EnvOption{completeStatefulFunctionEnv, cel.Variable("state1", completeStatefulFunctionType)},
			},
			args: args{
				expression: `complete(1u, 1, state1)`,
				variables:  map[string]any{"state1": &CompleteStatefulFunction{}},
			},
			want: types.Bool(true),
		},
		{
			name: "complete with double, double step",
			celEnvOpts: celEnvOpts{
				function:  makeCompleteFloat64(),
				variables: []cel.EnvOption{completeStatefulFunctionEnv, cel.Variable("state1", completeStatefulFunctionType)},
			},
			args: args{
				expression: `complete(1.0, 1.0, state1)`,
				variables:  map[string]any{"state1": &CompleteStatefulFunction{}},
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
