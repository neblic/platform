package rule

import (
	"fmt"

	"github.com/google/cel-go/cel"
	"github.com/neblic/platform/sampler/sample"
)

const sampleKey = "sample"

type Builder struct {
	schema sample.Schema
	env    *cel.Env
}

type SupportedFunctions int

const (
	StreamFunctions SupportedFunctions = iota
	CheckFunctions
)

func NewBuilder(schema sample.Schema, supportedFunctions SupportedFunctions) (*Builder, error) {
	var celEnvOpts []cel.EnvOption

	// Add custom functions to the environemnt
	switch supportedFunctions {
	case StreamFunctions:
		celEnvOpts = append(celEnvOpts, StreamFunctionsEnvOptions...)
	case CheckFunctions:
		celEnvOpts = append(celEnvOpts, CheckFunctionsEnvOptions...)
	}

	switch s := schema.(type) {
	case sample.ProtoSchema:
		typ := string(s.Proto.ProtoReflect().Descriptor().FullName())
		celEnvOpts = append(celEnvOpts,
			cel.Types(s.Proto),
			cel.Variable(sampleKey,
				cel.ObjectType(typ),
			),
		)
	case sample.DynamicSchema:
		celEnvOpts = append(celEnvOpts,
			cel.Variable(sampleKey, cel.MapType(cel.StringType, cel.DynType)),
		)
	default:
		return nil, fmt.Errorf("unknown schema %T", schema)
	}

	// TODO: Investigate limiting CEL environment
	env, err := cel.NewEnv(celEnvOpts...)
	if err != nil {
		return nil, fmt.Errorf("couldn't create a CEL environment: %w", err)
	}

	return &Builder{
		schema: schema,
		env:    env,
	}, nil
}

func (rb *Builder) Build(rule string) (*Rule, error) {
	env := rb.env

	ast, iss := env.Compile(rule)
	if iss != nil && iss.Err() != nil {
		return nil, fmt.Errorf("couldn't compile rule: %w", iss.Err())
	}

	if ast.OutputType() != cel.BoolType {
		return nil, fmt.Errorf("rule expects return type of bool, not %s", ast.OutputType())
	}

	expr, err := cel.AstToCheckedExpr(ast)
	if err != nil {
		return nil, fmt.Errorf("couldn't convert AST to CheckedExpr: %w", err)
	}

	// Inject state to stateful functions
	checkedExprModifier := NewCheckedExprModifier(expr)
	statefulFunctions, err := checkedExprModifier.InjectState()
	if err != nil {
		return nil, err
	}

	ast = cel.CheckedExprToAst(expr)

	// TODO: Investigate interesting program options: e.g. cost estimation/limit
	prg, err := env.Program(ast)
	if err != nil {
		return nil, fmt.Errorf("couldn't build CEL program: %w", err)
	}

	return New(rb.schema, prg, statefulFunctions), nil
}
