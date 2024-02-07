package function

import "github.com/google/cel-go/common/types/ref"

type StatefulFunction interface {
	ref.Type
	ref.Val
	Call(value any) bool
}

type StatefulFunctionProvider struct {
	StateName               string
	statefulFunctionBuilder func(state any) StatefulFunction
	stateBuilder            func() any
	globalStatefulFunction  StatefulFunction
	keyedStates             map[string]any
}

func NewStatefulFunctionProvider(stateName string, statefulFunctionBuilder func(state any) StatefulFunction, stateBuilder func() any) *StatefulFunctionProvider {
	return &StatefulFunctionProvider{
		StateName:               stateName,
		statefulFunctionBuilder: statefulFunctionBuilder,
		stateBuilder:            stateBuilder,
		keyedStates:             make(map[string]any),
	}
}

func (sfp *StatefulFunctionProvider) GlobalStatefulFunction() StatefulFunction {
	if sfp.globalStatefulFunction == nil {
		sfp.globalStatefulFunction = sfp.statefulFunctionBuilder(sfp.stateBuilder())
	}
	return sfp.globalStatefulFunction
}

func (sfp *StatefulFunctionProvider) KeyedStatefulFunction(key string) StatefulFunction {
	state, ok := sfp.keyedStates[key]
	if !ok {
		state = sfp.stateBuilder()
		sfp.keyedStates[key] = state
	}
	return sfp.statefulFunctionBuilder(state)
}
