package filter

import (
	"reflect"
	"testing"
)

func TestNew(t *testing.T) {
	type args struct {
		config *Config
	}
	tests := []struct {
		name    string
		args    args
		want    *Filter
		wantErr bool
	}{
		{
			name: "Fail when allowlist and errodenylistrlist are provided",
			args: args{
				config: &Config{
					Allowlist: Predicates{NewString("string1")},
					Denylist:  Predicates{NewString("string2")},
				},
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := New(tt.args.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("New() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFilter_Evaluate(t *testing.T) {
	type fields struct {
		predicates []Predicate
		evalFunc   func(predicates []Predicate, element string) bool
	}
	type args struct {
		element string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "Evaluate element matching allowlist",
			fields: fields{
				predicates: Predicates{NewString("string1")},
				evalFunc:   allowlistEvalFunc,
			},
			args: args{
				element: "string1",
			},
			want: true,
		},
		{
			name: "Evaluate element NOT matching allowlist",
			fields: fields{
				predicates: Predicates{NewString("string1")},
				evalFunc:   allowlistEvalFunc,
			},
			args: args{
				element: "string2",
			},
			want: false,
		},
		{
			name: "Evaluate element matching denylist",
			fields: fields{
				predicates: Predicates{NewString("string1")},
				evalFunc:   denylistEvalFunc,
			},
			args: args{
				element: "string1",
			},
			want: false,
		},
		{
			name: "Evaluate element NOT matching denylist",
			fields: fields{
				predicates: Predicates{NewString("string1")},
				evalFunc:   denylistEvalFunc,
			},
			args: args{
				element: "string2",
			},
			want: true,
		},
		{
			name: "Evaluate element without allowlist neither denylist",
			fields: fields{
				evalFunc: trueEvalFunc,
			},
			args: args{
				element: "string1",
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &Filter{
				predicates: tt.fields.predicates,
				evalFunc:   tt.fields.evalFunc,
			}
			if got := f.Evaluate(tt.args.element); got != tt.want {
				t.Errorf("Filter.Evaluate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFilter_EvaluateList(t *testing.T) {
	type fields struct {
		predicates []Predicate
		evalFunc   func(predicates []Predicate, element string) bool
	}
	type args struct {
		elements []string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   []string
	}{
		{
			name: "Evaluate list using allowlist",
			fields: fields{
				predicates: Predicates{NewString("string1")},
				evalFunc:   allowlistEvalFunc,
			},
			args: args{
				elements: []string{"string0", "string1", "string2"},
			},
			want: []string{"string1"},
		},
		{
			name: "Evaluate list using denylist",
			fields: fields{
				predicates: Predicates{NewString("string1")},
				evalFunc:   denylistEvalFunc,
			},
			args: args{
				elements: []string{"string0", "string1", "string2"},
			},
			want: []string{"string0", "string2"},
		},
		{
			name: "Evaluate list without allowlist neither denylist",
			fields: fields{
				predicates: Predicates{NewString("string1")},
				evalFunc:   trueEvalFunc,
			},
			args: args{
				elements: []string{"string0", "string1", "string2"},
			},
			want: []string{"string0", "string1", "string2"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &Filter{
				predicates: tt.fields.predicates,
				evalFunc:   tt.fields.evalFunc,
			}
			if got := f.EvaluateList(tt.args.elements); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Filter.EvaluateList() = %v, want %v", got, tt.want)
			}
		})
	}
}
