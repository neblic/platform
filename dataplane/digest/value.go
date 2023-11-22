package digest

import (
	"errors"
	"fmt"
	"math"

	"github.com/neblic/platform/controlplane/control"
	"github.com/neblic/platform/dataplane/digest/types"
	"github.com/neblic/platform/internal/pkg/data"
	"google.golang.org/protobuf/encoding/protojson"
)

type ValueType int8

const (
	UnknownValueType = ValueType(iota)
	NumberValueType
	StringValueType
	BooleanValueType
	ObjValueType
	ArrayValueType
	NilValueType
)

type Value struct {
	maxProcessedFields int

	fieldsProcessed int
	digest          *types.ObjValue
}

func (v *Value) isDigest() {}

func NewValue(maxProcessedFields int) *Value {
	return &Value{
		maxProcessedFields: maxProcessedFields,

		digest: types.NewObjValue(),
	}
}

func (v *Value) updateMinStat(state *types.MinStat, value float64) {
	state.Value = math.Min(state.Value, value)
}

func (v *Value) updateAvgStat(state *types.AvgStat, value float64) {
	state.Sum += value
	state.Count++
}

func (v *Value) updateMaxStat(state *types.MaxStat, value float64) {
	state.Value = math.Max(state.Value, value)
}

func (v *Value) updateBoolean(state *types.BooleanValue, boolean *bool) (*types.BooleanValue, error) {
	if err := v.incrFieldsProcessed(); err != nil {
		return nil, err
	}

	if boolean != nil {
		if *boolean {
			state.TrueCount++
		} else {
			state.DefaultCount++
			state.FalseCount++
		}
	} else {
		state.NullCount++
	}
	state.TotalCount++

	return state, nil
}

func (v *Value) updateNum(state *types.NumberValue, number *float64) (*types.NumberValue, error) {
	if err := v.incrFieldsProcessed(); err != nil {
		return nil, err
	}

	if number != nil {
		if *number == 0.0 {
			state.DefaultCount++
		}
		v.updateMinStat(state.Min, *number)
		v.updateAvgStat(state.Avg, *number)
		v.updateMaxStat(state.Max, *number)
		state.HyperLogLog.InsertFloat64(*number)
	} else {
		state.NullCount++
	}
	state.TotalCount++

	return state, nil

}

func (v *Value) updateString(state *types.StringValue, str *string) (*types.StringValue, error) {
	if err := v.incrFieldsProcessed(); err != nil {
		return nil, err
	}

	if str != nil {
		if *str == "" {
			state.DefaultCount++
		}
		state.HyperLogLog.InsertBytes([]byte(*str))
		v.updateMinStat(state.Length.Min, float64(len(*str)))
		v.updateAvgStat(state.Length.Avg, float64(len(*str)))
		v.updateMaxStat(state.Length.Max, float64(len(*str)))
		state.Length.HyperLogLog.InsertInt64(int64(len(*str)))

	} else {
		state.NullCount++
	}
	state.TotalCount++

	return state, nil
}

func (v *Value) updateArray(state *types.ArrayValue, array []interface{}) (*types.ArrayValue, error) {
	if err := v.incrFieldsProcessed(); err != nil {
		return nil, err
	}

	if array != nil {
		if len(array) == 0 {
			state.DefaultCount++
		}

		// Just one digest is kept for the full array, update it with all the array entries
		for _, value := range array {
			_, err := v.updateValue(state.Values, value)
			if err != nil {
				return state, err
			}
		}
	} else {
		state.NullCount++
	}
	state.TotalCount++

	return state, nil
}

func (v *Value) updateObj(state *types.ObjValue, m map[string]interface{}) (*types.ObjValue, error) {
	if err := v.incrFieldsProcessed(); err != nil {
		return nil, err
	}

	if m != nil {
		// Update digest fields
		for key, value := range m {
			valueState, ok := state.Fields[key]
			if !ok {
				nonNullCounter := state.TotalCount - state.NullCount
				valueState = types.NewValueValue()
				valueState.NullCount = nonNullCounter
				valueState.TotalCount = nonNullCounter
				state.Fields[key] = valueState
			}
			_, err := v.updateValue(valueState, value)
			if err != nil {
				return state, err
			}
		}

		// Update fields not present in the received map
		for key, valueState := range state.Fields {
			_, ok := m[key]
			if !ok {
				_, err := v.updateValue(valueState, nil)
				if err != nil {
					return nil, err
				}
			}
		}

		if len(m) == 0 {
			state.DefaultCount++
		}

	} else {
		state.NullCount++
	}
	state.TotalCount++

	return state, nil
}

func (v *Value) updateValue(state *types.ValueValue, jsonInterface interface{}) (*types.ValueValue, error) {
	if err := v.incrFieldsProcessed(); err != nil {
		return nil, err
	}

	computedDigestType := UnknownValueType

	switch jsonValue := jsonInterface.(type) {
	case bool:
		// Initialize boolean digest if necessary. Update otherwise
		if state.Boolean == nil {
			// First time a boolean is registered in this digest. The number of times this digest was seen is
			// stored in the structure digest statistics, use it as a reference to track how many times this digest was null
			state.Boolean = types.NewBooleanValue()
			state.Boolean.NullCount = state.TotalCount
			state.Boolean.TotalCount = state.TotalCount
		}

		_, err := v.updateBoolean(state.Boolean, &jsonValue)
		if err != nil {
			return state, err
		}

		state.TotalCount++

		computedDigestType = BooleanValueType

	case float64:
		// Initialize float digest if necessary
		if state.Number == nil {
			// First time a float is registered in this digest. The number of times this digest was seen is
			// stored in the structure digest statistics, use it as a reference to track how many times this digest was null
			state.Number = types.NewNumberValue()
			state.Number.NullCount = state.TotalCount
			state.Number.TotalCount = state.TotalCount
		}

		_, err := v.updateNum(state.Number, &jsonValue)
		if err != nil {
			return state, err
		}

		state.TotalCount++

		computedDigestType = NumberValueType

	case string:
		// Initialize string digest if necessary. Update otherwise
		if state.String == nil {
			// First time a string is registered in this digest. The number of times this digest was seen is
			// stored in the structure digest statistics, use it as a reference to track how many times this digest was null
			state.String = types.NewStringValue()
			state.String.NullCount = state.TotalCount
			state.String.TotalCount = state.TotalCount
		}

		_, err := v.updateString(state.String, &jsonValue)
		if err != nil {
			return state, err
		}

		state.TotalCount++

		computedDigestType = StringValueType

	case []interface{}:
		// Initialize array digest if necessary. Update otherwise
		if state.Array == nil {
			// First time an array is registered in this digest. The number of times this digest was seen is
			// stored in the structure digest statistics, use it as a reference to track how many times this digest was null
			state.Array = types.NewArrayValue()
			state.Array.NullCount = state.TotalCount
			state.Array.TotalCount = state.TotalCount
		}

		_, err := v.updateArray(state.Array, jsonValue)
		if err != nil {
			return state, err
		}

		state.TotalCount++

		computedDigestType = ArrayValueType

	case map[string]interface{}:
		// Initialize map digest if necessary. Update otherwise
		if state.Obj == nil {
			state.Obj = types.NewObjValue()
			state.Obj.NullCount = state.TotalCount
			state.Obj.TotalCount = state.TotalCount
		}

		_, err := v.updateObj(state.Obj, jsonValue)
		if err != nil {
			return state, err
		}

		state.TotalCount++

		computedDigestType = ObjValueType

	case nil:
		state.NullCount++
		state.TotalCount++

		computedDigestType = NilValueType
	default:
		return nil, fmt.Errorf("value digest does not support %T type", jsonInterface)
	}

	// Update null statistics for the initialized digests that don't
	// correspond with the computed digest.
	if computedDigestType != NumberValueType && state.Number != nil {
		state.Number.NullCount++
		state.Number.TotalCount++
	}
	if computedDigestType != StringValueType && state.String != nil {
		state.String.NullCount++
		state.String.TotalCount++
	}
	if computedDigestType != BooleanValueType && state.Boolean != nil {
		state.Boolean.NullCount++
		state.Boolean.TotalCount++
	}
	if computedDigestType != ObjValueType && state.Obj != nil {
		state.Obj.NullCount++
		state.Obj.TotalCount++
	}
	if computedDigestType != ArrayValueType && state.Array != nil {
		state.Array.NullCount++
		state.Array.TotalCount++
	}

	return state, nil
}

func (v *Value) incrFieldsProcessed() error {
	v.fieldsProcessed++

	if v.fieldsProcessed > v.maxProcessedFields {
		return fmt.Errorf("%w %d", errMaxFieldsProcessed, v.maxProcessedFields)
	}

	return nil
}

// AddSample is not thread safe
func (v *Value) AddSampleData(sampleData *data.Data) error {
	dataMap, err := sampleData.Map()
	if err != nil {
		return err
	}

	v.fieldsProcessed = 0
	updatedObj, err := v.updateObj(v.digest, dataMap)
	if err != nil && !errors.Is(err, errMaxFieldsProcessed) {
		return err
	}

	v.digest = updatedObj

	return nil
}

func (v *Value) JSON() ([]byte, error) {
	json, err := protojson.Marshal(v.digest.ToProto())
	if err != nil {
		return nil, fmt.Errorf("failed to marshal digest to json: %w", err)
	}

	return json, nil
}

func (v *Value) Reset() {
	v.digest = types.NewObjValue()
}

func (v *Value) String() string {
	return "ValueDigest"
}

func (v *Value) SampleType() control.SampleType {
	return control.ValueDigestSampleType
}
