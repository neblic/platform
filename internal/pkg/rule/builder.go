package rule

import (
	"fmt"

	"github.com/google/cel-go/cel"
)

type Builder struct {
	schema Schema
	env    *cel.Env
}

const sampleKey = "sample"

func NewBuilder(schema Schema) (*Builder, error) {
	var celEnvOpts []cel.EnvOption

	// Add custom functions to the environemnt
	celEnvOpts = append(celEnvOpts, CelFunctions...)

	switch s := schema.(type) {
	case ProtoSchema:
		typ := string(s.Proto.ProtoReflect().Descriptor().FullName())
		celEnvOpts = append(celEnvOpts,
			cel.Types(s.Proto),
			cel.Variable(sampleKey,
				cel.ObjectType(typ),
			),
		)
	case DynamicSchema:
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
	ast, iss := rb.env.Compile(rule)
	if iss != nil && iss.Err() != nil {
		return nil, fmt.Errorf("couldn't compile rule: %w", iss.Err())
	}

	if ast.OutputType() != cel.BoolType {
		return nil, fmt.Errorf("rule expects return type of bool, not %s", ast.OutputType())
	}

	// TODO: Investigate interesting program options: e.g. cost estimation/limit
	prg, err := rb.env.Program(ast)
	if err != nil {
		return nil, fmt.Errorf("couldn't build CEL program: %w", err)
	}

	return New(rb.schema, prg), nil
}
