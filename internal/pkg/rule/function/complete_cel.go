package function

import (
	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
)

// complete (stateful)
var CompleteStatefulFunctionEnv = cel.Types(&CompleteStatefulFunction{})
var CompleteStatefulFunctionType = cel.ObjectType("CompleteStatefulFunction")

func MakeCompleteIntDummy() cel.EnvOption {
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
func MakeCompleteInt() cel.EnvOption {
	return cel.Function("complete",
		cel.Overload("complete_int_int_state",
			[]*cel.Type{cel.IntType, cel.IntType, CompleteStatefulFunctionType},
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
func MakeCompleteUintDummy() cel.EnvOption {
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
func MakeCompleteUint() cel.EnvOption {
	return cel.Function("complete",
		cel.Overload("complete_uint_int_state",
			[]*cel.Type{cel.UintType, cel.IntType, CompleteStatefulFunctionType},
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
func MakeCompleteFloat64Dummy() cel.EnvOption {
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
func MakeCompleteFloat64() cel.EnvOption {
	return cel.Function("complete",
		cel.Overload("complete_double_double_state",
			[]*cel.Type{cel.DoubleType, cel.DoubleType, CompleteStatefulFunctionType},
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
