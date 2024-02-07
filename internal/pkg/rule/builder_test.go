package rule

import (
	"testing"

	"github.com/neblic/platform/sampler/sample"
)

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
