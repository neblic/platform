package rule

import (
	"context"
	"fmt"

	"github.com/google/cel-go/cel"
	"github.com/neblic/platform/sampler/defs"
	"github.com/neblic/platform/sampler/internal/sample"
)

type sampleCompatibility uint8

const (
	jsonComp sampleCompatibility = 1 << iota
	nativeComp
	protoComp
)

type Rule struct {
	schema     defs.Schema
	prg        cel.Program
	sampleComp sampleCompatibility
}

func New(schema defs.Schema, prg cel.Program) *Rule {
	r := &Rule{
		schema: schema,
		prg:    prg,
	}
	r.setCompatibility(schema)

	return r
}

func (r *Rule) setCompatibility(schema defs.Schema) {
	switch schema.(type) {
	case defs.DynamicSchema:
		r.sampleComp = jsonComp | nativeComp | protoComp
	case defs.ProtoSchema:
		r.sampleComp = protoComp
	}
}

func (r *Rule) checkCompatibility(sampleData *sample.Data) error {
	switch sampleData.Origin {
	case sample.JSONOrig:
		if !(r.sampleComp&jsonComp != 0) {
			return fmt.Errorf("incompatible sample format")
		}
	case sample.NativeOrig:
		if !(r.sampleComp&nativeComp != 0) {
			return fmt.Errorf("incompatible sample format")
		}
	case sample.ProtoOrig:
		if !(r.sampleComp&protoComp != 0) {
			return fmt.Errorf("incompatible sample format")
		}
	default:
		return fmt.Errorf("unknown sample origin: %s", sampleData.Origin)
	}

	return nil
}

func (r *Rule) Eval(ctx context.Context, sampleData *sample.Data) (bool, error) {
	if err := r.checkCompatibility(sampleData); err != nil {
		return false, err
	}

	var (
		smpl any
		err  error
	)
	switch sampleData.Origin {
	case sample.ProtoOrig:
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

	val, _, err := r.prg.ContextEval(ctx, map[string]interface{}{sampleKey: smpl})
	if err != nil {
		return false, fmt.Errorf("failed to evaluate sample: %w", err)
	}

	// It is guaranteed to be a boolean because the rule has been checked at build time
	return val.Value().(bool), nil
}
