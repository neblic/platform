package rule

import (
	"fmt"

	expr "google.golang.org/genproto/googleapis/api/expr/v1alpha1"
)

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

func getOrderFromExpression(expression *expr.Expr) (OrderType, error) {
	constExpr, ok := expression.ExprKind.(*expr.Expr_ConstExpr)
	if !ok {
		return OrderTypeUnknown, fmt.Errorf("sequence function second argument requires a constant string with value 'asc' or 'desc'")
	}

	stringExpr, ok := constExpr.ConstExpr.ConstantKind.(*expr.Constant_StringValue)
	if !ok {
		return OrderTypeUnknown, fmt.Errorf("sequence function second argument requires a constant string with value 'asc' or 'desc'")
	}

	var order OrderType
	switch stringExpr.StringValue {
	case "asc":
		order = OrderTypeAsc
	case "desc":
		order = OrderTypeDesc
	default:
		return OrderTypeUnknown, fmt.Errorf("sequence function second argument requires a constant string with value 'asc' or 'desc'")
	}

	return order, nil
}

type CheckedExprModifier struct {
	currentStateId int64
	currentExprId  int64
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
		currentStateId: 0,
		currentExprId:  id,
		CheckedExpr:    checkedExpr,
	}
}

func (cem *CheckedExprModifier) injectStateToCall(exprID int64, callExpr *expr.Expr_Call, messageType string) string {
	stateName := "state" + fmt.Sprintf("%d", cem.currentStateId)

	argumentExpression := &expr.Expr{
		Id: cem.currentExprId,
		ExprKind: &expr.Expr_IdentExpr{
			IdentExpr: &expr.Expr_Ident{
				Name: stateName,
			},
		},
	}

	// Add state argument
	callExpr.Args = append(callExpr.Args, argumentExpression)

	// Add expression id to type map
	cem.CheckedExpr.TypeMap[cem.currentExprId] = &expr.Type{
		TypeKind: &expr.Type_MessageType{
			MessageType: messageType,
		},
	}

	// Modify overloads in the reference map
	for i, overload := range cem.CheckedExpr.ReferenceMap[exprID].OverloadId {
		cem.CheckedExpr.ReferenceMap[exprID].OverloadId[i] = overload + "_state"
	}

	cem.currentStateId++
	cem.currentExprId++

	return stateName
}

func (cem *CheckedExprModifier) InjectState() ([]StatefulFunction, error) {
	return cem.injectState(cem.CheckedExpr.Expr)
}

func (cem *CheckedExprModifier) injectState(expression *expr.Expr) ([]StatefulFunction, error) {
	if expression == nil {
		return []StatefulFunction{}, nil
	}

	statefulFunctions := []StatefulFunction{}
	switch expression.ExprKind.(type) {
	case *expr.Expr_SelectExpr:
		selectExpr := expression.GetSelectExpr()

		// Recursive check of the operand
		selectStatefulFunctions, err := cem.injectState(selectExpr.Operand)
		if err != nil {
			return []StatefulFunction{}, err
		}
		statefulFunctions = append(statefulFunctions, selectStatefulFunctions...)

	case *expr.Expr_CallExpr:
		callExpr := expression.GetCallExpr()

		switch callExpr.Function {
		case "sequence":

			order, err := getOrderFromExpression(callExpr.Args[1])
			if err != nil {
				return []StatefulFunction{}, err
			}

			stateName := cem.injectStateToCall(expression.Id, callExpr, sequenceStatefulFunctionType.String())

			// Add stateful function
			statefulFunctions = append(statefulFunctions, &SequenceStatefulFunction{stateName: stateName, order: order})

		case "complete":
			step, err := getStepFromExpression(callExpr.Args[1])
			if err != nil {
				return []StatefulFunction{}, err
			}

			stateName := cem.injectStateToCall(expression.Id, callExpr, completeStatefulFunctionType.String())

			// Add stateful function
			statefulFunctions = append(statefulFunctions, &CompleteStatefulFunction{stateName: stateName, step: step})
		}

		// Recursive check of the target
		targetStatefulFunctions, err := cem.injectState(callExpr.Target)
		if err != nil {
			return []StatefulFunction{}, err
		}
		statefulFunctions = append(statefulFunctions, targetStatefulFunctions...)
		// Recursive check of the arguments
		for _, arg := range callExpr.Args {
			argsStatefulFunctions, err := cem.injectState(arg)
			if err != nil {
				return []StatefulFunction{}, err
			}
			statefulFunctions = append(statefulFunctions, argsStatefulFunctions...)
		}

	case *expr.Expr_ListExpr:
		listExpr := expression.GetListExpr()

		// Recursive check of the elements
		for _, element := range listExpr.Elements {
			elementSatatefulFunctions, err := cem.injectState(element)
			if err != nil {
				return []StatefulFunction{}, err
			}
			statefulFunctions = append(statefulFunctions, elementSatatefulFunctions...)
		}

	case *expr.Expr_StructExpr:
		structExpr := expression.GetStructExpr()

		// Recursive check of the entries
		for _, entry := range structExpr.Entries {
			structStatefulFunction, err := cem.injectState(entry.Value)
			if err != nil {
				return []StatefulFunction{}, err
			}
			statefulFunctions = append(statefulFunctions, structStatefulFunction...)
		}

	case *expr.Expr_ComprehensionExpr:
		comprehensionExpr := expression.GetComprehensionExpr()

		// Recursive check of the itereration range
		iterRangeStatefulFunction, err := cem.injectState(comprehensionExpr.IterRange)
		if err != nil {
			return []StatefulFunction{}, err
		}
		statefulFunctions = append(statefulFunctions, iterRangeStatefulFunction...)
		// Recursive check of the accumulator initialization
		accuIntStatefulFunctions, err := cem.injectState(comprehensionExpr.AccuInit)
		if err != nil {
			return []StatefulFunction{}, err
		}
		statefulFunctions = append(statefulFunctions, accuIntStatefulFunctions...)
		// Recursive check of the loop condition
		loopConditionStatefulFunctions, err := cem.injectState(comprehensionExpr.LoopCondition)
		if err != nil {
			return []StatefulFunction{}, err
		}
		statefulFunctions = append(statefulFunctions, loopConditionStatefulFunctions...)
		// Recursive check of the loop step
		loopStepStatefulFunctions, err := cem.injectState(comprehensionExpr.LoopStep)
		if err != nil {
			return []StatefulFunction{}, err
		}
		statefulFunctions = append(statefulFunctions, loopStepStatefulFunctions...)
		// Recursive check of the result
		resultStatefulFunctions, err := cem.injectState(comprehensionExpr.Result)
		if err != nil {
			return []StatefulFunction{}, err
		}
		statefulFunctions = append(statefulFunctions, resultStatefulFunctions...)
	}

	return statefulFunctions, nil
}
