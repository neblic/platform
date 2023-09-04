package rule

import (
	"math"
	"time"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
)

// CelFunctions contains all the custom functions used in the rules
var CelFunctions = []cel.EnvOption{
	// aba
	absDoubleFunc,
	absIntFunc,
	// now
	nowFunc,
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
