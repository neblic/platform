package function

import (
	"fmt"
	"reflect"

	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
	"golang.org/x/exp/constraints"
)

type OrderType int8

const (
	OrderTypeUnknown OrderType = iota
	OrderTypeNone
	OrderTypeAsc
	OrderTypeDesc
)

type SequenceParameters struct {
	Order OrderType
}

// SequenceState
type SequenceStateOf[T constraints.Ordered] struct {
	Last          *T
	ExpectedOrder OrderType
	ResultOrder   OrderType
}

func NewSequenceStateOf[T constraints.Ordered](parameters *SequenceParameters) *SequenceStateOf[T] {
	return &SequenceStateOf[T]{
		ExpectedOrder: parameters.Order,
		ResultOrder:   parameters.Order,
	}
}

func CallSequence[T constraints.Ordered](ss *SequenceStateOf[T], parameters *SequenceParameters, value T) bool {
	if ss.Last == nil {
		ss.Last = &value
		return true
	}

	isOrdered := true
	switch ss.ExpectedOrder {
	case OrderTypeAsc:
		if value < *ss.Last {
			isOrdered = false
			ss.ResultOrder = OrderTypeNone
		}
	case OrderTypeDesc:
		if value > *ss.Last {
			isOrdered = false
			ss.ResultOrder = OrderTypeNone
		}
	}

	ss.Last = &value

	return isOrdered
}

type SequenceState struct {
	ofInt64   *SequenceStateOf[int64]
	ofUint64  *SequenceStateOf[uint64]
	ofFloat64 *SequenceStateOf[float64]
	ofString  *SequenceStateOf[string]
}

// SequenceStatefulFunction
type SequenceStatefulFunction struct {
	Parameters *SequenceParameters
	State      *SequenceState
}

func NewSequenceStatefulFunction(parameters *SequenceParameters, state *SequenceState) *SequenceStatefulFunction {
	return &SequenceStatefulFunction{
		Parameters: parameters,
		State:      state,
	}
}

func (ssf *SequenceStatefulFunction) HasTrait(trait int) bool {
	return false
}

func (ssf *SequenceStatefulFunction) TypeName() string {
	return "SequenceStatefulFunction"
}

func (ssf *SequenceStatefulFunction) ConvertToNative(typeDesc reflect.Type) (any, error) {
	return nil, fmt.Errorf("unsupported type conversion from '%s' to %v", SequenceStatefulFunctionType, typeDesc)
}

// ConvertToType supports type conversions between CEL value types supported by the expression language.
func (ssf *SequenceStatefulFunction) ConvertToType(typeValue ref.Type) ref.Val {
	return types.NewErr("type conversion error from '%s' to '%s", SequenceStatefulFunctionType, typeValue)
}

// Equal returns true if the `other` value has the same type and content as the implementing struct.
func (ssf *SequenceStatefulFunction) Equal(other ref.Val) ref.Val {
	return types.NewErr("equal operation not supported")
}

// Type returns the TypeValue of the value.
func (ssf *SequenceStatefulFunction) Type() ref.Type {
	return SequenceStatefulFunctionType
}

// Value returns the raw value of the instance which may not be directly compatible with the expression
// language types.
func (ssf *SequenceStatefulFunction) Value() any {
	return ssf
}

func (ssf *SequenceStatefulFunction) Call(value any) bool {
	var returnValue bool

	switch v := value.(type) {
	case int64:
		if ssf.State.ofInt64 == nil {
			ssf.State.ofInt64 = NewSequenceStateOf[int64](ssf.Parameters)
		}
		returnValue = CallSequence[int64](ssf.State.ofInt64, ssf.Parameters, v)
	case uint64:
		if ssf.State.ofUint64 == nil {
			ssf.State.ofUint64 = NewSequenceStateOf[uint64](ssf.Parameters)
		}
		returnValue = CallSequence[uint64](ssf.State.ofUint64, ssf.Parameters, v)
	case float64:
		if ssf.State.ofFloat64 == nil {
			ssf.State.ofFloat64 = NewSequenceStateOf[float64](ssf.Parameters)
		}
		returnValue = CallSequence[float64](ssf.State.ofFloat64, ssf.Parameters, v)
	case string:
		if ssf.State.ofString == nil {
			ssf.State.ofString = NewSequenceStateOf[string](ssf.Parameters)
		}
		returnValue = CallSequence[string](ssf.State.ofString, ssf.Parameters, v)
	default:
		panic(fmt.Sprintf("sequence function requires a value of type int64, uint64, float64 or string. Type %T is not supported", value))
	}

	return returnValue
}
