package rule

import (
	"context"
	"encoding/json"
	"fmt"

	structpb "github.com/golang/protobuf/ptypes/struct"
	"github.com/google/cel-go/cel"
	"github.com/neblic/platform/sampler/defs"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

type sampleOrigin uint8

const (
	unknown sampleOrigin = iota
	jsonOrig
	nativeOrig
	protoOrig
)

func (so sampleOrigin) String() string {
	switch so {
	case jsonOrig:
		return "json"
	case nativeOrig:
		return "native"
	case protoOrig:
		return "proto"
	case unknown:
		fallthrough
	default:
		return "unknown"
	}
}

type EvalSample struct {
	origin sampleOrigin

	protoEncoded bool
	proto        proto.Message

	jsonEncoded bool
	json        string

	asMap map[string]any
}

func jsonToStructPb(jsonMsg string) (*structpb.Struct, error) {
	var spb structpb.Struct
	if err := protojson.Unmarshal([]byte(jsonMsg), &spb); err != nil {
		return nil, fmt.Errorf("couldn't unmarshal JSON into structpb.Struct: %w", err)
	}

	return &spb, nil
}

func protoToStructPb(protoMsg proto.Message) (*structpb.Struct, error) {
	var (
		protoJSON []byte
		err       error
	)

	if protoJSON, err = protojson.Marshal(protoMsg); err != nil {
		return nil, fmt.Errorf("couldn't marshall proto message: %w", err)
	}

	return jsonToStructPb(string(protoJSON))
}

func NewEvalSampleFromJSON(jsonSample string) (*EvalSample, error) {
	// TODO: It should be possible to create a new CEL `interpreter.Activation` instead of a structpb.Struct
	// internally using https://github.com/tidwall/gjson or https://github.com/buger/jsonparser
	// that can be passed to Eval which would be faster and won't require mem allocations
	spb, err := jsonToStructPb(jsonSample)
	if err != nil {
		return nil, err
	}

	return &EvalSample{
		origin: jsonOrig,

		protoEncoded: true,
		proto:        spb,

		jsonEncoded: true,
		json:        jsonSample,

		asMap: spb.AsMap(),
	}, nil
}

// NewEvalSampleFromNative build a sample from any Go struct. Only exported fields will be part of the sample
func NewEvalSampleFromNative(nativeSample any) (*EvalSample, error) {
	// TODO: Similarly, we could use reflection to generate a map[string]interface{} directly without
	// converting first to json
	jsonSample, err := json.Marshal(nativeSample)
	if err != nil {
		return nil, fmt.Errorf("couldn't marshal to JSON struct: %w", err)
	}

	// TODO: It should be possible to create a new CEL `interpreter.Activation` instead of a structpb.Struct
	// internally using https://github.com/tidwall/gjson or https://github.com/buger/jsonparser
	// that can be passed to Eval which would be faster and won't require mem allocations
	spb, err := jsonToStructPb(string(jsonSample))
	if err != nil {
		return nil, err
	}

	return &EvalSample{
		origin: nativeOrig,

		protoEncoded: true,
		proto:        spb,

		jsonEncoded: true,
		json:        string(jsonSample),

		asMap: spb.AsMap(),
	}, nil
}

func NewEvalSampleFromProto(protoSample proto.Message) (*EvalSample, error) {
	spb, err := protoToStructPb(protoSample)
	if err != nil {
		return nil, err
	}

	return &EvalSample{
		origin: protoOrig,

		protoEncoded: true,
		proto:        protoSample,

		asMap: spb.AsMap(),
	}, nil
}

func (s *EvalSample) JSON() string {
	if s.jsonEncoded {
		return s.json
	} else if s.protoEncoded {
		jsonStr, _ := protojson.Marshal(s.proto)
		return string(jsonStr)
	}

	return ""
}

func (s *EvalSample) AsMap() map[string]any {
	return s.asMap
}

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

func (r *Rule) checkCompatibility(sample *EvalSample) error {
	switch sample.origin {
	case jsonOrig:
		if !(r.sampleComp&jsonComp != 0) {
			return fmt.Errorf("incompatible sample format")
		}
	case nativeOrig:
		if !(r.sampleComp&nativeComp != 0) {
			return fmt.Errorf("incompatible sample format")
		}
	case protoOrig:
		if !(r.sampleComp&protoComp != 0) {
			return fmt.Errorf("incompatible sample format")
		}
	default:
		return fmt.Errorf("unknown sample origin: %s", sample.origin)
	}

	return nil
}

func (r *Rule) Eval(ctx context.Context, sample *EvalSample) (bool, error) {
	if err := r.checkCompatibility(sample); err != nil {
		return false, err
	}

	if sample.protoEncoded {
		val, _, err := r.prg.ContextEval(ctx, map[string]interface{}{"sample": sample.proto})
		if err != nil {
			return false, fmt.Errorf("failed to evaluate sample: %w", err)
		}

		// It is guaranteed to be a boolean because the rule has been checked at build time
		return val.Value().(bool), nil
	}

	return false, fmt.Errorf("EvalSample can't be evaluated since it is missing required internal data")
}
