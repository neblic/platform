package grpc

import (
	"context"

	oldProto "github.com/golang/protobuf/proto"
	"github.com/neblic/platform/sampler"
	"github.com/neblic/platform/sampler/sample"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
)

func toProtoMessage(msg interface{}) proto.Message {
	var protoMsg proto.Message
	if version1Msg, ok := msg.(oldProto.Message); ok {
		protoMsg = oldProto.MessageV2(version1Msg)
	} else {
		protoMsg, _ = msg.(proto.Message)
	}

	return protoMsg
}

func getProtoSampler(samplers map[string]sampler.Sampler, key string, schema proto.Message) sampler.Sampler {
	var (
		smplr sampler.Sampler
		ok    bool
	)
	if smplr, ok = samplers[key]; !ok {
		smplr, _ = sampler.New(
			key,
			sample.NewProtoSchema(schema),
		)
		samplers[key] = smplr
	}

	return smplr
}

// UnaryClientInterceptor provides a gRPC unary client interceptor that lazily creates two samplers
// per each gRPC method called by the client. The samplers capture the request and response gRPC messages.
func UnaryClientInterceptor() grpc.UnaryClientInterceptor {
	samplerReqs := make(map[string]sampler.Sampler)
	samplerResps := make(map[string]sampler.Sampler)

	return func(
		ctx context.Context,
		method string,
		req, reply interface{},
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		callOpts ...grpc.CallOption,
	) error {
		reqProtoMsg := toProtoMessage(req)
		if reqProtoMsg != nil {
			samplerName := method + "Req"
			samplerReq := getProtoSampler(samplerReqs, samplerName, reqProtoMsg)
			// TODO: allow the user to provide a way to get the determinant from the request
			samplerReq.Sample(ctx, sample.ProtoSample(reqProtoMsg, ""))
		}

		err := invoker(ctx, method, req, reply, cc, callOpts...)

		resProtoMsg := toProtoMessage(reply)
		if resProtoMsg != nil {
			samplerName := method + "Res"
			samplerRes := getProtoSampler(samplerResps, samplerName, resProtoMsg)
			// TODO: allow the user to provide a way to get the determinant from the response
			samplerRes.Sample(ctx, sample.ProtoSample(resProtoMsg, ""))
		}

		return err
	}
}

// UnaryServerInterceptor provides a gRPC unary server interceptor that lazily creates two samplers
// per each gRPC method served The samplers capture the request and response gRPC messages.
func UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	samplerReqs := make(map[string]sampler.Sampler)
	samplerResps := make(map[string]sampler.Sampler)

	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		reqProtoMsg := toProtoMessage(req)
		if reqProtoMsg != nil {
			samplerName := info.FullMethod + "Req"
			samplerReq := getProtoSampler(samplerReqs, samplerName, reqProtoMsg)
			// TODO: allow the user to provide a way to get the determinant from the request
			samplerReq.Sample(ctx, sample.ProtoSample(reqProtoMsg, ""))
		}

		reply, err := handler(ctx, req)

		resProtoMsg := toProtoMessage(reply)
		if resProtoMsg != nil {
			samplerName := info.FullMethod + "Res"
			samplerRes := getProtoSampler(samplerResps, samplerName, resProtoMsg)
			// TODO: allow the user to provide a way to get the determinant from the response
			samplerRes.Sample(ctx, sample.ProtoSample(resProtoMsg, ""))
		}

		return reply, err
	}
}
