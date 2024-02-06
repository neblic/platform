package rule

import (
	"context"
	"fmt"

	"github.com/google/cel-go/cel"
	"github.com/neblic/platform/internal/pkg/data"
	"github.com/neblic/platform/sampler/sample"
)

type sampleCompatibility uint8

const (
	jsonComp sampleCompatibility = 1 << iota
	nativeComp
	protoComp
)

type Rule struct {
	schema            sample.Schema
	prg               cel.Program
	sampleComp        sampleCompatibility
	statefulFunctions []StatefulFunction
}

func New(schema sample.Schema, prg cel.Program, statefulFunctions []StatefulFunction) *Rule {
	r := &Rule{
		schema:            schema,
		prg:               prg,
		statefulFunctions: statefulFunctions,
	}
	r.setCompatibility(schema)

	return r
}

func (r *Rule) setCompatibility(schema sample.Schema) {
	switch schema.(type) {
	case sample.DynamicSchema:
		r.sampleComp = jsonComp | nativeComp | protoComp
	case sample.ProtoSchema:
		r.sampleComp = protoComp
	}
}

func (r *Rule) checkCompatibility(sampleData *data.Data) error {
	switch sampleData.Origin {
	case data.JSONOrigin:
		if !(r.sampleComp&jsonComp != 0) {
			return fmt.Errorf("incompatible sample format")
		}
	case data.NativeOrigin:
		if !(r.sampleComp&nativeComp != 0) {
			return fmt.Errorf("incompatible sample format")
		}
	case data.ProtoOrigin:
		if !(r.sampleComp&protoComp != 0) {
			return fmt.Errorf("incompatible sample format")
		}
	default:
		return fmt.Errorf("unknown sample origin: %s", sampleData.Origin)
	}

	return nil
}

func (r *Rule) Eval(ctx context.Context, sampleData *data.Data) (bool, error) {
	if err := r.checkCompatibility(sampleData); err != nil {
		return false, err
	}

	var (
		smpl any
		err  error
	)
	switch sampleData.Origin {
	case data.ProtoOrigin:
		smpl, err = sampleData.Proto()
		if err != nil {
			return false, fmt.Errorf("failed to get proto message from sample: %w", err)
		}
	default:
		smpl, err = sampleData.Map()
		if err != nil {
			return false, fmt.Errorf("failed to get map from sample: %w", err)
		}
	}

	// Add state variables for the stateful functions
	vars := map[string]interface{}{sampleKey: smpl}
	for _, sf := range r.statefulFunctions {
		vars[sf.StateName()] = sf
	}

	val, _, err := r.prg.ContextEval(ctx, vars)
	if err != nil {
		return false, fmt.Errorf("failed to evaluate sample: %w", err)
	}

	// It is guaranteed to be a boolean because the rule has been checked at build time
	return val.Value().(bool), nil
}
