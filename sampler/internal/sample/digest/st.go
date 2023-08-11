package digest

import (
	"errors"
	"fmt"
	"math"
	"reflect"

	"github.com/neblic/platform/sampler/internal/sample"
	"github.com/neblic/platform/sampler/protos"
	"google.golang.org/protobuf/encoding/protojson"
)

var errMaxFieldsProcessed = fmt.Errorf("max number of fields processed reached")

type St struct {
	maxProcessedFields int
	notifyErr          func(error)

	fieldsProcessed int
	digest          *protos.StructureDigest
}

func (s *St) isDigest() {}

func NewStDigest(maxProcessedFields int, notifyErr func(error)) *St {
	return &St{
		maxProcessedFields: maxProcessedFields,
		notifyErr:          notifyErr,

		digest: &protos.StructureDigest{},
	}
}

func num64(n interface{}) interface{} {
	switch n := n.(type) {
	case int:
		return int64(n)
	case int8:
		return int64(n)
	case int16:
		return int64(n)
	case int32:
		return int64(n)
	case int64:
		return int64(n)
	case uint:
		return uint64(n)
	case uint8:
		return uint64(n)
	case uint16:
		return uint64(n)
	case uint32:
		return uint64(n)
	case uint64:
		return uint64(n)
	case float32:
		return float64(n)
	case float64:
		return float64(n)
	}
	return nil
}

func (s *St) updateNum(prev *protos.NumberType, x interface{}) (*protos.NumberType, error) {
	if err := s.incrFieldsProcessed(); err != nil {
		return nil, err
	}

	if prev == nil {
		prev = &protos.NumberType{}
	}

	switch num64(x).(type) {
	case int64:
		if prev.IntegerNum == nil {
			prev.IntegerNum = &protos.IntNumType{}
		}
		prev.IntegerNum.Count++
	case uint64:
		if prev.UintegerNum == nil {
			prev.UintegerNum = &protos.UIntNumType{}
		}
		prev.UintegerNum.Count++
	case float64:
		if prev.FloatNum == nil {
			prev.FloatNum = &protos.FloatNumType{}
		}
		prev.FloatNum.Count++
	default:
		return nil, fmt.Errorf("invalid number type %T", x)
	}

	return prev, nil
}

func (s *St) updateString(prev *protos.StringType, x string) (*protos.StringType, error) {
	if err := s.incrFieldsProcessed(); err != nil {
		return nil, err
	}

	if prev == nil {
		prev = &protos.StringType{}
	}
	prev.Count++

	return prev, nil
}

func (s *St) updateBoolean(prev *protos.BooleanType, x bool) (*protos.BooleanType, error) {
	if err := s.incrFieldsProcessed(); err != nil {
		return nil, err
	}

	if prev == nil {
		prev = &protos.BooleanType{}
	}
	prev.Count++

	return prev, nil
}

func (s *St) updateValue(prevVal *protos.Value, x interface{}) (*protos.Value, error) {
	if err := s.incrFieldsProcessed(); err != nil {
		return nil, err
	}

	if prevVal == nil {
		prevVal = &protos.Value{}
	}

	v := reflect.ValueOf(x)
	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		updatedNum, err := s.updateNum(prevVal.GetNumber(), v.Interface())
		if err != nil {
			return nil, err
		}
		prevVal.Number = updatedNum
	case reflect.String:
		updatedString, err := s.updateString(prevVal.GetString_(), v.String())
		if err != nil {
			return nil, err
		}
		prevVal.String_ = updatedString
	case reflect.Bool:
		updatedBoolean, err := s.updateBoolean(prevVal.GetBoolean(), v.Bool())
		if err != nil {
			return nil, err
		}
		prevVal.Boolean = updatedBoolean
	case reflect.Slice:
		updatedArray, err := s.updateArray(prevVal.GetArray(), v.Interface())
		if err != nil {
			return nil, err
		}
		prevVal.Array = updatedArray
	case reflect.Map:
		updatedObj, err := s.updateObj(prevVal.GetObj(), v.Interface())
		if err != nil {
			return nil, err
		}
		prevVal.Obj = updatedObj
	default:
		return nil, fmt.Errorf("invalid type %T", v.Interface())
	}

	return prevVal, nil
}

func (s *St) updateArray(prev *protos.ArrayType, x interface{}) (*protos.ArrayType, error) {
	if err := s.incrFieldsProcessed(); err != nil {
		return nil, err
	}

	v := reflect.ValueOf(x)
	if v.Kind() != reflect.Slice {
		return nil, fmt.Errorf("invalid type %T", x)
	}

	if prev == nil {
		prev = &protos.ArrayType{}
	}
	prev.Count++

	if prev.GetFixedLengthOrderedArray() == nil && prev.GetVariableLengthArray() == nil {
		prev.FixedLengthOrderedArray = &protos.FixedLengthOrderedArrayType{
			Fields: make([]*protos.Value, v.Len()),
		}
	}

	if fixedLenArr := prev.GetFixedLengthOrderedArray(); fixedLenArr != nil && len(fixedLenArr.GetFields()) == v.Len() {
		// as long as the array has the same length as the previous array, we consider it a fixed length array
		for i := 0; i < v.Len(); i++ {
			prevVal := fixedLenArr.GetFields()[i]
			prevVal, err := s.updateValue(prevVal, v.Index(i).Interface())
			if err != nil {
				s.notifyErr(err)
				continue
			}

			fixedLenArr.GetFields()[i] = prevVal
		}

	} else {
		// as soon as we find an array with a different length, we consider it a variable length array
		varLenArr := prev.GetVariableLengthArray()
		if varLenArr == nil {
			varLenArr = &protos.VariableLengthArrayType{}
			prev.VariableLengthArray = varLenArr
		}

		if prev.FixedLengthOrderedArray != nil {
			// updating from fixed length array to variable length array
			varLenArr.MinLength = int64(len(prev.FixedLengthOrderedArray.GetFields()))
			varLenArr.MaxLength = int64(v.Len())
			varLenArr.SumLength = int64(v.Len() + len(prev.FixedLengthOrderedArray.GetFields()))

			prev.FixedLengthOrderedArray = nil
		} else {
			varLenArr.MinLength = int64(math.Min(float64(varLenArr.MinLength), float64(v.Len())))
			varLenArr.MaxLength = int64(math.Max(float64(varLenArr.MaxLength), float64(v.Len())))
			varLenArr.SumLength += int64(v.Len())
		}
	}

	return prev, nil
}

func (s *St) updateObj(prev *protos.ObjType, x interface{}) (*protos.ObjType, error) {
	if err := s.incrFieldsProcessed(); err != nil {
		return nil, err
	}

	v := reflect.ValueOf(x)
	if v.Kind() != reflect.Map {
		return nil, fmt.Errorf("invalid type %T", x)
	}

	if prev == nil {
		prev = &protos.ObjType{
			Fields: make(map[string]*protos.Value),
		}
	}
	prev.Count++

	for _, k := range v.MapKeys() {
		kv := v.MapIndex(k)
		if k.Kind() != reflect.String {
			s.notifyErr(fmt.Errorf("invalid type %T", k.Interface()))
			continue
		}

		prevVal := prev.Fields[k.String()]
		prevVal, err := s.updateValue(prevVal, kv.Interface())
		if err != nil {
			s.notifyErr(err)
			continue
		}
		prev.Fields[k.String()] = prevVal
	}

	return prev, nil
}

func (s *St) incrFieldsProcessed() error {
	s.fieldsProcessed++

	if s.fieldsProcessed > s.maxProcessedFields {
		return fmt.Errorf("%w %d", errMaxFieldsProcessed, s.maxProcessedFields)
	}

	return nil
}

// AddSample is not thread safe
func (s *St) AddSampleData(sampleData *sample.Data) error {
	dataMap, err := sampleData.Map()
	if err != nil {
		return err
	}

	s.fieldsProcessed = 0
	updatedObj, err := s.updateObj(s.digest.GetObj(), dataMap)
	if err != nil && !errors.Is(err, errMaxFieldsProcessed) {
		return err
	}

	s.digest.Obj = updatedObj

	return nil
}

func (s *St) JSON() ([]byte, error) {
	json, err := protojson.Marshal(s.digest)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal digest to json: %w", err)
	}

	return json, nil
}

func (s *St) Reset() {
	s.digest.Reset()
}

func (s *St) String() string {
	return "StructDigest"
}
