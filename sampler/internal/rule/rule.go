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

	jsonEncoded bool
	json        string

	nativeEncoded bool
	native        any

	protoEncoded bool
	proto        proto.Message

	asMap map[string]any
}

func jsonToStructPb(jsonMsg string) (*structpb.Struct, error) {
	var spb structpb.Struct
	if err := protojson.Unmarshal([]byte(jsonMsg), &spb); err != nil {
		return nil, fmt.Errorf("couldn't unmarshal JSON into structpb.Struct: %w", err)
	}

	return &spb, nil
}

func NewEvalSampleFromJSON(jsonSample string) (*EvalSample, error) {
	return &EvalSample{
		origin: jsonOrig,

		jsonEncoded: true,
		json:        jsonSample,
	}, nil
}

// NewEvalSampleFromNative build a sample from any Go struct. Only exported fields will be part of the sample
func NewEvalSampleFromNative(nativeSample any) (*EvalSample, error) {
	return &EvalSample{
		origin: nativeOrig,

		nativeEncoded: true,
		native:        nativeSample,
	}, nil
}

func NewEvalSampleFromProto(protoSample proto.Message) (*EvalSample, error) {
	return &EvalSample{
		origin: protoOrig,

		protoEncoded: true,
		proto:        protoSample,
	}, nil
}

func (s *EvalSample) JSON() (string, error) {
	if s.jsonEncoded {
		return s.json, nil
	} else if s.origin == protoOrig {
		// if the original sample was a proto message, we use it to create a JSON representation since
		// it will not alter the original message structure
		jsonStr, err := protojson.Marshal(s.proto)
		if err != nil {
			return "", fmt.Errorf("couldn't marshal to JSON struct: %w", err)
		}

		s.jsonEncoded = true
		s.json = string(jsonStr)

		return s.json, nil
	} else if s.origin == nativeOrig {
		// if the original sample was a native Go struct, only the exported fields will be part of the
		// JSON representation
		jsonSample, err := json.Marshal(s.native)
		if err != nil {
			return "", fmt.Errorf("couldn't marshal to JSON struct: %w", err)
		}

		s.jsonEncoded = true
		s.json = string(jsonSample)

		return s.json, nil
	}

	return "", fmt.Errorf("couldn't get a JSON encoded message")
}

func (s *EvalSample) Proto() (proto.Message, error) {
	if !s.protoEncoded {
		// TODO: If it is a native sample, it should be possibe to create a map[string]interface{} directly using reflections
		// without converting first to json so we can feed it to the CEL evaluator directly
		// For now, just convert to JSON and then to a generic structpb proto message
		jsonSample, err := s.JSON()
		if err != nil {
			return nil, err
		}

		// TODO: It should be possible to create a new CEL `interpreter.Activation` instead of a structpb.Struct
		// internally using https://github.com/tidwall/gjson or https://github.com/buger/jsonparser
		// that can be passed to Eval which would be faster and won't require mem allocations
		spb, err := jsonToStructPb(jsonSample)
		if err != nil {
			return nil, err
		}

		s.protoEncoded = true
		s.proto = spb
	}

	return s.proto, nil
}

func (s *EvalSample) Map() (map[string]any, error) {
	if s.asMap == nil {
		structPb, protoIsStructPb := s.proto.(*structpb.Struct)

		if s.protoEncoded && protoIsStructPb {
			s.asMap = structPb.AsMap()
		} else {
			jsonSample, err := s.JSON()
			if err != nil {
				return nil, err
			}

			asMap := make(map[string]interface{})
			if err := json.Unmarshal([]byte(jsonSample), &asMap); err != nil {
				return nil, fmt.Errorf("couldn't unmarshal JSON into a map: %w", err)
			}

			s.asMap = asMap
		}
	}

	return s.asMap, nil
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

	protoSample, err := sample.Proto()
	if err != nil {
		return false, fmt.Errorf("failed to get proto message from sample: %w", err)
	}

	val, _, err := r.prg.ContextEval(ctx, map[string]interface{}{"sample": protoSample})
	if err != nil {
		return false, fmt.Errorf("failed to evaluate sample: %w", err)
	}

	// It is guaranteed to be a boolean because the rule has been checked at build time
	return val.Value().(bool), nil
}
