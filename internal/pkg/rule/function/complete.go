package function

import (
	"fmt"
	"reflect"

	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
	"golang.org/x/exp/constraints"
)

// CompleteStateOf
type Number interface {
	constraints.Float | constraints.Integer
}

type CompleteParameters struct {
	Step float64
}

type CompleteStateOf[T Number] struct {
	Next        *T
	Step        T
	AllComplete bool
}

func NewCompleteStateOf[T Number](parameters *CompleteParameters) *CompleteStateOf[T] {
	return &CompleteStateOf[T]{
		Step:        T(parameters.Step),
		AllComplete: true,
	}
}

func CallComplete[T Number](cs *CompleteStateOf[T], value T) bool {
	if cs.Next == nil {
		cs.Next = new(T)
		*cs.Next = value
	}

	isComplete := value == *cs.Next
	if !isComplete {
		cs.AllComplete = false
	}

	*cs.Next = value + cs.Step

	return isComplete
}

type CompleteState struct {
	ofInt64   *CompleteStateOf[int64]
	ofUint64  *CompleteStateOf[uint64]
	ofFloat64 *CompleteStateOf[float64]
}

// CompleteStatefulFunction
type CompleteStatefulFunction struct {
	Parameters *CompleteParameters
	State      *CompleteState
}

func NewCompleteStatefulFunction(parameters *CompleteParameters, state *CompleteState) *CompleteStatefulFunction {
	return &CompleteStatefulFunction{
		Parameters: parameters,
		State:      state,
	}
}

func (csf *CompleteStatefulFunction) HasTrait(_ int) bool {
	return false
}

func (csf *CompleteStatefulFunction) TypeName() string {
	return "CompleteStatefulFunction"
}

func (csf *CompleteStatefulFunction) ConvertToNative(typeDesc reflect.Type) (any, error) {
	return nil, fmt.Errorf("unsupported type conversion from '%s' to %v", CompleteStatefulFunctionType, typeDesc)
}

// ConvertToType supports type conversions between CEL value types supported by the expression language.
func (csf *CompleteStatefulFunction) ConvertToType(typeValue ref.Type) ref.Val {
	return types.NewErr("type conversion error from '%s' to '%s", CompleteStatefulFunctionType, typeValue)
}

// Equal returns true if the `other` value has the same type and content as the implementing struct.
func (csf *CompleteStatefulFunction) Equal(_ ref.Val) ref.Val {
	return types.NewErr("equal operation not supported")
}

// Type returns the TypeValue of the value.
func (csf *CompleteStatefulFunction) Type() ref.Type {
	return CompleteStatefulFunctionType
}

// Value returns the raw value of the instance which may not be directly compatible with the expression
// language types.
func (csf *CompleteStatefulFunction) Value() any {
	return csf
}

func (csf *CompleteStatefulFunction) Call(value any) bool {
	var returnValue bool

	switch v := value.(type) {
	case int64:
		if csf.State.ofInt64 == nil {
			csf.State.ofInt64 = NewCompleteStateOf[int64](csf.Parameters)
		}
		returnValue = CallComplete[int64](csf.State.ofInt64, v)
	case uint64:
		if csf.State.ofUint64 == nil {
			csf.State.ofUint64 = NewCompleteStateOf[uint64](csf.Parameters)
		}
		returnValue = CallComplete[uint64](csf.State.ofUint64, v)
	case float64:
		if csf.State.ofFloat64 == nil {
			csf.State.ofFloat64 = NewCompleteStateOf[float64](csf.Parameters)
		}
		returnValue = CallComplete[float64](csf.State.ofFloat64, v)
	default:
		panic(fmt.Sprintf("complete function requires a value of type int64, uint64 or float64. Type %T is not supported", value))
	}

	return returnValue
}
