package rule

import (
	"reflect"
	"testing"

	"github.com/google/cel-go/cel"
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
	tests := []struct {
		name    string
		args    args
		want    []StatefulFunction
		wantErr bool
	}{
		{
			name: "sequence function with string field",
			args: args{
				expr: compileExpression(t, `sequence("foo", "asc")`),
			},
			want:    []StatefulFunction{&SequenceStatefulFunction{stateName: "state0", order: OrderTypeAsc}},
			wantErr: false,
		},
		{
			name: "sequence function with int field",
			args: args{
				expr: compileExpression(t, `sequence(1, "asc")`),
			},
			want:    []StatefulFunction{&SequenceStatefulFunction{stateName: "state0", order: OrderTypeAsc}},
			wantErr: false,
		},
		{
			name: "sequence function with uint field",
			args: args{
				expr: compileExpression(t, `sequence(1u, "asc")`),
			},
			want:    []StatefulFunction{&SequenceStatefulFunction{stateName: "state0", order: OrderTypeAsc}},
			wantErr: false,
		},
		{
			name: "sequence function with float field",
			args: args{
				expr: compileExpression(t, `sequence(1.0, "asc")`),
			},
			want:    []StatefulFunction{&SequenceStatefulFunction{stateName: "state0", order: OrderTypeAsc}},
			wantErr: false,
		},
		{
			name: "complete function with int field",
			args: args{
				expr: compileExpression(t, `complete(1, 1)`),
			},
			want:    []StatefulFunction{&CompleteStatefulFunction{stateName: "state0", step: 1}},
			wantErr: false,
		},
		{
			name: "complete function with uint field",
			args: args{
				expr: compileExpression(t, `complete(1u, 1)`),
			},
			want:    []StatefulFunction{&CompleteStatefulFunction{stateName: "state0", step: 1}},
			wantErr: false,
		},
		{
			name: "sequence function with float field",
			args: args{
				expr: compileExpression(t, `complete(1.0, 1.0)`),
			},
			want:    []StatefulFunction{&CompleteStatefulFunction{stateName: "state0", step: 1}},
			wantErr: false,
		},
		{
			name: "stateful function not in the root of the expression",
			args: args{
				expr: compileExpression(t, `sequence("foo", "asc") && "bar" == "bar"`),
			},
			want:    []StatefulFunction{&SequenceStatefulFunction{stateName: "state0", order: OrderTypeAsc}},
			wantErr: false,
		},
		{
			name: "multiple stateful functions",
			args: args{
				expr: compileExpression(t, `sequence("foo", "asc") && sequence(1, "desc")`),
			},
			want:    []StatefulFunction{&SequenceStatefulFunction{stateName: "state0", order: OrderTypeAsc}, &SequenceStatefulFunction{stateName: "state1", order: OrderTypeDesc}},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checkedExprModifier := NewCheckedExprModifier(tt.args.expr)
			got, err := checkedExprModifier.InjectState()
			if (err != nil) != tt.wantErr {
				t.Errorf("CheckedExprModifier.InjectState() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CheckedExprModifier.InjectState() = %v, want %v", got, tt.want)
			}
		})
	}
}
