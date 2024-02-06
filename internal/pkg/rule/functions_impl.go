package rule

import (
	"fmt"
	"reflect"

	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
	"golang.org/x/exp/constraints"
)

type StatefulFunction interface {
	ref.Type
	ref.Val
	StateName() string
	Call(value any) bool
}

type OrderType int8

const (
	OrderTypeUnknown OrderType = iota
	OrderTypeNone
	OrderTypeAsc
	OrderTypeDesc
)

// SequenceStatefulFunctionOf
type SequenceStatefulFunctionOf[T constraints.Ordered] struct {
	last          *T
	expectedOrder OrderType
	Order         OrderType
}

func NewSequenceStatefulFunctionOf[T constraints.Ordered](orderType OrderType) *SequenceStatefulFunctionOf[T] {
	return &SequenceStatefulFunctionOf[T]{
		expectedOrder: orderType,
		Order:         orderType,
	}
}

func (ss *SequenceStatefulFunctionOf[T]) Call(value T) bool {
	if ss.last == nil {
		ss.last = &value
		return true
	}

	isOrdered := true
	switch ss.expectedOrder {
	case OrderTypeAsc:
		if value < *ss.last {
			isOrdered = false
			ss.Order = OrderTypeNone
		}
	case OrderTypeDesc:
		if value > *ss.last {
			isOrdered = false
			ss.Order = OrderTypeNone
		}
	}

	ss.last = &value

	return isOrdered
}

// SequenceStatefulFunction
type SequenceStatefulFunction struct {
	stateName string
	order     OrderType
	ofInt64   *SequenceStatefulFunctionOf[int64]
	ofUint64  *SequenceStatefulFunctionOf[uint64]
	ofFloat64 *SequenceStatefulFunctionOf[float64]
	ofString  *SequenceStatefulFunctionOf[string]
}

func (ssf *SequenceStatefulFunction) HasTrait(trait int) bool {
	return false
}

func (ssf *SequenceStatefulFunction) TypeName() string {
	return "SequenceStatefulFunction"
}

func (ssf *SequenceStatefulFunction) ConvertToNative(typeDesc reflect.Type) (any, error) {
	return nil, fmt.Errorf("unsupported type conversion from '%s' to %v", sequenceStatefulFunctionType, typeDesc)
}

// ConvertToType supports type conversions between CEL value types supported by the expression language.
func (ssf *SequenceStatefulFunction) ConvertToType(typeValue ref.Type) ref.Val {
	return types.NewErr("type conversion error from '%s' to '%s", sequenceStatefulFunctionType, typeValue)
}

// Equal returns true if the `other` value has the same type and content as the implementing struct.
func (ssf *SequenceStatefulFunction) Equal(other ref.Val) ref.Val {
	ov, ok := other.(*SequenceStatefulFunction)
	if !ok {
		return types.False
	}

	return types.Bool(ssf.stateName != ov.stateName && ssf.order == ov.order)
}

// Type returns the TypeValue of the value.
func (ssf *SequenceStatefulFunction) Type() ref.Type {
	return sequenceStatefulFunctionType
}

// Value returns the raw value of the instance which may not be directly compatible with the expression
// language types.
func (ssf *SequenceStatefulFunction) Value() any {
	return ssf
}

func (ssf *SequenceStatefulFunction) StateName() string {
	return ssf.stateName
}

func (ssf *SequenceStatefulFunction) Call(value any) bool {
	var returnValue bool

	switch v := value.(type) {
	case int64:
		if ssf.ofInt64 == nil {
			ssf.ofInt64 = NewSequenceStatefulFunctionOf[int64](ssf.order)
		}
		returnValue = ssf.ofInt64.Call(v)
	case uint64:
		if ssf.ofUint64 == nil {
			ssf.ofUint64 = NewSequenceStatefulFunctionOf[uint64](ssf.order)
		}
		returnValue = ssf.ofUint64.Call(v)
	case float64:
		if ssf.ofFloat64 == nil {
			ssf.ofFloat64 = NewSequenceStatefulFunctionOf[float64](ssf.order)
		}
		returnValue = ssf.ofFloat64.Call(v)
	case string:
		if ssf.ofString == nil {
			ssf.ofString = NewSequenceStatefulFunctionOf[string](ssf.order)
		}
		returnValue = ssf.ofString.Call(v)
	default:
		panic(fmt.Sprintf("sequence function requires a value of type int64, uint64, float64 or string. Type %T is not supported", value))
	}

	return returnValue
}

// CompleteStatefulFunctionOf
type Number interface {
	constraints.Float | constraints.Integer
}

type CompleteStatefulFunctionOf[T Number] struct {
	next        *T
	step        T
	AllComplete bool
}

func NewCompleteStatefulFunctionOf[T Number](step T) *CompleteStatefulFunctionOf[T] {
	return &CompleteStatefulFunctionOf[T]{
		step:        step,
		AllComplete: true,
	}
}

func (ss *CompleteStatefulFunctionOf[T]) Call(value T) bool {
	if ss.next == nil {
		ss.next = new(T)
		*ss.next = value
	}

	isComplete := value == *ss.next
	if !isComplete {
		ss.AllComplete = false
	}

	*ss.next = value + ss.step

	return isComplete
}

// CompleteStatefulFunction
type CompleteStatefulFunction struct {
	stateName string
	step      float64
	ofInt64   *CompleteStatefulFunctionOf[int64]
	ofUint64  *CompleteStatefulFunctionOf[uint64]
	ofFloat64 *CompleteStatefulFunctionOf[float64]
}

func (csf *CompleteStatefulFunction) HasTrait(trait int) bool {
	return false
}

func (csf *CompleteStatefulFunction) TypeName() string {
	return "CompleteStatefulFunction"
}

func (csf *CompleteStatefulFunction) ConvertToNative(typeDesc reflect.Type) (any, error) {
	return nil, fmt.Errorf("unsupported type conversion from '%s' to %v", completeStatefulFunctionType, typeDesc)
}

// ConvertToType supports type conversions between CEL value types supported by the expression language.
func (csf *CompleteStatefulFunction) ConvertToType(typeValue ref.Type) ref.Val {
	return types.NewErr("type conversion error from '%s' to '%s", completeStatefulFunctionType, typeValue)
}

// Equal returns true if the `other` value has the same type and content as the implementing struct.
func (csf *CompleteStatefulFunction) Equal(other ref.Val) ref.Val {
	ov, ok := other.(*CompleteStatefulFunction)
	if !ok {
		return types.False
	}

	return types.Bool(csf.stateName != ov.stateName && csf.step == ov.step)
}

// Type returns the TypeValue of the value.
func (csf *CompleteStatefulFunction) Type() ref.Type {
	return completeStatefulFunctionType
}

// Value returns the raw value of the instance which may not be directly compatible with the expression
// language types.
func (csf *CompleteStatefulFunction) Value() any {
	return csf
}

func (csf *CompleteStatefulFunction) StateName() string {
	return csf.stateName
}

func (csf *CompleteStatefulFunction) Call(value any) bool {
	var returnValue bool

	switch v := value.(type) {
	case int64:
		if csf.ofInt64 == nil {
			csf.ofInt64 = NewCompleteStatefulFunctionOf[int64](int64(csf.step))
		}
		returnValue = csf.ofInt64.Call(v)
	case uint64:
		if csf.ofUint64 == nil {
			csf.ofUint64 = NewCompleteStatefulFunctionOf[uint64](uint64(csf.step))
		}
		returnValue = csf.ofUint64.Call(v)
	case float64:
		if csf.ofFloat64 == nil {
			csf.ofFloat64 = NewCompleteStatefulFunctionOf[float64](csf.step)
		}
		returnValue = csf.ofFloat64.Call(v)
	default:
		panic(fmt.Sprintf("complete function requires a value of type int64, uint64 or float64. Type %T is not supported", value))
	}

	return returnValue
}
