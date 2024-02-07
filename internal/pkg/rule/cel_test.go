package rule

import (
	"reflect"
	"testing"

	"github.com/google/cel-go/cel"
	"github.com/neblic/platform/internal/pkg/rule/function"
	expr "google.golang.org/genproto/googleapis/api/expr/v1alpha1"
)

func compileExpression(t *testing.T, expression string) *expr.CheckedExpr {
	env, err := cel.NewEnv(CheckFunctionsEnvOptions...)
	if err != nil {
		t.Fatal(err)
	}

	ast, iss := env.Compile(expression)
	if iss != nil && iss.Err() != nil {
		t.Fatal(iss.Err())
	}

	expr, err := cel.AstToCheckedExpr(ast)
	if err != nil {
		t.Fatal(err)
	}

	return expr
}

func TestCheckedExprModifier_InjectState(t *testing.T) {
	type args struct {
		expr *expr.CheckedExpr
	}
	type want struct {
		stateNames        []string
		statefulFunctions []function.StatefulFunction
		err               bool
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "ascendant sequence function with string field",
			args: args{
				expr: compileExpression(t, `sequence("foo", "asc")`),
			},
			want: want{
				stateNames: []string{"state0"},
				statefulFunctions: []function.StatefulFunction{
					&function.SequenceStatefulFunction{Parameters: &function.SequenceParameters{Order: function.OrderTypeAsc}, State: &function.SequenceState{}},
				},
				err: false,
			},
		},
		{
			name: "descendant sequence function with int field",
			args: args{
				expr: compileExpression(t, `sequence(1, "desc")`),
			},
			want: want{
				stateNames: []string{"state0"},
				statefulFunctions: []function.StatefulFunction{
					&function.SequenceStatefulFunction{Parameters: &function.SequenceParameters{Order: function.OrderTypeDesc}, State: &function.SequenceState{}},
				},
				err: false,
			},
		},
		{
			name: "ascendant sequence function with uint field",
			args: args{
				expr: compileExpression(t, `sequence(1u, "asc")`),
			},
			want: want{
				stateNames: []string{"state0"},
				statefulFunctions: []function.StatefulFunction{
					&function.SequenceStatefulFunction{Parameters: &function.SequenceParameters{Order: function.OrderTypeAsc}, State: &function.SequenceState{}},
				},
				err: false,
			},
		},
		{
			name: "ascendant sequence function with float field",
			args: args{
				expr: compileExpression(t, `sequence(1.0, "asc")`),
			},
			want: want{
				stateNames: []string{"state0"},
				statefulFunctions: []function.StatefulFunction{
					&function.SequenceStatefulFunction{Parameters: &function.SequenceParameters{Order: function.OrderTypeAsc}, State: &function.SequenceState{}},
				},
				err: false,
			},
		},
		{
			name: "complete function with int field",
			args: args{
				expr: compileExpression(t, `complete(1, 1)`),
			},
			want: want{
				stateNames: []string{"state0"},
				statefulFunctions: []function.StatefulFunction{
					&function.CompleteStatefulFunction{Parameters: &function.CompleteParameters{Step: 1}, State: &function.CompleteState{}},
				},
				err: false,
			},
		},
		{
			name: "complete function with uint field",
			args: args{
				expr: compileExpression(t, `complete(1u, 1)`),
			},
			want: want{
				stateNames: []string{"state0"},
				statefulFunctions: []function.StatefulFunction{
					&function.CompleteStatefulFunction{Parameters: &function.CompleteParameters{Step: 1}, State: &function.CompleteState{}},
				},
				err: false,
			},
		},
		{
			name: "complete function with float field",
			args: args{
				expr: compileExpression(t, `complete(1.0, 1.0)`),
			},
			want: want{
				stateNames: []string{"state0"},
				statefulFunctions: []function.StatefulFunction{
					&function.CompleteStatefulFunction{Parameters: &function.CompleteParameters{Step: 1}, State: &function.CompleteState{}},
				},
				err: false,
			},
		},
		{
			name: "stateful function not in the root of the expression",
			args: args{
				expr: compileExpression(t, `sequence("foo", "asc") && "bar" == "bar"`),
			},
			want: want{
				stateNames: []string{"state0"},
				statefulFunctions: []function.StatefulFunction{
					&function.SequenceStatefulFunction{Parameters: &function.SequenceParameters{Order: function.OrderTypeAsc}, State: &function.SequenceState{}},
				},
				err: false,
			},
		},
		{
			name: "multiple stateful functions",
			args: args{
				expr: compileExpression(t, `sequence("foo", "asc") && sequence(1, "desc")`),
			},
			want: want{
				stateNames: []string{"state0", "state1"},
				statefulFunctions: []function.StatefulFunction{
					&function.SequenceStatefulFunction{Parameters: &function.SequenceParameters{Order: function.OrderTypeAsc}, State: &function.SequenceState{}},
					&function.SequenceStatefulFunction{Parameters: &function.SequenceParameters{Order: function.OrderTypeDesc}, State: &function.SequenceState{}},
				},
				err: false,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checkedExprModifier := NewCheckedExprModifier(tt.args.expr)
			got, err := checkedExprModifier.InjectState()
			if (err != nil) != tt.want.err {
				t.Errorf("CheckedExprModifier.InjectState() error = %v, wantErr %v", err, tt.want.err)
				return
			}

			if len(got) != len(tt.want.stateNames) {
				t.Errorf("CheckedExprModifier.InjectState() state names length = %v, want %v", len(got), len(tt.want.stateNames))
				return
			}

			if len(got) != len(tt.want.statefulFunctions) {
				t.Errorf("CheckedExprModifier.InjectState() stateful functions length = %v, want %v", len(got), len(tt.want.statefulFunctions))
				return
			}

			for i, provider := range got {
				gotStateName := provider.StateName
				if gotStateName != tt.want.stateNames[i] {
					t.Errorf("CheckedExprModifier.InjectState() state name = %v, want %v", gotStateName, tt.want.stateNames)
				}

				gotStatefulFunction := provider.GlobalStatefulFunction()
				if !reflect.DeepEqual(gotStatefulFunction, tt.want.statefulFunctions[i]) {
					t.Errorf("CheckedExprModifier.InjectState() stateful function = %v, want %v", gotStatefulFunction, tt.want.statefulFunctions)
				}
			}
		})
	}
}
