package metric

import (
	"testing"
)

func TestPath_AddPart(t *testing.T) {
	type args struct {
		field     string
		fieldType FieldType
	}
	tests := []struct {
		name string
		p    Path
		args args
		want Path
	}{
		{
			name: "add number field to emtpy path",
			p:    NewPath(),
			args: args{
				field:     `$`,
				fieldType: NumberType,
			},
			want: Path(`["$":number]`),
		},
		{
			name: "add string field to emtpy path",
			p:    NewPath(),
			args: args{
				field:     `$`,
				fieldType: StringType,
			},
			want: Path(`["$":string]`),
		},
		{
			name: "add boolean field to emtpy path",
			p:    NewPath(),
			args: args{
				field:     `$`,
				fieldType: BooleanType,
			},
			want: Path(`["$":boolean]`),
		},
		{
			name: "add array field to emtpy path",
			p:    NewPath(),
			args: args{
				field:     `$`,
				fieldType: ArrayType,
			},
			want: Path(`["$":array]`),
		},
		{
			name: "add obj field to emtpy path",
			p:    NewPath(),
			args: args{
				field:     `$`,
				fieldType: ObjectType,
			},
			want: Path(`["$":obj]`),
		},
		{
			name: "add string field with \" to obj path",
			p:    `["$":obj]`,
			args: args{
				field:     `field"1"`,
				fieldType: NumberType,
			},
			want: Path(`["$":obj]["field\"1\"":number]`),
		},
		{
			name: "add string field with \\ to obj path",
			p:    `["$":obj]`,
			args: args{
				field:     `field\1`,
				fieldType: NumberType,
			},
			want: Path(`["$":obj]["field\\1":number]`),
		},
		{
			name: "add number field to obj path",
			p:    `["$":obj]`,
			args: args{
				field:     `field1`,
				fieldType: NumberType,
			},
			want: Path(`["$":obj]["field1":number]`),
		},
		{
			name: "add array field to obj path",
			p:    `["$":obj]`,
			args: args{
				field:     `*`,
				fieldType: ArrayType,
			},
			want: Path(`["$":obj]["*":array]`),
		},
		{
			name: "add obj field to array path",
			p:    `["$":obj]`,
			args: args{
				field:     `*`,
				fieldType: ArrayType,
			},
			want: Path(`["$":obj]["*":array]`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.p.AddPart(tt.args.field, tt.args.fieldType); got != tt.want {
				t.Errorf("Path.AddPart() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestJSONPath_IsEmpty(t *testing.T) {
	tests := []struct {
		name string
		p    Path
		want bool
	}{
		{
			name: "empty path",
			p:    NewPath(),
			want: true,
		},
		{
			name: "non-empty path",
			p:    `["$":boolean]`,
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.p.IsEmpty(); got != tt.want {
				t.Errorf("Path.IsEmpty() = %v, want %v", got, tt.want)
			}
		})
	}
}
