package digest

import (
	"math"
	"reflect"
	"testing"

	"github.com/neblic/platform/dataplane/digest/types"
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
		state   *types.BooleanValue
		boolean *bool
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *types.BooleanValue
		wantErr bool
	}{
		{
			name: "update nil boolean without state",
			fields: fields{
				maxProcessedFields: 100,
				fieldsProcessed:    0,
			},
			args: args{
				state:   types.NewBooleanValue(),
				boolean: nil,
			},
			want: &types.BooleanValue{
				TotalCount:   1,
				NullCount:    1,
				DefaultCount: 0,
				FalseCount:   0,
				TrueCount:    0,
			},
		},
		{
			name: "update default boolean without state",
			fields: fields{
				maxProcessedFields: 100,
				fieldsProcessed:    0,
			},
			args: args{
				state:   types.NewBooleanValue(),
				boolean: &falseBoolean,
			},
			want: &types.BooleanValue{
				TotalCount:   1,
				NullCount:    0,
				DefaultCount: 1,
				FalseCount:   1,
				TrueCount:    0,
			},
		},
		{
			name: "update non-default boolean without state",
			fields: fields{
				maxProcessedFields: 100,
				fieldsProcessed:    0,
			},
			args: args{
				state:   types.NewBooleanValue(),
				boolean: &trueBoolean,
			},
			want: &types.BooleanValue{
				TotalCount:   1,
				NullCount:    0,
				DefaultCount: 0,
				FalseCount:   0,
				TrueCount:    1,
			},
		},
		{
			name: "update default boolean with state",
			fields: fields{
				maxProcessedFields: 100,
				fieldsProcessed:    0,
			},
			args: args{
				state: &types.BooleanValue{
					TotalCount:   10,
					NullCount:    8,
					DefaultCount: 1,
					FalseCount:   1,
					TrueCount:    1,
				},
				boolean: &falseBoolean,
			},
			want: &types.BooleanValue{
				TotalCount:   11,
				NullCount:    8,
				DefaultCount: 2,
				FalseCount:   2,
				TrueCount:    1,
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
		state  *types.NumberValue
		number *float64
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *types.NumberValue
		wantErr bool
	}{
		{
			name: "update nil number without state",
			fields: fields{
				maxProcessedFields: 100,
				fieldsProcessed:    0,
			},
			args: args{
				state:  types.NewNumberValue(),
				number: nil,
			},
			want: &types.NumberValue{
				TotalCount:   1,
				NullCount:    1,
				DefaultCount: 0,
				Min: &types.MinStat{
					Value: math.Inf(+1),
				},
				Avg: &types.AvgStat{
					Sum:   0.0,
					Count: 0,
				},
				Max: &types.MaxStat{
					Value: math.Inf(-1),
				},
				HyperLogLog: types.NewHyperLogLog(),
			},
		},
		{
			name: "update default number without state",
			fields: fields{
				maxProcessedFields: 100,
				fieldsProcessed:    0,
			},
			args: args{
				state:  types.NewNumberValue(),
				number: &zeroNumber,
			},
			want: &types.NumberValue{
				TotalCount:   1,
				NullCount:    0,
				DefaultCount: 1,
				Min: &types.MinStat{
					Value: 0.0,
				},
				Avg: &types.AvgStat{
					Sum:   0.0,
					Count: 1,
				},
				Max: &types.MaxStat{
					Value: 0.0,
				},
				HyperLogLog: types.NewHyperLogLog().InsertFloat64(0.0),
			},
		},
		{
			name: "update non-default number without state",
			fields: fields{
				maxProcessedFields: 100,
				fieldsProcessed:    0,
			},
			args: args{
				state:  types.NewNumberValue(),
				number: &oneNumber,
			},
			want: &types.NumberValue{
				TotalCount:   1,
				NullCount:    0,
				DefaultCount: 0,
				Min: &types.MinStat{
					Value: 1.0,
				},
				Avg: &types.AvgStat{
					Sum:   1.0,
					Count: 1,
				},
				Max: &types.MaxStat{
					Value: 1.0,
				},
				HyperLogLog: types.NewHyperLogLog().InsertFloat64(1.0),
			},
		},
		{
			name: "update nil number with state",
			fields: fields{
				maxProcessedFields: 100,
				fieldsProcessed:    0,
			},
			args: args{
				state: &types.NumberValue{
					TotalCount:   10,
					NullCount:    9,
					DefaultCount: 0,
					Min: &types.MinStat{
						Value: 1.0,
					},
					Avg: &types.AvgStat{
						Sum:   1.0,
						Count: 1,
					},
					Max: &types.MaxStat{
						Value: 1.0,
					},
					HyperLogLog: types.NewHyperLogLog(),
				},
				number: nil,
			},
			want: &types.NumberValue{
				TotalCount:   11,
				NullCount:    10,
				DefaultCount: 0,
				Min: &types.MinStat{
					Value: 1.0,
				},
				Avg: &types.AvgStat{
					Sum:   1.0,
					Count: 1,
				},
				Max: &types.MaxStat{
					Value: 1.0,
				},
				HyperLogLog: types.NewHyperLogLog(),
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
		state *types.StringValue
		str   *string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *types.StringValue
		wantErr bool
	}{
		{
			name: "update nil string without state",
			fields: fields{
				maxProcessedFields: 100,
				fieldsProcessed:    0,
			},
			args: args{
				state: types.NewStringValue(),
				str:   nil,
			},
			want: &types.StringValue{
				TotalCount:   1,
				NullCount:    1,
				DefaultCount: 0,
				HyperLogLog:  types.NewHyperLogLog(),
				Length: &types.NumberStat{
					Min: &types.MinStat{
						Value: math.Inf(+1),
					},
					Avg: &types.AvgStat{
						Sum:   0.0,
						Count: 0,
					},
					Max: &types.MaxStat{
						Value: math.Inf(-1),
					},
					HyperLogLog: types.NewHyperLogLog(),
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
				state: types.NewStringValue(),
				str:   &emptyString,
			},
			want: &types.StringValue{
				TotalCount:   1,
				NullCount:    0,
				DefaultCount: 1,
				HyperLogLog:  types.NewHyperLogLog().InsertBytes([]byte("")),
				Length: &types.NumberStat{
					Min: &types.MinStat{
						Value: 0.0,
					},
					Avg: &types.AvgStat{
						Sum:   0.0,
						Count: 1,
					},
					Max: &types.MaxStat{
						Value: 0.0,
					},
					HyperLogLog: types.NewHyperLogLog().InsertInt64(0),
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
				state: types.NewStringValue(),
				str:   &somethingString,
			},
			want: &types.StringValue{
				TotalCount:   1,
				NullCount:    0,
				DefaultCount: 0,
				HyperLogLog:  types.NewHyperLogLog().InsertBytes([]byte("something")),
				Length: &types.NumberStat{
					Min: &types.MinStat{
						Value: 9.0,
					},
					Avg: &types.AvgStat{
						Sum:   9.0,
						Count: 1,
					},
					Max: &types.MaxStat{
						Value: 9.0,
					},
					HyperLogLog: types.NewHyperLogLog().InsertInt64(9),
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
				state: &types.StringValue{
					TotalCount:   10,
					NullCount:    9,
					DefaultCount: 0,
					HyperLogLog:  types.NewHyperLogLog(),
					Length: &types.NumberStat{
						Min: &types.MinStat{
							Value: 9.0,
						},
						Avg: &types.AvgStat{
							Sum:   9.0,
							Count: 1,
						},
						Max: &types.MaxStat{
							Value: 9.0,
						},
						HyperLogLog: types.NewHyperLogLog(),
					},
				},
				str: nil,
			},
			want: &types.StringValue{
				TotalCount:   11,
				NullCount:    10,
				DefaultCount: 0,
				HyperLogLog:  types.NewHyperLogLog(),
				Length: &types.NumberStat{
					Min: &types.MinStat{
						Value: 9.0,
					},
					Avg: &types.AvgStat{
						Sum:   9.0,
						Count: 1,
					},
					Max: &types.MaxStat{
						Value: 9.0,
					},
					HyperLogLog: types.NewHyperLogLog(),
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
		state *types.ArrayValue
		array []interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *types.ArrayValue
		wantErr bool
	}{
		{
			name: "update nil array without state",
			fields: fields{
				maxProcessedFields: 100,
				fieldsProcessed:    0,
			},
			args: args{
				state: types.NewArrayValue(),
				array: nil,
			},
			want: &types.ArrayValue{
				TotalCount:   1,
				NullCount:    1,
				DefaultCount: 0,
				Values:       types.NewValueValue(),
			},
		},
		{
			name: "update default array without state",
			fields: fields{
				maxProcessedFields: 100,
				fieldsProcessed:    0,
			},
			args: args{
				state: types.NewArrayValue(),
				array: []interface{}{},
			},
			want: &types.ArrayValue{
				TotalCount:   1,
				NullCount:    0,
				DefaultCount: 1,
				Values:       types.NewValueValue(),
			},
		},
		{
			name: "update boolean array without state",
			fields: fields{
				maxProcessedFields: 100,
				fieldsProcessed:    0,
			},
			args: args{
				state: types.NewArrayValue(),
				array: []interface{}{nil, true, false},
			},
			want: &types.ArrayValue{
				TotalCount:   1,
				NullCount:    0,
				DefaultCount: 0,
				Values: &types.ValueValue{
					TotalCount: 3,
					NullCount:  1,
					Number:     nil,
					String_:    nil,
					Boolean: &types.BooleanValue{
						TotalCount:   3,
						NullCount:    1,
						DefaultCount: 1,
						FalseCount:   1,
						TrueCount:    1,
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
				state: &types.ArrayValue{
					TotalCount:   10,
					NullCount:    9,
					DefaultCount: 0,
					Values: &types.ValueValue{
						TotalCount: 1,
						NullCount:  0,
						Number:     nil,
						String_:    nil,
						Boolean: &types.BooleanValue{
							TotalCount:   1,
							NullCount:    0,
							DefaultCount: 1,
							FalseCount:   1,
							TrueCount:    0,
						},
						Array: nil,
						Obj:   nil,
					},
				},
				array: []interface{}{true},
			},
			want: &types.ArrayValue{
				TotalCount:   11,
				NullCount:    9,
				DefaultCount: 0,
				Values: &types.ValueValue{
					TotalCount: 2,
					NullCount:  0,
					Number:     nil,
					String_:    nil,
					Boolean: &types.BooleanValue{
						TotalCount:   2,
						NullCount:    0,
						DefaultCount: 1,
						FalseCount:   1,
						TrueCount:    1,
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
		state *types.ObjValue
		m     map[string]interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *types.ObjValue
		wantErr bool
	}{
		{
			name: "update nil object without state",
			fields: fields{
				maxProcessedFields: 100,
				fieldsProcessed:    0,
			},
			args: args{
				state: types.NewObjValue(),
				m:     nil,
			},
			want: &types.ObjValue{
				TotalCount:   1,
				NullCount:    1,
				DefaultCount: 0,
				Fields:       map[string]*types.ValueValue{},
			},
		},
		{
			name: "update default object without state",
			fields: fields{
				maxProcessedFields: 100,
				fieldsProcessed:    0,
			},
			args: args{
				state: types.NewObjValue(),
				m:     map[string]interface{}{},
			},
			want: &types.ObjValue{
				TotalCount:   1,
				NullCount:    0,
				DefaultCount: 1,
				Fields:       map[string]*types.ValueValue{},
			},
		},
		{
			name: "update non-default object without state",
			fields: fields{
				maxProcessedFields: 100,
				fieldsProcessed:    0,
			},
			args: args{
				state: types.NewObjValue(),
				m: map[string]interface{}{
					"booleanField": true,
				},
			},
			want: &types.ObjValue{
				TotalCount:   1,
				NullCount:    0,
				DefaultCount: 0,
				Fields: map[string]*types.ValueValue{
					"booleanField": {
						TotalCount: 1,
						NullCount:  0,
						Boolean: &types.BooleanValue{
							TotalCount:   1,
							NullCount:    0,
							DefaultCount: 0,
							FalseCount:   0,
							TrueCount:    1,
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
				state: &types.ObjValue{
					TotalCount:   3,
					NullCount:    1,
					DefaultCount: 0,
					Fields: map[string]*types.ValueValue{
						"booleanField": {
							TotalCount: 2,
							NullCount:  0,
							Boolean: &types.BooleanValue{
								TotalCount:   2,
								NullCount:    0,
								DefaultCount: 0,
								FalseCount:   0,
								TrueCount:    2,
							},
						},
					},
				},
				m: map[string]interface{}{
					"numberField": 1.0,
				},
			},
			want: &types.ObjValue{
				TotalCount:   4,
				NullCount:    1,
				DefaultCount: 0,
				Fields: map[string]*types.ValueValue{
					"booleanField": {
						TotalCount: 3,
						NullCount:  1,
						Boolean: &types.BooleanValue{
							TotalCount:   3,
							NullCount:    1,
							DefaultCount: 0,
							FalseCount:   0,
							TrueCount:    2,
						},
					},
					"numberField": {
						TotalCount: 3,
						NullCount:  2,
						Number: &types.NumberValue{
							TotalCount:   3,
							NullCount:    2,
							DefaultCount: 0,
							Min: &types.MinStat{
								Value: 1.0,
							},
							Avg: &types.AvgStat{
								Sum:   1.0,
								Count: 1,
							},
							Max: &types.MaxStat{
								Value: 1.0,
							},
							HyperLogLog: types.NewHyperLogLog().InsertFloat64(1.0),
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
		state         *types.ValueValue
		jsonInterface interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *types.ValueValue
		wantErr bool
	}{
		{
			name: "update nil value without state",
			fields: fields{
				maxProcessedFields: 100,
				fieldsProcessed:    0,
			},
			args: args{
				state:         types.NewValueValue(),
				jsonInterface: nil,
			},
			want: &types.ValueValue{
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
				state:         types.NewValueValue(),
				jsonInterface: true,
			},
			want: &types.ValueValue{
				TotalCount: 1,
				NullCount:  0,
				Number:     nil,
				String_:    nil,
				Boolean: &types.BooleanValue{
					TotalCount:   1,
					NullCount:    0,
					DefaultCount: 0,
					FalseCount:   0,
					TrueCount:    1,
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
				state: &types.ValueValue{
					TotalCount: 1,
					NullCount:  0,
					Number:     nil,
					String_:    nil,
					Boolean: &types.BooleanValue{
						TotalCount:   1,
						NullCount:    0,
						DefaultCount: 0,
						FalseCount:   0,
						TrueCount:    1,
					},
					Array: nil,
					Obj:   nil,
				},
				jsonInterface: true,
			},
			want: &types.ValueValue{
				TotalCount: 2,
				NullCount:  0,
				Number:     nil,
				String_:    nil,
				Boolean: &types.BooleanValue{
					TotalCount:   2,
					NullCount:    0,
					DefaultCount: 0,
					FalseCount:   0,
					TrueCount:    2,
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
				state: &types.ValueValue{
					TotalCount: 1,
					NullCount:  0,
					Number:     nil,
					String_:    nil,
					Boolean: &types.BooleanValue{
						TotalCount:   1,
						NullCount:    0,
						DefaultCount: 0,
						FalseCount:   0,
						TrueCount:    1,
					},
					Array: nil,
					Obj:   nil,
				},
				jsonInterface: oneNumber,
			},
			want: &types.ValueValue{
				TotalCount: 2,
				NullCount:  0,
				Number: &types.NumberValue{
					TotalCount:   2,
					NullCount:    1,
					DefaultCount: 0,
					Min: &types.MinStat{
						Value: 1.0,
					},
					Avg: &types.AvgStat{
						Sum:   1.0,
						Count: 1,
					},
					Max: &types.MaxStat{
						Value: 1.0,
					},
					HyperLogLog: types.NewHyperLogLog().InsertFloat64(1.0),
				},
				String_: nil,
				Boolean: &types.BooleanValue{
					TotalCount:   2,
					NullCount:    1,
					DefaultCount: 0,
					FalseCount:   0,
					TrueCount:    1,
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
		digest             *types.ObjValue
	}
	type args struct {
		sampleData *data.Data
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *types.ObjValue
		wantErr bool
	}{
		{
			name: "add boolean field to empty digest",
			fields: fields{
				maxProcessedFields: 100,
				fieldsProcessed:    0,
				digest:             types.NewObjValue(),
			},
			args: args{
				sampleData: data.NewSampleDataFromJSON(`{"booleanField": true}`),
			},
			want: &types.ObjValue{
				TotalCount:   1,
				NullCount:    0,
				DefaultCount: 0,
				Fields: map[string]*types.ValueValue{
					"booleanField": {
						TotalCount: 1,
						NullCount:  0,
						Boolean: &types.BooleanValue{
							TotalCount:   1,
							NullCount:    0,
							DefaultCount: 0,
							FalseCount:   0,
							TrueCount:    1,
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
				digest:             types.NewObjValue(),
			},
			args: args{
				sampleData: data.NewSampleDataFromJSON(`{"arrayField": [0.0, 1.0, 2.0, 3.0, null]}`),
			},
			want: &types.ObjValue{
				TotalCount:   1,
				NullCount:    0,
				DefaultCount: 0,
				Fields: map[string]*types.ValueValue{
					"arrayField": {
						TotalCount: 1,
						NullCount:  0,
						Array: &types.ArrayValue{
							TotalCount:   1,
							NullCount:    0,
							DefaultCount: 0,
							Values: &types.ValueValue{
								TotalCount: 5,
								NullCount:  1,
								Number: &types.NumberValue{
									TotalCount:   5,
									NullCount:    1,
									DefaultCount: 1,
									Min: &types.MinStat{
										Value: 0.0,
									},
									Avg: &types.AvgStat{
										Sum:   6,
										Count: 4,
									},
									Max: &types.MaxStat{
										Value: 3.0,
									},
									HyperLogLog: types.NewHyperLogLog().InsertFloat64(0.0).InsertFloat64(1.0).InsertFloat64(2.0).InsertFloat64(3.0),
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
				digest:             types.NewObjValue(),
			},
			args: args{
				sampleData: data.NewSampleDataFromJSON(`{"arrayField": [{"numberField": 5}, {"numberField": 10, "booleanField": false}, {}, null]}`),
			},
			want: &types.ObjValue{
				TotalCount:   1,
				NullCount:    0,
				DefaultCount: 0,
				Fields: map[string]*types.ValueValue{
					"arrayField": {
						TotalCount: 1,
						NullCount:  0,
						Array: &types.ArrayValue{
							TotalCount:   1,
							NullCount:    0,
							DefaultCount: 0,
							Values: &types.ValueValue{
								TotalCount: 4,
								NullCount:  1,
								Obj: &types.ObjValue{
									TotalCount:   4,
									NullCount:    1,
									DefaultCount: 1,
									Fields: map[string]*types.ValueValue{
										"numberField": {
											TotalCount: 3,
											NullCount:  1,
											Number: &types.NumberValue{
												TotalCount:   3,
												NullCount:    1,
												DefaultCount: 0,
												Min: &types.MinStat{
													Value: 5.0,
												},
												Avg: &types.AvgStat{
													Sum:   15.0,
													Count: 2,
												},
												Max: &types.MaxStat{
													Value: 10.0,
												},
												HyperLogLog: types.NewHyperLogLog().InsertFloat64(5.0).InsertFloat64(10.0),
											},
										},
										"booleanField": {
											TotalCount: 3,
											NullCount:  2,
											Boolean: &types.BooleanValue{
												TotalCount:   3,
												NullCount:    2,
												DefaultCount: 1,
												FalseCount:   1,
												TrueCount:    0,
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
