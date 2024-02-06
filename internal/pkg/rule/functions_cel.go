package rule

import (
	"math"
	"time"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
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
	// sequence
	sequenceStatefulFunctionEnv,
	makeSequenceIntDummy(),
	makeSequenceInt(),
	makeSequenceUintDummy(),
	makeSequenceUint(),
	makeSequenceFloat64Dummy(),
	makeSequenceFloat64(),
	makeSequenceStringDummy(),
	makeSequenceString(),
	// complete
	completeStatefulFunctionEnv,
	makeCompleteIntDummy(),
	makeCompleteInt(),
	makeCompleteUintDummy(),
	makeCompleteUint(),
	makeCompleteFloat64Dummy(),
	makeCompleteFloat64(),
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
var sequenceStatefulFunctionEnv = cel.Types(&SequenceStatefulFunction{})
var sequenceStatefulFunctionType = cel.ObjectType("SequenceStatefulFunction")

func makeSequenceIntDummy() cel.EnvOption {
	return cel.Function("sequence",
		cel.Overload("sequence_int_string",
			[]*cel.Type{cel.IntType, cel.StringType},
			cel.BoolType,
			cel.FunctionBinding(func(args ...ref.Val) ref.Val {
				panic("called sequence int dummy function")
			}),
		),
	)
}
func makeSequenceInt() cel.EnvOption {
	return cel.Function("sequence",
		cel.Overload("sequence_int_string_state",
			[]*cel.Type{cel.IntType, cel.StringType, sequenceStatefulFunctionType},
			cel.BoolType,
			cel.FunctionBinding(func(args ...ref.Val) ref.Val {
				value := args[0].Value()
				state := args[2].Value().(*SequenceStatefulFunction)

				ok := state.Call(value)
				return types.Bool(ok)
			}),
		),
	)
}
func makeSequenceUintDummy() cel.EnvOption {
	return cel.Function("sequence",
		cel.Overload("sequence_uint_string",
			[]*cel.Type{cel.UintType, cel.StringType},
			cel.BoolType,
			cel.FunctionBinding(func(args ...ref.Val) ref.Val {
				panic("called sequence uint dummy function")
			}),
		),
	)
}
func makeSequenceUint() cel.EnvOption {
	return cel.Function("sequence",
		cel.Overload("sequence_uint_string_state",
			[]*cel.Type{cel.UintType, cel.StringType, sequenceStatefulFunctionType},
			cel.BoolType,
			cel.FunctionBinding(func(args ...ref.Val) ref.Val {
				value := args[0].Value()
				state := args[2].Value().(*SequenceStatefulFunction)

				ok := state.Call(value)
				return types.Bool(ok)
			}),
		),
	)
}
func makeSequenceFloat64Dummy() cel.EnvOption {
	return cel.Function("sequence",
		cel.Overload("sequence_double_string",
			[]*cel.Type{cel.DoubleType, cel.StringType},
			cel.BoolType,
			cel.FunctionBinding(func(args ...ref.Val) ref.Val {
				panic("called sequence double dummy function")
			}),
		),
	)
}
func makeSequenceFloat64() cel.EnvOption {
	return cel.Function("sequence",
		cel.Overload("sequence_double_string_state",
			[]*cel.Type{cel.DoubleType, cel.StringType, sequenceStatefulFunctionType},
			cel.BoolType,
			cel.FunctionBinding(func(args ...ref.Val) ref.Val {
				value := args[0].Value()
				state := args[2].Value().(*SequenceStatefulFunction)

				ok := state.Call(value)
				return types.Bool(ok)
			}),
		),
	)
}
func makeSequenceStringDummy() cel.EnvOption {
	return cel.Function("sequence",
		cel.Overload("sequence_string_string",
			[]*cel.Type{cel.StringType, cel.StringType},
			cel.BoolType,
			cel.FunctionBinding(func(args ...ref.Val) ref.Val {
				panic("called sequence string dummy function")
			}),
		),
	)
}
func makeSequenceString() cel.EnvOption {
	return cel.Function("sequence",
		cel.Overload("sequence_string_string_state",
			[]*cel.Type{cel.StringType, cel.StringType, sequenceStatefulFunctionType},
			cel.BoolType,
			cel.FunctionBinding(func(args ...ref.Val) ref.Val {
				value := args[0].Value()
				state := args[2].Value().(*SequenceStatefulFunction)

				ok := state.Call(value)
				return types.Bool(ok)
			}),
		),
	)
}

// complete (stateful)
var completeStatefulFunctionEnv = cel.Types(&CompleteStatefulFunction{})
var completeStatefulFunctionType = cel.ObjectType("CompleteStatefulFunction")

func makeCompleteIntDummy() cel.EnvOption {
	return cel.Function("complete",
		cel.Overload("complete_int_int",
			[]*cel.Type{cel.IntType, cel.IntType},
			cel.BoolType,
			cel.FunctionBinding(func(args ...ref.Val) ref.Val {
				panic("called complete int dummy function")
			}),
		),
	)
}
func makeCompleteInt() cel.EnvOption {
	return cel.Function("complete",
		cel.Overload("complete_int_int_state",
			[]*cel.Type{cel.IntType, cel.IntType, completeStatefulFunctionType},
			cel.BoolType,
			cel.FunctionBinding(func(args ...ref.Val) ref.Val {
				value := args[0].Value()
				state := args[2].Value().(*CompleteStatefulFunction)

				ok := state.Call(value)
				return types.Bool(ok)
			}),
		),
	)
}
func makeCompleteUintDummy() cel.EnvOption {
	return cel.Function("complete",
		cel.Overload("complete_uint_int",
			[]*cel.Type{cel.UintType, cel.IntType},
			cel.BoolType,
			cel.FunctionBinding(func(args ...ref.Val) ref.Val {
				panic("called complete uint dummy function")
			}),
		),
	)
}
func makeCompleteUint() cel.EnvOption {
	return cel.Function("complete",
		cel.Overload("complete_uint_int_state",
			[]*cel.Type{cel.UintType, cel.IntType, completeStatefulFunctionType},
			cel.BoolType,
			cel.FunctionBinding(func(args ...ref.Val) ref.Val {
				value := args[0].Value()
				state := args[2].Value().(*CompleteStatefulFunction)

				ok := state.Call(value)
				return types.Bool(ok)
			}),
		),
	)
}
func makeCompleteFloat64Dummy() cel.EnvOption {
	return cel.Function("complete",
		cel.Overload("complete_double_double",
			[]*cel.Type{cel.DoubleType, cel.DoubleType},
			cel.BoolType,
			cel.FunctionBinding(func(args ...ref.Val) ref.Val {
				panic("called complete double dummy function")
			}),
		),
	)
}
func makeCompleteFloat64() cel.EnvOption {
	return cel.Function("complete",
		cel.Overload("complete_double_double_state",
			[]*cel.Type{cel.DoubleType, cel.DoubleType, completeStatefulFunctionType},
			cel.BoolType,
			cel.FunctionBinding(func(args ...ref.Val) ref.Val {
				value := args[0].Value()
				state := args[2].Value().(*CompleteStatefulFunction)

				ok := state.Call(value)
				return types.Bool(ok)
			}),
		),
	)
}
