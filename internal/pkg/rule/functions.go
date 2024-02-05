package rule

import (
	"fmt"
	"math"
	"time"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
	expr "google.golang.org/genproto/googleapis/api/expr/v1alpha1"
)

// StreamFunctionsEnvOptions contains all the custom functions used when defining stream rules.
var StreamFunctionsEnvOptions = []cel.EnvOption{
	// abs
	absDoubleFunc,
	absIntFunc,
	// now
	nowFunc,
}

// CheckFunctionsEnvOptions contains all the custom functions used when defining check rules.
// Stateful functions added here are dummy functions that are overloaded on-demand with a
// correctly initialized state when a rule is created.
var CheckFunctionsEnvOptions = []cel.EnvOption{
	// abs
	absDoubleFunc,
	absIntFunc,
	// now
	nowFunc,
	// sequence (dummy)
	makeSequenceInt(nil),
	makeSequenceUint(nil),
	makeSequenceFloat64(nil),
	makeSequenceString(nil),
	// complete (dummy)
	makeCompleteInt(nil),
	makeCompleteUint(nil),
	makeCompleteFloat64(nil),
}

// abs
var absDoubleFunc = cel.Function("abs",
	cel.Overload("abs_double",
		[]*cel.Type{cel.DoubleType},
		cel.DoubleType,
		cel.UnaryBinding(func(value ref.Val) ref.Val {
			return types.Double(math.Abs(value.Value().(float64)))
		}),
	),
)
var absIntFunc = cel.Function("abs",
	cel.Overload("abs_int",
		[]*cel.Type{cel.IntType},
		cel.IntType,
		cel.UnaryBinding(func(value ref.Val) ref.Val {
			intValue := value.Value().(int64)
			if intValue < 0 {
				intValue = -intValue
			}
			return types.Int(intValue)
		}),
	),
)

// now
var nowFunc = cel.Function("now",
	cel.Overload("now",
		[]*cel.Type{},
		cel.TimestampType,
		cel.FunctionBinding(func(...ref.Val) ref.Val {
			return types.Timestamp{Time: time.Now()}
		}),
	),
)

// sequence (stateful)
type SequenceStateProvider interface {
	SequenceAddInt64(int64) bool
	SequenceAddUint64(uint64) bool
	SequenceAddFloat64(float64) bool
	SequenceAddString(string) bool
}

func makeSequenceInt(stateProvider SequenceStateProvider) cel.EnvOption {
	return cel.Function("sequence",
		cel.Overload("sequence_int_string",
			[]*cel.Type{cel.IntType, cel.StringType},
			cel.BoolType,
			cel.FunctionBinding(func(args ...ref.Val) ref.Val {
				value := args[0].Value().(int64)

				ok := stateProvider.SequenceAddInt64(value)
				return types.Bool(ok)
			}),
		),
	)
}
func makeSequenceUint(stateProvider SequenceStateProvider) cel.EnvOption {
	return cel.Function("sequence",
		cel.Overload("sequence_uint_string",
			[]*cel.Type{cel.UintType, cel.StringType},
			cel.BoolType,
			cel.FunctionBinding(func(args ...ref.Val) ref.Val {
				value := args[0].Value().(uint64)

				ok := stateProvider.SequenceAddUint64(value)
				return types.Bool(ok)
			}),
		),
	)
}
func makeSequenceFloat64(stateProvider SequenceStateProvider) cel.EnvOption {
	return cel.Function("sequence",
		cel.Overload("sequence_double_string",
			[]*cel.Type{cel.DoubleType, cel.StringType},
			cel.BoolType,
			cel.FunctionBinding(func(args ...ref.Val) ref.Val {
				value := args[0].Value().(float64)

				ok := stateProvider.SequenceAddFloat64(value)
				return types.Bool(ok)
			}),
		),
	)
}
func makeSequenceString(stateProvider SequenceStateProvider) cel.EnvOption {
	return cel.Function("sequence",
		cel.Overload("sequence_string_string",
			[]*cel.Type{cel.StringType, cel.StringType},
			cel.BoolType,
			cel.FunctionBinding(func(args ...ref.Val) ref.Val {
				value := args[0].Value().(string)

				ok := stateProvider.SequenceAddString(value)
				return types.Bool(ok)
			}),
		),
	)
}

type SequenceStatefulFunction struct {
	enabled bool
	Order   OrderType
}

func (ssf *SequenceStatefulFunction) GetName() string {
	return "sequence"
}

func (ssf *SequenceStatefulFunction) Enabled() bool {
	return ssf.enabled
}

func (ssf *SequenceStatefulFunction) ParseCallExpression(callExpression *expr.Expr_Call) error {
	if ssf.enabled {
		return fmt.Errorf("expression already contains a sequence function. Having multiple stateful functions of the same type is not supported")
	}
	ssf.enabled = true

	var err error
	ssf.Order, err = getOrderFromExpression(callExpression.Args[1])
	if err != nil {
		return err
	}

	return nil
}

func (ssf *SequenceStatefulFunction) GetCelEnvs(stateProvider *StateProvider) []cel.EnvOption {
	stateProvider.WithSequenceState(ssf.Order)
	return []cel.EnvOption{
		makeSequenceInt(stateProvider),
		makeSequenceUint(stateProvider),
		makeSequenceFloat64(stateProvider),
		makeSequenceString(stateProvider),
	}
}

// complete (stateful)
type CompleteStateProvider interface {
	CompleteAddInt64(int64) bool
	CompleteAddUint64(uint64) bool
	CompleteAddFloat64(float64) bool
}

func makeCompleteInt(stateProvider CompleteStateProvider) cel.EnvOption {
	return cel.Function("complete",
		cel.Overload("complete_int_int",
			[]*cel.Type{cel.IntType, cel.IntType},
			cel.BoolType,
			cel.FunctionBinding(func(args ...ref.Val) ref.Val {
				value := args[0].Value().(int64)

				ok := stateProvider.CompleteAddInt64(value)
				return types.Bool(ok)
			}),
		),
	)
}
func makeCompleteUint(stateProvider CompleteStateProvider) cel.EnvOption {
	return cel.Function("complete",
		cel.Overload("complete_uint_int",
			[]*cel.Type{cel.UintType, cel.IntType},
			cel.BoolType,
			cel.FunctionBinding(func(args ...ref.Val) ref.Val {
				value := args[0].Value().(uint64)

				ok := stateProvider.CompleteAddUint64(value)
				return types.Bool(ok)
			}),
		),
	)
}
func makeCompleteFloat64(stateProvider CompleteStateProvider) cel.EnvOption {
	return cel.Function("complete",
		cel.Overload("complete_double_double",
			[]*cel.Type{cel.DoubleType, cel.DoubleType},
			cel.BoolType,
			cel.FunctionBinding(func(args ...ref.Val) ref.Val {
				value := args[0].Value().(float64)

				ok := stateProvider.CompleteAddFloat64(value)
				return types.Bool(ok)
			}),
		),
	)
}

type CompleteStatefulFunction struct {
	enabled bool
	Step    float64
}

func (csf *CompleteStatefulFunction) GetName() string {
	return "complete"
}

func (csf *CompleteStatefulFunction) Enabled() bool {
	return csf.enabled
}

func (csf *CompleteStatefulFunction) ParseCallExpression(callExpression *expr.Expr_Call) error {
	if csf.enabled {
		return fmt.Errorf("expression already contains a complete function. Having multiple stateful functions of the same type is not supported")
	}
	csf.enabled = true

	var err error
	csf.Step, err = getStepFromExpression(callExpression.Args[1])
	if err != nil {
		return err
	}

	return nil
}

func (csf *CompleteStatefulFunction) GetCelEnvs(stateProvider *StateProvider) []cel.EnvOption {
	stateProvider.WithCompleteState(csf.Step)
	return []cel.EnvOption{
		makeCompleteInt(stateProvider),
		makeCompleteUint(stateProvider),
		makeCompleteFloat64(stateProvider),
	}
}
