package rule

import (
	"fmt"

	"github.com/google/cel-go/cel"
	"github.com/neblic/platform/sampler/defs"
)

type Builder struct {
	schema defs.Schema
	env    *cel.Env
}

const sampleKey = "sample"

func NewBuilder(ruleSchema defs.Schema) (*Builder, error) {
	var celEnvOpts []cel.EnvOption
	switch s := ruleSchema.(type) {
	case defs.ProtoSchema:
		typ := string(s.Proto.ProtoReflect().Descriptor().FullName())
		celEnvOpts = append(celEnvOpts,
			cel.Types(s.Proto),
			cel.Variable(sampleKey,
				cel.ObjectType(typ),
			),
		)
	case defs.DynamicSchema:
		celEnvOpts = append(celEnvOpts,
			cel.Variable(sampleKey, cel.MapType(cel.StringType, cel.DynType)),
		)
	default:
		return nil, fmt.Errorf("unknown schema %T", ruleSchema)
	}

	// TODO: Investigate limiting CEL environment
	env, err := cel.NewEnv(celEnvOpts...)
	if err != nil {
		return nil, fmt.Errorf("couldn't create a CEL environment: %w", err)
	}

	return &Builder{
		schema: ruleSchema,
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
