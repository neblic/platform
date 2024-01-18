package rule

import (
	"reflect"
	"testing"

	"github.com/google/cel-go/cel"
	"github.com/neblic/platform/sampler/sample"
	expr "google.golang.org/genproto/googleapis/api/expr/v1alpha1"
)

func compileExpression(t *testing.T, expression string) *expr.Expr {
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

	return expr.Expr
}

func Test_ParseStatefulFunctions(t *testing.T) {
	type args struct {
		statefulFunctions []StatefulFunction
		expr              *expr.Expr
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
				statefulFunctions: []StatefulFunction{&SequenceStatefulFunction{}, &CompleteStatefulFunction{}},
				expr:              compileExpression(t, `sequence("foo", "asc")`),
			},
			want:    []StatefulFunction{&SequenceStatefulFunction{enabled: true, Order: OrderTypeAsc}, &CompleteStatefulFunction{}},
			wantErr: false,
		},
		{
			name: "sequence function with int field",
			args: args{
				statefulFunctions: []StatefulFunction{&SequenceStatefulFunction{}, &CompleteStatefulFunction{}},
				expr:              compileExpression(t, `sequence(1, "asc")`),
			},
			want:    []StatefulFunction{&SequenceStatefulFunction{enabled: true, Order: OrderTypeAsc}, &CompleteStatefulFunction{}},
			wantErr: false,
		},
		{
			name: "sequence function with uint field",
			args: args{
				statefulFunctions: []StatefulFunction{&SequenceStatefulFunction{}, &CompleteStatefulFunction{}},
				expr:              compileExpression(t, `sequence(1u, "asc")`),
			},
			want:    []StatefulFunction{&SequenceStatefulFunction{enabled: true, Order: OrderTypeAsc}, &CompleteStatefulFunction{}},
			wantErr: false,
		},
		{
			name: "sequence function with float field",
			args: args{
				statefulFunctions: []StatefulFunction{&SequenceStatefulFunction{}, &CompleteStatefulFunction{}},
				expr:              compileExpression(t, `sequence(1.0, "asc")`),
			},
			want:    []StatefulFunction{&SequenceStatefulFunction{enabled: true, Order: OrderTypeAsc}, &CompleteStatefulFunction{}},
			wantErr: false,
		},
		{
			name: "complete function with int field",
			args: args{
				statefulFunctions: []StatefulFunction{&SequenceStatefulFunction{}, &CompleteStatefulFunction{}},
				expr:              compileExpression(t, `complete(1, 1)`),
			},
			want:    []StatefulFunction{&SequenceStatefulFunction{}, &CompleteStatefulFunction{enabled: true, Step: 1}},
			wantErr: false,
		},
		{
			name: "complete function with uint field",
			args: args{
				statefulFunctions: []StatefulFunction{&SequenceStatefulFunction{}, &CompleteStatefulFunction{}},
				expr:              compileExpression(t, `complete(1u, 1)`),
			},
			want:    []StatefulFunction{&SequenceStatefulFunction{}, &CompleteStatefulFunction{enabled: true, Step: 1}},
			wantErr: false,
		},
		{
			name: "sequence function with float field",
			args: args{
				statefulFunctions: []StatefulFunction{&SequenceStatefulFunction{}, &CompleteStatefulFunction{}},
				expr:              compileExpression(t, `complete(1.0, 1.0)`),
			},
			want:    []StatefulFunction{&SequenceStatefulFunction{}, &CompleteStatefulFunction{enabled: true, Step: 1}},
			wantErr: false,
		},
		{
			name: "stateful function not in the root of the expression",
			args: args{
				statefulFunctions: []StatefulFunction{&SequenceStatefulFunction{}, &CompleteStatefulFunction{}},
				expr:              compileExpression(t, `sequence("foo", "asc") && "bar" == "bar"`),
			},
			want:    []StatefulFunction{&SequenceStatefulFunction{enabled: true, Order: OrderTypeAsc}, &CompleteStatefulFunction{}},
			wantErr: false,
		},
		{
			name: "multiple stateful functions returns error",
			args: args{
				statefulFunctions: []StatefulFunction{&SequenceStatefulFunction{}, &CompleteStatefulFunction{}},
				expr:              compileExpression(t, `sequence("foo", "asc") && sequence("bar", "asc")`),
			},
			want:    []StatefulFunction{&SequenceStatefulFunction{enabled: true, Order: OrderTypeAsc}, &CompleteStatefulFunction{}},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ParseStatefulFunctions(tt.args.statefulFunctions, tt.args.expr)
			if (err != nil) != tt.wantErr {
				t.Errorf("Builder.Build() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(tt.args.statefulFunctions, tt.want) {
				t.Errorf("ParseStatefulFunctions() = %v, want %v", tt.args.statefulFunctions, tt.want)
			}
		})
	}
}

func TestBuilder_Build(t *testing.T) {
	type fields struct {
		schema sample.Schema
	}
	type args struct {
		rule string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "build rule with abs function",
			fields: fields{
				schema: sample.NewDynamicSchema(),
			},
			args: args{
				rule: `abs(-1) == 1`,
			},
			wantErr: false,
		},
		{
			name: "build rule with sequence function",
			fields: fields{
				schema: sample.NewDynamicSchema(),
			},
			args: args{
				rule: `sequence(-1, "asc")`,
			},
			wantErr: false,
		},
		{
			name: "build rule with complete function",
			fields: fields{
				schema: sample.NewDynamicSchema(),
			},
			args: args{
				rule: `complete(0, 1)`,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rb, err := NewBuilder(tt.fields.schema, CheckFunctions)
			if err != nil {
				t.Errorf("Builder.NewBuilder() error = %v", err)
				return
			}
			_, err = rb.Build(tt.args.rule)
			if (err != nil) != tt.wantErr {
				t.Errorf("Builder.Build() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
