package rule

import (
	"fmt"

	"github.com/google/cel-go/cel"
	expr "google.golang.org/genproto/googleapis/api/expr/v1alpha1"
)

type StatefulFunction interface {
	GetName() string
	Enabled() bool
	ParseCallExpression(callExpression *expr.Expr_Call) error
	GetCelEnvs(stateProvider *StateProvider) []cel.EnvOption
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

func ParseStatefulFunctions(statefulFunctions []StatefulFunction, expression *expr.Expr) error {
	if expression == nil {
		return nil
	}

	switch expression.ExprKind.(type) {
	case *expr.Expr_SelectExpr:
		selectExpr := expression.GetSelectExpr()

		// Recursive check of the operand
		err := ParseStatefulFunctions(statefulFunctions, selectExpr.Operand)
		if err != nil {
			return err
		}

	case *expr.Expr_CallExpr:
		callExpr := expression.GetCallExpr()

		// Check if function name matches any of the stateful functions
		for _, statefulFunction := range statefulFunctions {
			if statefulFunction.GetName() == callExpr.Function {
				err := statefulFunction.ParseCallExpression(callExpr)
				if err != nil {
					return err
				}
				break
			}
		}

		// Recursive check of the target
		err := ParseStatefulFunctions(statefulFunctions, callExpr.Target)
		if err != nil {
			return err
		}
		// Recursive check of the arguments
		for _, arg := range callExpr.Args {
			err := ParseStatefulFunctions(statefulFunctions, arg)
			if err != nil {
				return err
			}
		}

	case *expr.Expr_ListExpr:
		listExpr := expression.GetListExpr()

		// Recursive check of the elements
		for _, element := range listExpr.Elements {
			err := ParseStatefulFunctions(statefulFunctions, element)
			if err != nil {
				return err
			}
		}

	case *expr.Expr_StructExpr:
		structExpr := expression.GetStructExpr()

		// Recursive check of the entries
		for _, entry := range structExpr.Entries {
			err := ParseStatefulFunctions(statefulFunctions, entry.Value)
			if err != nil {
				return err
			}
		}

	case *expr.Expr_ComprehensionExpr:
		comprehensionExpr := expression.GetComprehensionExpr()

		// Recursive check of the itereration range
		err := ParseStatefulFunctions(statefulFunctions, comprehensionExpr.IterRange)
		if err != nil {
			return err
		}
		// Recursive check of the accumulator initialization
		err = ParseStatefulFunctions(statefulFunctions, comprehensionExpr.AccuInit)
		if err != nil {
			return err
		}
		// Recursive check of the loop condition
		err = ParseStatefulFunctions(statefulFunctions, comprehensionExpr.LoopCondition)
		if err != nil {
			return err
		}
		// Recursive check of the loop step
		err = ParseStatefulFunctions(statefulFunctions, comprehensionExpr.LoopStep)
		if err != nil {
			return err
		}
		// Recursive check of the result
		err = ParseStatefulFunctions(statefulFunctions, comprehensionExpr.Result)
		if err != nil {
			return err
		}
	}

	return nil
}
