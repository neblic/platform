package digest

import (
	"math"
	"reflect"
	"testing"

	"github.com/neblic/platform/dataplane/protos"
	"github.com/neblic/platform/internal/pkg/data"
)

var trueBoolean = true
var falseBoolean = false

func TestValue_updateBoolean(t *testing.T) {
	type fields struct {
		maxProcessedFields int
		fieldsProcessed    int
	}
	type args struct {
		state   *protos.BooleanValue
		boolean *bool
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *protos.BooleanValue
		wantErr bool
	}{
		{
			name: "update nil boolean without state",
			fields: fields{
				maxProcessedFields: 100,
				fieldsProcessed:    0,
			},
			args: args{
				state:   protos.NewBooleanValue(),
				boolean: nil,
			},
			want: &protos.BooleanValue{
				TotalCount:   1,
				NullCount:    1,
				DefaultCount: 0,
				TotalFalse:   0,
				TotalTrue:    0,
			},
		},
		{
			name: "update default boolean without state",
			fields: fields{
				maxProcessedFields: 100,
				fieldsProcessed:    0,
			},
			args: args{
				state:   protos.NewBooleanValue(),
				boolean: &falseBoolean,
			},
			want: &protos.BooleanValue{
				TotalCount:   1,
				NullCount:    0,
				DefaultCount: 1,
				TotalFalse:   1,
				TotalTrue:    0,
			},
		},
		{
			name: "update non-default boolean without state",
			fields: fields{
				maxProcessedFields: 100,
				fieldsProcessed:    0,
			},
			args: args{
				state:   protos.NewBooleanValue(),
				boolean: &trueBoolean,
			},
			want: &protos.BooleanValue{
				TotalCount:   1,
				NullCount:    0,
				DefaultCount: 0,
				TotalFalse:   0,
				TotalTrue:    1,
			},
		},
		{
			name: "update default boolean with state",
			fields: fields{
				maxProcessedFields: 100,
				fieldsProcessed:    0,
			},
			args: args{
				state: &protos.BooleanValue{
					TotalCount:   10,
					NullCount:    8,
					DefaultCount: 1,
					TotalFalse:   1,
					TotalTrue:    1,
				},
				boolean: &falseBoolean,
			},
			want: &protos.BooleanValue{
				TotalCount:   11,
				NullCount:    8,
				DefaultCount: 2,
				TotalFalse:   2,
				TotalTrue:    1,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := &Value{
				maxProcessedFields: tt.fields.maxProcessedFields,
				fieldsProcessed:    tt.fields.fieldsProcessed,
			}
			got, err := v.updateBoolean(tt.args.state, tt.args.boolean)
			if (err != nil) != tt.wantErr {
				t.Errorf("Value.updateBoolean() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Value.updateBoolean() = %v, want %v", got, tt.want)
			}
		})
	}
}

var zeroNumber = float64(0.0)
var oneNumber = float64(1.0)

func TestValue_updateNum(t *testing.T) {
	type fields struct {
		maxProcessedFields int
		fieldsProcessed    int
	}
	type args struct {
		state  *protos.NumberValue
		number *float64
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *protos.NumberValue
		wantErr bool
	}{
		{
			name: "update nil number without state",
			fields: fields{
				maxProcessedFields: 100,
				fieldsProcessed:    0,
			},
			args: args{
				state:  protos.NewNumberValue(),
				number: nil,
			},
			want: &protos.NumberValue{
				TotalCount:   1,
				NullCount:    1,
				DefaultCount: 0,
				Min: &protos.MinStat{
					Value: math.Inf(+1),
				},
				Avg: &protos.AvgStat{
					Sum:   0.0,
					Count: 0,
				},
				Max: &protos.MaxStat{
					Value: math.Inf(-1),
				},
			},
		},
		{
			name: "update default number without state",
			fields: fields{
				maxProcessedFields: 100,
				fieldsProcessed:    0,
			},
			args: args{
				state:  protos.NewNumberValue(),
				number: &zeroNumber,
			},
			want: &protos.NumberValue{
				TotalCount:   1,
				NullCount:    0,
				DefaultCount: 1,
				Min: &protos.MinStat{
					Value: 0.0,
				},
				Avg: &protos.AvgStat{
					Sum:   0.0,
					Count: 1,
				},
				Max: &protos.MaxStat{
					Value: 0.0,
				},
			},
		},
		{
			name: "update non-default number without state",
			fields: fields{
				maxProcessedFields: 100,
				fieldsProcessed:    0,
			},
			args: args{
				state:  protos.NewNumberValue(),
				number: &oneNumber,
			},
			want: &protos.NumberValue{
				TotalCount:   1,
				NullCount:    0,
				DefaultCount: 0,
				Min: &protos.MinStat{
					Value: 1.0,
				},
				Avg: &protos.AvgStat{
					Sum:   1.0,
					Count: 1,
				},
				Max: &protos.MaxStat{
					Value: 1.0,
				},
			},
		},
		{
			name: "update nil number with state",
			fields: fields{
				maxProcessedFields: 100,
				fieldsProcessed:    0,
			},
			args: args{
				state: &protos.NumberValue{
					TotalCount:   10,
					NullCount:    9,
					DefaultCount: 0,
					Min: &protos.MinStat{
						Value: 1.0,
					},
					Avg: &protos.AvgStat{
						Sum:   1.0,
						Count: 1,
					},
					Max: &protos.MaxStat{
						Value: 1.0,
					},
				},
				number: nil,
			},
			want: &protos.NumberValue{
				TotalCount:   11,
				NullCount:    10,
				DefaultCount: 0,
				Min: &protos.MinStat{
					Value: 1.0,
				},
				Avg: &protos.AvgStat{
					Sum:   1.0,
					Count: 1,
				},
				Max: &protos.MaxStat{
					Value: 1.0,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := &Value{
				maxProcessedFields: tt.fields.maxProcessedFields,
				fieldsProcessed:    tt.fields.fieldsProcessed,
			}
			got, err := v.updateNum(tt.args.state, tt.args.number)
			if (err != nil) != tt.wantErr {
				t.Errorf("Value.updateNum() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Value.updateNum() = %v, want %v", got, tt.want)
			}
		})
	}
}

var emptyString = ""
var somethingString = "something"

func TestValue_updateString(t *testing.T) {
	type fields struct {
		maxProcessedFields int
		fieldsProcessed    int
	}
	type args struct {
		state *protos.StringValue
		str   *string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *protos.StringValue
		wantErr bool
	}{
		{
			name: "update nil string without state",
			fields: fields{
				maxProcessedFields: 100,
				fieldsProcessed:    0,
			},
			args: args{
				state: protos.NewStringValue(),
				str:   nil,
			},
			want: &protos.StringValue{
				TotalCount:   1,
				NullCount:    1,
				DefaultCount: 0,
				Length: &protos.NumberStat{
					Min: &protos.MinStat{
						Value: math.Inf(+1),
					},
					Avg: &protos.AvgStat{
						Sum:   0.0,
						Count: 0,
					},
					Max: &protos.MaxStat{
						Value: math.Inf(-1),
					},
				},
			},
		},
		{
			name: "update default string without state",
			fields: fields{
				maxProcessedFields: 100,
				fieldsProcessed:    0,
			},
			args: args{
				state: protos.NewStringValue(),
				str:   &emptyString,
			},
			want: &protos.StringValue{
				TotalCount:   1,
				NullCount:    0,
				DefaultCount: 1,
				Length: &protos.NumberStat{
					Min: &protos.MinStat{
						Value: 0.0,
					},
					Avg: &protos.AvgStat{
						Sum:   0.0,
						Count: 1,
					},
					Max: &protos.MaxStat{
						Value: 0.0,
					},
				},
			},
		},
		{
			name: "update non-default string without state",
			fields: fields{
				maxProcessedFields: 100,
				fieldsProcessed:    0,
			},
			args: args{
				state: protos.NewStringValue(),
				str:   &somethingString,
			},
			want: &protos.StringValue{
				TotalCount:   1,
				NullCount:    0,
				DefaultCount: 0,
				Length: &protos.NumberStat{
					Min: &protos.MinStat{
						Value: 9.0,
					},
					Avg: &protos.AvgStat{
						Sum:   9.0,
						Count: 1,
					},
					Max: &protos.MaxStat{
						Value: 9.0,
					},
				},
			},
		},
		{
			name: "update nil string with state",
			fields: fields{
				maxProcessedFields: 100,
				fieldsProcessed:    0,
			},
			args: args{
				state: &protos.StringValue{
					TotalCount:   10,
					NullCount:    9,
					DefaultCount: 0,
					Length: &protos.NumberStat{
						Min: &protos.MinStat{
							Value: 9.0,
						},
						Avg: &protos.AvgStat{
							Sum:   9.0,
							Count: 1,
						},
						Max: &protos.MaxStat{
							Value: 9.0,
						},
					},
				},
				str: nil,
			},
			want: &protos.StringValue{
				TotalCount:   11,
				NullCount:    10,
				DefaultCount: 0,
				Length: &protos.NumberStat{
					Min: &protos.MinStat{
						Value: 9.0,
					},
					Avg: &protos.AvgStat{
						Sum:   9.0,
						Count: 1,
					},
					Max: &protos.MaxStat{
						Value: 9.0,
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := &Value{
				maxProcessedFields: tt.fields.maxProcessedFields,
				fieldsProcessed:    tt.fields.fieldsProcessed,
			}
			got, err := v.updateString(tt.args.state, tt.args.str)
			if (err != nil) != tt.wantErr {
				t.Errorf("Value.updateString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Value.updateString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValue_updateArray(t *testing.T) {
	type fields struct {
		maxProcessedFields int
		fieldsProcessed    int
	}
	type args struct {
		state *protos.ArrayValue
		array []interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *protos.ArrayValue
		wantErr bool
	}{
		{
			name: "update nil array without state",
			fields: fields{
				maxProcessedFields: 100,
				fieldsProcessed:    0,
			},
			args: args{
				state: protos.NewArrayValue(),
				array: nil,
			},
			want: &protos.ArrayValue{
				TotalCount:   1,
				NullCount:    1,
				DefaultCount: 0,
				Values:       protos.NewValueValue(),
			},
		},
		{
			name: "update default array without state",
			fields: fields{
				maxProcessedFields: 100,
				fieldsProcessed:    0,
			},
			args: args{
				state: protos.NewArrayValue(),
				array: []interface{}{},
			},
			want: &protos.ArrayValue{
				TotalCount:   1,
				NullCount:    0,
				DefaultCount: 1,
				Values:       protos.NewValueValue(),
			},
		},
		{
			name: "update boolean array without state",
			fields: fields{
				maxProcessedFields: 100,
				fieldsProcessed:    0,
			},
			args: args{
				state: protos.NewArrayValue(),
				array: []interface{}{nil, true, false},
			},
			want: &protos.ArrayValue{
				TotalCount:   1,
				NullCount:    0,
				DefaultCount: 0,
				Values: &protos.ValueValue{
					TotalCount: 3,
					NullCount:  1,
					Number:     nil,
					String_:    nil,
					Boolean: &protos.BooleanValue{
						TotalCount:   3,
						NullCount:    1,
						DefaultCount: 1,
						TotalFalse:   1,
						TotalTrue:    1,
					},
					Array: nil,
					Obj:   nil,
				},
			},
		},
		{
			name: "update boolean array with state",
			fields: fields{
				maxProcessedFields: 100,
				fieldsProcessed:    0,
			},
			args: args{
				state: &protos.ArrayValue{
					TotalCount:   10,
					NullCount:    9,
					DefaultCount: 0,
					Values: &protos.ValueValue{
						TotalCount: 1,
						NullCount:  0,
						Number:     nil,
						String_:    nil,
						Boolean: &protos.BooleanValue{
							TotalCount:   1,
							NullCount:    0,
							DefaultCount: 1,
							TotalFalse:   1,
							TotalTrue:    0,
						},
						Array: nil,
						Obj:   nil,
					},
				},
				array: []interface{}{true},
			},
			want: &protos.ArrayValue{
				TotalCount:   11,
				NullCount:    9,
				DefaultCount: 0,
				Values: &protos.ValueValue{
					TotalCount: 2,
					NullCount:  0,
					Number:     nil,
					String_:    nil,
					Boolean: &protos.BooleanValue{
						TotalCount:   2,
						NullCount:    0,
						DefaultCount: 1,
						TotalFalse:   1,
						TotalTrue:    1,
					},
					Array: nil,
					Obj:   nil,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := &Value{
				maxProcessedFields: tt.fields.maxProcessedFields,
				fieldsProcessed:    tt.fields.fieldsProcessed,
			}
			got, err := v.updateArray(tt.args.state, tt.args.array)
			if (err != nil) != tt.wantErr {
				t.Errorf("Value.updateArray() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Value.updateArray() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValue_updateObj(t *testing.T) {
	type fields struct {
		maxProcessedFields int
		fieldsProcessed    int
	}
	type args struct {
		state *protos.ObjValue
		m     map[string]interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *protos.ObjValue
		wantErr bool
	}{
		{
			name: "update nil object without state",
			fields: fields{
				maxProcessedFields: 100,
				fieldsProcessed:    0,
			},
			args: args{
				state: protos.NewObjValue(),
				m:     nil,
			},
			want: &protos.ObjValue{
				TotalCount:   1,
				NullCount:    1,
				DefaultCount: 0,
				Fields:       map[string]*protos.ValueValue{},
			},
		},
		{
			name: "update default object without state",
			fields: fields{
				maxProcessedFields: 100,
				fieldsProcessed:    0,
			},
			args: args{
				state: protos.NewObjValue(),
				m:     map[string]interface{}{},
			},
			want: &protos.ObjValue{
				TotalCount:   1,
				NullCount:    0,
				DefaultCount: 1,
				Fields:       map[string]*protos.ValueValue{},
			},
		},
		{
			name: "update non-default object without state",
			fields: fields{
				maxProcessedFields: 100,
				fieldsProcessed:    0,
			},
			args: args{
				state: protos.NewObjValue(),
				m: map[string]interface{}{
					"booleanField": true,
				},
			},
			want: &protos.ObjValue{
				TotalCount:   1,
				NullCount:    0,
				DefaultCount: 0,
				Fields: map[string]*protos.ValueValue{
					"booleanField": {
						TotalCount: 1,
						NullCount:  0,
						Boolean: &protos.BooleanValue{
							TotalCount:   1,
							NullCount:    0,
							DefaultCount: 0,
							TotalFalse:   0,
							TotalTrue:    1,
						},
					},
				},
			},
		},
		{
			name: "update non-default object with state",
			fields: fields{
				maxProcessedFields: 100,
				fieldsProcessed:    0,
			},
			args: args{
				state: &protos.ObjValue{
					TotalCount:   3,
					NullCount:    1,
					DefaultCount: 0,
					Fields: map[string]*protos.ValueValue{
						"booleanField": {
							TotalCount: 2,
							NullCount:  0,
							Boolean: &protos.BooleanValue{
								TotalCount:   2,
								NullCount:    0,
								DefaultCount: 0,
								TotalFalse:   0,
								TotalTrue:    2,
							},
						},
					},
				},
				m: map[string]interface{}{
					"numberField": 1.0,
				},
			},
			want: &protos.ObjValue{
				TotalCount:   4,
				NullCount:    1,
				DefaultCount: 0,
				Fields: map[string]*protos.ValueValue{
					"booleanField": {
						TotalCount: 3,
						NullCount:  1,
						Boolean: &protos.BooleanValue{
							TotalCount:   3,
							NullCount:    1,
							DefaultCount: 0,
							TotalFalse:   0,
							TotalTrue:    2,
						},
					},
					"numberField": {
						TotalCount: 3,
						NullCount:  2,
						Number: &protos.NumberValue{
							TotalCount:   3,
							NullCount:    2,
							DefaultCount: 0,
							Min: &protos.MinStat{
								Value: 1.0,
							},
							Avg: &protos.AvgStat{
								Sum:   1.0,
								Count: 1,
							},
							Max: &protos.MaxStat{
								Value: 1.0,
							},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := &Value{
				maxProcessedFields: tt.fields.maxProcessedFields,
				fieldsProcessed:    tt.fields.fieldsProcessed,
			}
			got, err := v.updateObj(tt.args.state, tt.args.m)
			if (err != nil) != tt.wantErr {
				t.Errorf("Value.updateObj() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Value.updateObj() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValue_updateValue(t *testing.T) {
	type fields struct {
		maxProcessedFields int
		fieldsProcessed    int
	}
	type args struct {
		state         *protos.ValueValue
		jsonInterface interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *protos.ValueValue
		wantErr bool
	}{
		{
			name: "update nil value without state",
			fields: fields{
				maxProcessedFields: 100,
				fieldsProcessed:    0,
			},
			args: args{
				state:         protos.NewValueValue(),
				jsonInterface: nil,
			},
			want: &protos.ValueValue{
				TotalCount: 1,
				NullCount:  1,
				Number:     nil,
				String_:    nil,
				Boolean:    nil,
				Array:      nil,
				Obj:        nil,
			},
		},
		{
			name: "update boolean value without state",
			fields: fields{
				maxProcessedFields: 100,
				fieldsProcessed:    0,
			},
			args: args{
				state:         protos.NewValueValue(),
				jsonInterface: true,
			},
			want: &protos.ValueValue{
				TotalCount: 1,
				NullCount:  0,
				Number:     nil,
				String_:    nil,
				Boolean: &protos.BooleanValue{
					TotalCount:   1,
					NullCount:    0,
					DefaultCount: 0,
					TotalFalse:   0,
					TotalTrue:    1,
				},
				Array: nil,
				Obj:   nil,
			},
		},
		{
			name: "update boolean value with boolean state",
			fields: fields{
				maxProcessedFields: 100,
				fieldsProcessed:    0,
			},
			args: args{
				state: &protos.ValueValue{
					TotalCount: 1,
					NullCount:  0,
					Number:     nil,
					String_:    nil,
					Boolean: &protos.BooleanValue{
						TotalCount:   1,
						NullCount:    0,
						DefaultCount: 0,
						TotalFalse:   0,
						TotalTrue:    1,
					},
					Array: nil,
					Obj:   nil,
				},
				jsonInterface: true,
			},
			want: &protos.ValueValue{
				TotalCount: 2,
				NullCount:  0,
				Number:     nil,
				String_:    nil,
				Boolean: &protos.BooleanValue{
					TotalCount:   2,
					NullCount:    0,
					DefaultCount: 0,
					TotalFalse:   0,
					TotalTrue:    2,
				},
				Array: nil,
				Obj:   nil,
			},
		},
		{
			name: "update number value with boolean state",
			fields: fields{
				maxProcessedFields: 100,
				fieldsProcessed:    0,
			},
			args: args{
				state: &protos.ValueValue{
					TotalCount: 1,
					NullCount:  0,
					Number:     nil,
					String_:    nil,
					Boolean: &protos.BooleanValue{
						TotalCount:   1,
						NullCount:    0,
						DefaultCount: 0,
						TotalFalse:   0,
						TotalTrue:    1,
					},
					Array: nil,
					Obj:   nil,
				},
				jsonInterface: oneNumber,
			},
			want: &protos.ValueValue{
				TotalCount: 2,
				NullCount:  0,
				Number: &protos.NumberValue{
					TotalCount:   2,
					NullCount:    1,
					DefaultCount: 0,
					Min: &protos.MinStat{
						Value: 1.0,
					},
					Avg: &protos.AvgStat{
						Sum:   1.0,
						Count: 1,
					},
					Max: &protos.MaxStat{
						Value: 1.0,
					},
				},
				String_: nil,
				Boolean: &protos.BooleanValue{
					TotalCount:   2,
					NullCount:    1,
					DefaultCount: 0,
					TotalFalse:   0,
					TotalTrue:    1,
				},
				Array: nil,
				Obj:   nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := &Value{
				maxProcessedFields: tt.fields.maxProcessedFields,
				fieldsProcessed:    tt.fields.fieldsProcessed,
			}
			got, err := v.updateValue(tt.args.state, tt.args.jsonInterface)
			if (err != nil) != tt.wantErr {
				t.Errorf("Value.updateValue() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Value.updateValue() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValue_AddSampleData(t *testing.T) {
	type fields struct {
		maxProcessedFields int
		fieldsProcessed    int
		digest             *protos.ObjValue
	}
	type args struct {
		sampleData *data.Data
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *protos.ObjValue
		wantErr bool
	}{
		{
			name: "add boolean field to empty digest",
			fields: fields{
				maxProcessedFields: 100,
				fieldsProcessed:    0,
				digest:             protos.NewObjValue(),
			},
			args: args{
				sampleData: data.NewSampleDataFromJSON(`{"booleanField": true}`),
			},
			want: &protos.ObjValue{
				TotalCount:   1,
				NullCount:    0,
				DefaultCount: 0,
				Fields: map[string]*protos.ValueValue{
					"booleanField": {
						TotalCount: 1,
						NullCount:  0,
						Boolean: &protos.BooleanValue{
							TotalCount:   1,
							NullCount:    0,
							DefaultCount: 0,
							TotalFalse:   0,
							TotalTrue:    1,
						},
					},
				},
			},
		},
		{
			name: "add number array field to empty digest",
			fields: fields{
				maxProcessedFields: 100,
				fieldsProcessed:    0,
				digest:             protos.NewObjValue(),
			},
			args: args{
				sampleData: data.NewSampleDataFromJSON(`{"arrayField": [0.0, 1.0, 2.0, 3.0, null]}`),
			},
			want: &protos.ObjValue{
				TotalCount:   1,
				NullCount:    0,
				DefaultCount: 0,
				Fields: map[string]*protos.ValueValue{
					"arrayField": {
						TotalCount: 1,
						NullCount:  0,
						Array: &protos.ArrayValue{
							TotalCount:   1,
							NullCount:    0,
							DefaultCount: 0,
							Values: &protos.ValueValue{
								TotalCount: 5,
								NullCount:  1,
								Number: &protos.NumberValue{
									TotalCount:   5,
									NullCount:    1,
									DefaultCount: 1,
									Min: &protos.MinStat{
										Value: 0.0,
									},
									Avg: &protos.AvgStat{
										Sum:   6,
										Count: 4,
									},
									Max: &protos.MaxStat{
										Value: 3.0,
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "add object array field to empty digest",
			fields: fields{
				maxProcessedFields: 100,
				fieldsProcessed:    0,
				digest:             protos.NewObjValue(),
			},
			args: args{
				sampleData: data.NewSampleDataFromJSON(`{"arrayField": [{"numberField": 5}, {"numberField": 10, "booleanField": false}, {}, null]}`),
			},
			want: &protos.ObjValue{
				TotalCount:   1,
				NullCount:    0,
				DefaultCount: 0,
				Fields: map[string]*protos.ValueValue{
					"arrayField": {
						TotalCount: 1,
						NullCount:  0,
						Array: &protos.ArrayValue{
							TotalCount:   1,
							NullCount:    0,
							DefaultCount: 0,
							Values: &protos.ValueValue{
								TotalCount: 4,
								NullCount:  1,
								Obj: &protos.ObjValue{
									TotalCount:   4,
									NullCount:    1,
									DefaultCount: 1,
									Fields: map[string]*protos.ValueValue{
										"numberField": {
											TotalCount: 3,
											NullCount:  1,
											Number: &protos.NumberValue{
												TotalCount:   3,
												NullCount:    1,
												DefaultCount: 0,
												Min: &protos.MinStat{
													Value: 5.0,
												},
												Avg: &protos.AvgStat{
													Sum:   15.0,
													Count: 2,
												},
												Max: &protos.MaxStat{
													Value: 10.0,
												},
											},
										},
										"booleanField": {
											TotalCount: 3,
											NullCount:  2,
											Boolean: &protos.BooleanValue{
												TotalCount:   3,
												NullCount:    2,
												DefaultCount: 1,
												TotalFalse:   1,
												TotalTrue:    0,
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Value{
				maxProcessedFields: tt.fields.maxProcessedFields,
				fieldsProcessed:    tt.fields.fieldsProcessed,
				digest:             tt.fields.digest,
			}
			if err := s.AddSampleData(tt.args.sampleData); (err != nil) != tt.wantErr {
				t.Errorf("Value.AddSampleData() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(s.digest, tt.want) {
				t.Errorf("Value.updateValue() = %v, want %v", s.digest, tt.want)
			}
		})
	}
}
