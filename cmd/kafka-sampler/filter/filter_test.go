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
			name: "Fail when allow and deny are provided",
			args: args{
				config: &Config{
					Allow: NewString("string1"),
					Deny:  NewString("string2"),
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
		predicate Predicate
		evalFunc  func(predicate Predicate, element string) bool
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
			name: "Evaluate element matching allow",
			fields: fields{
				predicate: NewString("string1"),
				evalFunc:  allowEvalFunc,
			},
			args: args{
				element: "string1",
			},
			want: true,
		},
		{
			name: "Evaluate element NOT matching allow",
			fields: fields{
				predicate: NewString("string1"),
				evalFunc:  allowEvalFunc,
			},
			args: args{
				element: "string2",
			},
			want: false,
		},
		{
			name: "Evaluate element matching deny",
			fields: fields{
				predicate: NewString("string1"),
				evalFunc:  denyEvalFunc,
			},
			args: args{
				element: "string1",
			},
			want: false,
		},
		{
			name: "Evaluate element NOT matching deny",
			fields: fields{
				predicate: NewString("string1"),
				evalFunc:  denyEvalFunc,
			},
			args: args{
				element: "string2",
			},
			want: true,
		},
		{
			name: "Evaluate element without allow neither deny",
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
				predicate: tt.fields.predicate,
				evalFunc:  tt.fields.evalFunc,
			}
			if got := f.Evaluate(tt.args.element); got != tt.want {
				t.Errorf("Filter.Evaluate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFilter_EvaluateList(t *testing.T) {
	type fields struct {
		predicate Predicate
		evalFunc  func(predicate Predicate, element string) bool
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
			name: "Evaluate list using allow",
			fields: fields{
				predicate: NewString("string1"),
				evalFunc:  allowEvalFunc,
			},
			args: args{
				elements: []string{"string0", "string1", "string2"},
			},
			want: []string{"string1"},
		},
		{
			name: "Evaluate list using deny",
			fields: fields{
				predicate: NewString("string1"),
				evalFunc:  denyEvalFunc,
			},
			args: args{
				elements: []string{"string0", "string1", "string2"},
			},
			want: []string{"string0", "string2"},
		},
		{
			name: "Evaluate list without allow neither deny",
			fields: fields{
				predicate: NewString("string1"),
				evalFunc:  trueEvalFunc,
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
				predicate: tt.fields.predicate,
				evalFunc:  tt.fields.evalFunc,
			}
			if got := f.EvaluateList(tt.args.elements); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Filter.EvaluateList() = %v, want %v", got, tt.want)
			}
		})
	}
}
