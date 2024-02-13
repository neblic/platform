package function

import (
	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
)

// sequence (stateful)
var SequenceStatefulFunctionEnv = cel.Types(&SequenceStatefulFunction{})
var SequenceStatefulFunctionType = cel.ObjectType("SequenceStatefulFunction")

func MakeSequenceIntDummy() cel.EnvOption {
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
func MakeSequenceInt() cel.EnvOption {
	return cel.Function("sequence",
		cel.Overload("sequence_int_string_state",
			[]*cel.Type{cel.IntType, cel.StringType, SequenceStatefulFunctionType},
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
func MakeSequenceUintDummy() cel.EnvOption {
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
func MakeSequenceUint() cel.EnvOption {
	return cel.Function("sequence",
		cel.Overload("sequence_uint_string_state",
			[]*cel.Type{cel.UintType, cel.StringType, SequenceStatefulFunctionType},
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
func MakeSequenceFloat64Dummy() cel.EnvOption {
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
func MakeSequenceFloat64() cel.EnvOption {
	return cel.Function("sequence",
		cel.Overload("sequence_double_string_state",
			[]*cel.Type{cel.DoubleType, cel.StringType, SequenceStatefulFunctionType},
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
func MakeSequenceStringDummy() cel.EnvOption {
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
func MakeSequenceString() cel.EnvOption {
	return cel.Function("sequence",
		cel.Overload("sequence_string_string_state",
			[]*cel.Type{cel.StringType, cel.StringType, SequenceStatefulFunctionType},
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
