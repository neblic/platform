package data

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/mitchellh/mapstructure"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/structpb"
)

type Origin uint8

const (
	Unknown Origin = iota
	JSONOrigin
	NativeOrigin
	ProtoOrigin
)

func (so Origin) String() string {
	switch so {
	case JSONOrigin:
		return "json"
	case NativeOrigin:
		return "native"
	case ProtoOrigin:
		return "proto"
	case Unknown:
		fallthrough
	default:
		return "unknown"
	}
}

type Data struct {
	Origin Origin

	jsonEncoded bool
	json        string

	nativeEncoded bool
	native        any

	protoEncoded bool
	proto        proto.Message

	asMap map[string]any
}

// NewSampleDataFromMap build a sample from a JSON object
func NewSampleDataFromJSON(jsonSample string) *Data {
	return &Data{
		Origin: JSONOrigin,

		jsonEncoded: true,
		json:        jsonSample,
	}
}

// NewSampleDataFromNative build a sample from any Go struct. Only exported fields will be part of the sample
func NewSampleDataFromNative(nativeSample any) *Data {
	return &Data{
		Origin: NativeOrigin,

		nativeEncoded: true,
		native:        nativeSample,
	}
}

func NewSampleDataFromProto(protoSample proto.Message) *Data {
	return &Data{
		Origin: ProtoOrigin,

		protoEncoded: true,
		proto:        protoSample,
	}
}

func (s *Data) JSON() (string, error) {
	if s.jsonEncoded {
		return s.json, nil
	}

	switch s.Origin {
	case ProtoOrigin:
		jsonStr, err := protojson.Marshal(s.proto)
		if err != nil {
			return "", fmt.Errorf("couldn't marshal to JSON struct: %w", err)
		}

		s.jsonEncoded = true
		s.json = string(jsonStr)

		return s.json, nil
	case NativeOrigin:
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

// Proto returns the proto.Message representation of the sample
// NOTE: converting a native sample to proto will produce a structp.Struct proto message
// that will lose type information on number fields
func (s *Data) Proto() (proto.Message, error) {
	if s.protoEncoded {
		return s.proto, nil
	}

	// it is actually not needed anymore to be able to convert any type of sample to proto
	// since we use their map representation to evaluate CEL expressions insteaf of structpb.Struct
	// as we did in previous versions
	switch s.Origin {
	case NativeOrigin:
		var nativeMap map[string]any
		err := mapstructure.Decode(s.native, &nativeMap)
		if err != nil {
			return nil, fmt.Errorf("couldn't decode native sample into a map: %w", err)
		}
		s.asMap = nativeMap

		// we lose the number type information here
		spb, err := structpb.NewStruct(nativeMap)
		if err != nil {
			return nil, fmt.Errorf("couldn't create a structpb.Struct from a map: %w", err)
		}

		s.protoEncoded = true
		s.proto = spb

		return s.proto, nil
	case JSONOrigin:
		// TODO: It should be possible to create a new CEL `interpreter.Activation` instead of a structpb.Struct
		// internally using https://github.com/tidwall/gjson or https://github.com/buger/jsonparser
		// that can be passed to Eval which would be faster and that may not require mem allocations
		var spb structpb.Struct
		if err := protojson.Unmarshal([]byte(s.json), &spb); err != nil {
			return nil, fmt.Errorf("couldn't unmarshal JSON into structpb.Struct: %w", err)
		}

		s.protoEncoded = true
		s.proto = &spb

		return s.proto, nil
	}

	return nil, fmt.Errorf("couldn't get a proto encoded message")
}

func protoValueToNative(fd protoreflect.FieldDescriptor, val protoreflect.Value) (any, error) {
	switch fd.Kind() {
	case protoreflect.BoolKind:
		if val.Bool() {
			return true, nil
		}
		return false, nil
	case protoreflect.Sint32Kind:
		fallthrough
	case protoreflect.Sfixed32Kind:
		fallthrough
	case protoreflect.Int32Kind:
		return int32(val.Int()), nil
	case protoreflect.Fixed32Kind:
		fallthrough
	case protoreflect.Uint32Kind:
		return uint32(val.Uint()), nil
	case protoreflect.Sint64Kind:
		fallthrough
	case protoreflect.Sfixed64Kind:
		fallthrough
	case protoreflect.Int64Kind:
		return int64(val.Int()), nil
	case protoreflect.Fixed64Kind:
		fallthrough
	case protoreflect.Uint64Kind:
		return uint64(val.Uint()), nil
	case protoreflect.FloatKind:
		return float32(val.Float()), nil
	case protoreflect.DoubleKind:
		return float64(val.Float()), nil
	case protoreflect.BytesKind:
		return val.Bytes(), nil
	case protoreflect.StringKind:
		return val.String(), nil
	case protoreflect.MessageKind:
		return protoObjectToMap(val.Message())
	case protoreflect.GroupKind:
		fallthrough
	case protoreflect.EnumKind:
		return int32(val.Enum()), nil
	}

	return nil, fmt.Errorf("unsupported proto field kind: %s", fd.Kind())
}

func protoObjectToMap(pb protoreflect.Message) (map[string]any, error) {
	var (
		obj    = make(map[string]any)
		aggErr error
	)

	pb.Range(func(fd protoreflect.FieldDescriptor, value protoreflect.Value) bool {
		if fd.IsMap() {
			m := make(map[any]any)

			value.Map().Range(func(key protoreflect.MapKey, value protoreflect.Value) bool {
				keyNative, err := protoValueToNative(fd.MapKey(), key.Value())
				if err != nil {
					aggErr = errors.Join(aggErr, fmt.Errorf("couldn't convert proto map keyue to native type: %w", err))
					return false
				}

				valNative, err := protoValueToNative(fd.MapValue(), value)
				if err != nil {
					aggErr = errors.Join(aggErr, fmt.Errorf("couldn't convert proto map value to native type: %w", err))
					return false
				}

				m[keyNative] = valNative
				return true
			})
			obj[fd.TextName()] = m
		} else if fd.IsList() {
			var l []any
			for i := 0; i < value.List().Len(); i++ {
				valNative, err := protoValueToNative(fd, value.List().Get(i))
				if err != nil {
					aggErr = errors.Join(aggErr, fmt.Errorf("couldn't convert proto list element to native type: %w", err))
					return false
				}
				l = append(l, valNative)
			}
			obj[fd.TextName()] = l
		} else {
			valNative, err := protoValueToNative(fd, value)
			if err != nil {
				aggErr = errors.Join(aggErr, fmt.Errorf("couldn't convert proto field to native type: %w", err))
				return false
			}
			obj[fd.TextName()] = valNative
		}

		return true
	})

	return obj, aggErr
}

func (s *Data) Map() (map[string]any, error) {
	if s.asMap != nil {
		return s.asMap, nil
	}

	switch s.Origin {
	case NativeOrigin:
		var asMap map[string]any
		err := mapstructure.Decode(s.native, &asMap)
		if err != nil {
			return nil, fmt.Errorf("couldn't decode native sample into a map: %w", err)
		}
		s.asMap = asMap
	case ProtoOrigin:
		asMap, err := protoObjectToMap(s.proto.ProtoReflect())
		if err != nil {
			return nil, fmt.Errorf("couldn't convert proto sample into a map: %w", err)
		}
		s.asMap = asMap
	case JSONOrigin:
		asMap := make(map[string]interface{})
		if err := json.Unmarshal([]byte(s.json), &asMap); err != nil {
			return nil, fmt.Errorf("couldn't unmarshal JSON into a map: %w", err)
		}
		s.asMap = asMap
	}

	return s.asMap, nil
}
