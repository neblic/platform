package event

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/google/cel-go/cel"
	"github.com/neblic/platform/internal/pkg/data"
)

const sampleKey = "sample"
const keyKey = "key"

type MetadataBuilder struct {
	noTmpl       bool
	interpolator cel.Program
}

func NewMetadataBuilder(exportTmpl string) (*MetadataBuilder, error) {
	if exportTmpl == "" {
		return &MetadataBuilder{
			noTmpl: true,
		}, nil
	}

	m := &MetadataBuilder{}
	if err := m.buildInterpolator(exportTmpl); err != nil {
		return nil, fmt.Errorf("couldn't build template interpolation environment: %w", err)
	}

	return m, nil
}

func (m *MetadataBuilder) buildInterpolator(exportTmplstring string) error {
	env, err := cel.NewEnv(
		cel.Variable(sampleKey, cel.MapType(cel.StringType, cel.DynType)),
		cel.Variable(keyKey, cel.StringType),
	)
	if err != nil {
		return fmt.Errorf("couldn't create a CEL environment: %w", err)
	}

	ast, iss := env.Compile(exportTmplstring)
	if iss != nil && iss.Err() != nil {
		return fmt.Errorf("couldn't compile export template: %w", iss.Err())
	}

	if !ast.OutputType().IsEquivalentType(cel.MapType(cel.StringType, cel.DynType)) {
		return fmt.Errorf("export template expects return type of map(string, dyn), not %s", ast.OutputType())
	}

	prg, err := env.Program(ast)
	if err != nil {
		return fmt.Errorf("couldn't build CEL program: %w", err)
	}

	m.interpolator = prg

	return nil
}

func (m *MetadataBuilder) Build(ctx context.Context, sampleData *data.Data, sampleDataKey string) (string, error) {
	if m.noTmpl {
		return sampleData.JSON()
	}

	smpl, err := sampleData.Map()
	if err != nil {
		return "", fmt.Errorf("failed to get map from sample: %w", err)
	}

	val, _, err := m.interpolator.ContextEval(ctx, map[string]interface{}{sampleKey: smpl, keyKey: sampleDataKey})
	if err != nil {
		return "", fmt.Errorf("failed to evaluate sample: %w", err)
	}

	// It is guaranteed to be a string because the rule has been checked at build time
	ret, err := val.ConvertToNative(reflect.TypeOf(map[string]interface{}{}))
	if err != nil {
		return "", fmt.Errorf("failed to convert result to native map: %w", err)
	}

	jsonRet, err := json.Marshal(ret)
	if err != nil {
		return "", fmt.Errorf("failed to marshal result to JSON: %w", err)
	}

	return string(jsonRet), nil
}
