package digest

import (
	"errors"
	"fmt"
	"math"
	"reflect"

	"github.com/neblic/platform/controlplane/control"
	"github.com/neblic/platform/dataplane/protos"
	"github.com/neblic/platform/internal/pkg/data"
	"google.golang.org/protobuf/encoding/protojson"
)

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

func (s *St) updateNum(prev *protos.NumberSt, x interface{}) (*protos.NumberSt, error) {
	if err := s.incrFieldsProcessed(); err != nil {
		return nil, err
	}

	if prev == nil {
		prev = &protos.NumberSt{}
	}

	switch num64(x).(type) {
	case int64:
		if prev.IntegerNum == nil {
			prev.IntegerNum = &protos.IntNumSt{}
		}
		prev.IntegerNum.Count++
	case uint64:
		if prev.UintegerNum == nil {
			prev.UintegerNum = &protos.UIntNumSt{}
		}
		prev.UintegerNum.Count++
	case float64:
		if prev.FloatNum == nil {
			prev.FloatNum = &protos.FloatNumSt{}
		}
		prev.FloatNum.Count++
	default:
		return nil, fmt.Errorf("invalid number type %T", x)
	}

	return prev, nil
}

func (s *St) updateString(prev *protos.StringSt, x string) (*protos.StringSt, error) {
	if err := s.incrFieldsProcessed(); err != nil {
		return nil, err
	}

	if prev == nil {
		prev = &protos.StringSt{}
	}
	prev.Count++

	return prev, nil
}

func (s *St) updateBoolean(prev *protos.BooleanSt, x bool) (*protos.BooleanSt, error) {
	if err := s.incrFieldsProcessed(); err != nil {
		return nil, err
	}

	if prev == nil {
		prev = &protos.BooleanSt{}
	}
	prev.Count++

	return prev, nil
}

func (s *St) updateValue(prevVal *protos.ValueSt, x interface{}) (*protos.ValueSt, error) {
	if err := s.incrFieldsProcessed(); err != nil {
		return nil, err
	}

	if prevVal == nil {
		prevVal = &protos.ValueSt{}
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

func (s *St) updateArray(prev *protos.ArraySt, x interface{}) (*protos.ArraySt, error) {
	if err := s.incrFieldsProcessed(); err != nil {
		return nil, err
	}

	v := reflect.ValueOf(x)
	if v.Kind() != reflect.Slice {
		return nil, fmt.Errorf("invalid type %T", x)
	}

	if prev == nil {
		prev = &protos.ArraySt{
			MinLength: math.Inf(+1),
			MaxLength: math.Inf(-1),
		}
	}
	prev.Count++

	for i := 0; i < v.Len(); i++ {
		var err error
		prev.Values, err = s.updateValue(prev.Values, v.Index(i).Interface())
		if err != nil {
			s.notifyErr(err)
		}
	}

	prev.MinLength = math.Min(float64(prev.MinLength), float64(v.Len()))
	prev.MaxLength = math.Max(float64(prev.MaxLength), float64(v.Len()))
	prev.SumLength += float64(v.Len())

	return prev, nil
}

func (s *St) updateObj(prev *protos.ObjSt, x interface{}) (*protos.ObjSt, error) {
	if err := s.incrFieldsProcessed(); err != nil {
		return nil, err
	}

	v := reflect.ValueOf(x)
	if v.Kind() != reflect.Map {
		return nil, fmt.Errorf("invalid type %T", x)
	}

	if prev == nil {
		prev = &protos.ObjSt{
			Fields: make(map[string]*protos.ValueSt),
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
func (s *St) AddSampleData(sampleData *data.Data) error {
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

func (s *St) SampleType() control.SampleType {
	return control.StructDigestSampleType
}
