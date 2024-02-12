package rule

import (
	"fmt"

	"github.com/google/cel-go/cel"
	"github.com/neblic/platform/internal/pkg/rule/function"
	expr "google.golang.org/genproto/googleapis/api/expr/v1alpha1"
)

// StreamFunctionsEnvOptions contains all the custom functions used when defining stream rules.
var StreamFunctionsEnvOptions = []cel.EnvOption{
	// abs
	function.AbsDouble,
	function.AbsInt,
	// now
	function.Now,
}

// CheckFunctionsEnvOptions contains all the custom functions used when defining check rules.
// Stateful functions added here are dummy functions that are overloaded on-demand with a
// correctly initialized state when a rule is created.
var CheckFunctionsEnvOptions = []cel.EnvOption{
	// abs
	function.AbsDouble,
	function.AbsInt,
	// now
	function.Now,
	// sequence
	function.SequenceStatefulFunctionEnv,
	function.MakeSequenceIntDummy(),
	function.MakeSequenceInt(),
	function.MakeSequenceUintDummy(),
	function.MakeSequenceUint(),
	function.MakeSequenceFloat64Dummy(),
	function.MakeSequenceFloat64(),
	function.MakeSequenceStringDummy(),
	function.MakeSequenceString(),
	// complete
	function.CompleteStatefulFunctionEnv,
	function.MakeCompleteIntDummy(),
	function.MakeCompleteInt(),
	function.MakeCompleteUintDummy(),
	function.MakeCompleteUint(),
	function.MakeCompleteFloat64Dummy(),
	function.MakeCompleteFloat64(),
}

func getStepFromExpression(expression *expr.Expr) (float64, error) {
	constExpr, ok := expression.ExprKind.(*expr.Expr_ConstExpr)
	if !ok {
		return 0, fmt.Errorf("complete function second argument requires a constant float")
	}

	var step float64
	switch v := constExpr.ConstExpr.ConstantKind.(type) {
	case *expr.Constant_Int64Value:
		step = float64(v.Int64Value)
	case *expr.Constant_DoubleValue:
		step = v.DoubleValue
	default:
		return 0, fmt.Errorf("complete function second argument requires a constant float or int")
	}

	return step, nil
}

func getOrderFromExpression(expression *expr.Expr) (function.OrderType, error) {
	constExpr, ok := expression.ExprKind.(*expr.Expr_ConstExpr)
	if !ok {
		return function.OrderTypeUnknown, fmt.Errorf("sequence function second argument requires a constant string with value 'asc' or 'desc'")
	}

	stringExpr, ok := constExpr.ConstExpr.ConstantKind.(*expr.Constant_StringValue)
	if !ok {
		return function.OrderTypeUnknown, fmt.Errorf("sequence function second argument requires a constant string with value 'asc' or 'desc'")
	}

	var order function.OrderType
	switch stringExpr.StringValue {
	case "asc":
		order = function.OrderTypeAsc
	case "desc":
		order = function.OrderTypeDesc
	default:
		return function.OrderTypeUnknown, fmt.Errorf("sequence function second argument requires a constant string with value 'asc' or 'desc'")
	}

	return order, nil
}

type CheckedExprModifier struct {
	currentStateID int64
	currentExprID  int64
	CheckedExpr    *expr.CheckedExpr
}

func NewCheckedExprModifier(checkedExpr *expr.CheckedExpr) *CheckedExprModifier {
	id := int64(0)
	for i := range checkedExpr.ReferenceMap {
		if i > id {
			id = i
		}
	}
	id++

	return &CheckedExprModifier{
		currentStateID: 0,
		currentExprID:  id,
		CheckedExpr:    checkedExpr,
	}
}

func (cem *CheckedExprModifier) injectStateToCall(exprID int64, callExpr *expr.Expr_Call, messageType string) string {
	stateName := "state" + fmt.Sprintf("%d", cem.currentStateID)

	argumentExpression := &expr.Expr{
		Id: cem.currentExprID,
		ExprKind: &expr.Expr_IdentExpr{
			IdentExpr: &expr.Expr_Ident{
				Name: stateName,
			},
		},
	}

	// Add state argument
	callExpr.Args = append(callExpr.Args, argumentExpression)

	// Add expression id to type map
	cem.CheckedExpr.TypeMap[cem.currentExprID] = &expr.Type{
		TypeKind: &expr.Type_MessageType{
			MessageType: messageType,
		},
	}

	// Modify overloads in the reference map
	for i, overload := range cem.CheckedExpr.ReferenceMap[exprID].OverloadId {
		cem.CheckedExpr.ReferenceMap[exprID].OverloadId[i] = overload + "_state"
	}

	cem.currentStateID++
	cem.currentExprID++

	return stateName
}

func (cem *CheckedExprModifier) InjectState() ([]*function.StatefulFunctionProvider, error) {
	return cem.injectState(cem.CheckedExpr.Expr)
}

func (cem *CheckedExprModifier) injectState(expression *expr.Expr) ([]*function.StatefulFunctionProvider, error) {
	if expression == nil {
		return []*function.StatefulFunctionProvider{}, nil
	}

	providers := []*function.StatefulFunctionProvider{}
	switch expression.ExprKind.(type) {
	case *expr.Expr_SelectExpr:
		selectExpr := expression.GetSelectExpr()

		// Recursive check of the operand
		selectStatefulFunctions, err := cem.injectState(selectExpr.Operand)
		if err != nil {
			return []*function.StatefulFunctionProvider{}, err
		}
		providers = append(providers, selectStatefulFunctions...)

	case *expr.Expr_CallExpr:
		callExpr := expression.GetCallExpr()

		switch callExpr.Function {
		case "sequence":
			order, err := getOrderFromExpression(callExpr.Args[1])
			if err != nil {
				return []*function.StatefulFunctionProvider{}, err
			}

			stateName := cem.injectStateToCall(expression.Id, callExpr, function.SequenceStatefulFunctionType.String())

			// Add stateful function
			statefulFunctionParameters := &function.SequenceParameters{
				Order: order,
			}
			statefulFunctionBuilder := func(state any) function.StatefulFunction {
				return function.NewSequenceStatefulFunction(statefulFunctionParameters, state.(*function.SequenceState))
			}
			stateBuilder := func() any {
				return &function.SequenceState{}
			}

			providers = append(providers, function.NewStatefulFunctionProvider(stateName, statefulFunctionBuilder, stateBuilder))

		case "complete":
			step, err := getStepFromExpression(callExpr.Args[1])
			if err != nil {
				return []*function.StatefulFunctionProvider{}, err
			}

			stateName := cem.injectStateToCall(expression.Id, callExpr, function.CompleteStatefulFunctionType.String())

			// Add stateful function
			statefulFunctionParameters := &function.CompleteParameters{
				Step: step,
			}
			statefulFunctionBuilder := func(state any) function.StatefulFunction {
				return function.NewCompleteStatefulFunction(statefulFunctionParameters, state.(*function.CompleteState))
			}
			stateBuilder := func() any {
				return &function.CompleteState{}
			}

			providers = append(providers, function.NewStatefulFunctionProvider(stateName, statefulFunctionBuilder, stateBuilder))
		}

		// Recursive check of the target
		targetStatefulFunctions, err := cem.injectState(callExpr.Target)
		if err != nil {
			return []*function.StatefulFunctionProvider{}, err
		}
		providers = append(providers, targetStatefulFunctions...)
		// Recursive check of the arguments
		for _, arg := range callExpr.Args {
			argsStatefulFunctions, err := cem.injectState(arg)
			if err != nil {
				return []*function.StatefulFunctionProvider{}, err
			}
			providers = append(providers, argsStatefulFunctions...)
		}

	case *expr.Expr_ListExpr:
		listExpr := expression.GetListExpr()

		// Recursive check of the elements
		for _, element := range listExpr.Elements {
			elementSatatefulFunctions, err := cem.injectState(element)
			if err != nil {
				return []*function.StatefulFunctionProvider{}, err
			}
			providers = append(providers, elementSatatefulFunctions...)
		}

	case *expr.Expr_StructExpr:
		structExpr := expression.GetStructExpr()

		// Recursive check of the entries
		for _, entry := range structExpr.Entries {
			structStatefulFunction, err := cem.injectState(entry.Value)
			if err != nil {
				return []*function.StatefulFunctionProvider{}, err
			}
			providers = append(providers, structStatefulFunction...)
		}

	case *expr.Expr_ComprehensionExpr:
		comprehensionExpr := expression.GetComprehensionExpr()

		// Recursive check of the itereration range
		iterRangeStatefulFunction, err := cem.injectState(comprehensionExpr.IterRange)
		if err != nil {
			return []*function.StatefulFunctionProvider{}, err
		}
		providers = append(providers, iterRangeStatefulFunction...)
		// Recursive check of the accumulator initialization
		accuIntStatefulFunctions, err := cem.injectState(comprehensionExpr.AccuInit)
		if err != nil {
			return []*function.StatefulFunctionProvider{}, err
		}
		providers = append(providers, accuIntStatefulFunctions...)
		// Recursive check of the loop condition
		loopConditionStatefulFunctions, err := cem.injectState(comprehensionExpr.LoopCondition)
		if err != nil {
			return []*function.StatefulFunctionProvider{}, err
		}
		providers = append(providers, loopConditionStatefulFunctions...)
		// Recursive check of the loop step
		loopStepStatefulFunctions, err := cem.injectState(comprehensionExpr.LoopStep)
		if err != nil {
			return []*function.StatefulFunctionProvider{}, err
		}
		providers = append(providers, loopStepStatefulFunctions...)
		// Recursive check of the result
		resultStatefulFunctions, err := cem.injectState(comprehensionExpr.Result)
		if err != nil {
			return []*function.StatefulFunctionProvider{}, err
		}
		providers = append(providers, resultStatefulFunctions...)
	}

	return providers, nil
}
