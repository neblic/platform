package stream

import (
	"context"
	"fmt"
	"time"

	"github.com/neblic/platform/controlplane/protos"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type grpcStream[F FromS, T ToS] interface {
	Send(T) error
	Recv() (F, error)
}

// Handler implements all the logic to handle a stream implementation (to a sampler or a client)
type Handler[F FromS, T ToS] interface {
	grpcStream(context.Context, protos.ControlPlaneClient) (grpcStream[F, T], error)
	recvServerReq(F) (bool, T, error)

	regReqMsg(uid string) T
	validateRegResMsg(msg F) (string, error)
	toServerMsg(uid string) T
}

var _ Handler[*protos.ServerToClient, *protos.ClientToServer] = (*ClientHandler)(nil)

type ClientHandler struct {
	// TODO: Pack return into a struct
	recvServerReqCb func(*protos.ServerToClient) (bool, *protos.ClientToServer, error)
}

func NewClientHandler(recvServerReqCb func(*protos.ServerToClient) (bool, *protos.ClientToServer, error)) Handler[*protos.ServerToClient, *protos.ClientToServer] {
	return &ClientHandler{
		recvServerReqCb: recvServerReqCb,
	}
}

func (ClientHandler) grpcStream(ctx context.Context, client protos.ControlPlaneClient) (grpcStream[*protos.ServerToClient, *protos.ClientToServer], error) {
	return client.ClientConn(ctx)
}

func (ch ClientHandler) recvServerReq(req *protos.ServerToClient) (bool, *protos.ClientToServer, error) {
	return ch.recvServerReqCb(req)
}

func (ClientHandler) toServerMsg(uid string) *protos.ClientToServer {
	return &protos.ClientToServer{
		ClientUid: uid,
		Timestamp: timestamppb.New(time.Now()),
	}
}

func (ch ClientHandler) regReqMsg(uid string) *protos.ClientToServer {
	toServerMsg := ch.toServerMsg(uid)
	toServerMsg.Message = &protos.ClientToServer_RegisterReq{}

	return toServerMsg
}

func (ClientHandler) validateRegResMsg(msg *protos.ServerToClient) (string, error) {
	regResMsg, ok := msg.GetMessage().(*protos.ServerToClient_RegisterRes)
	if !ok {
		return "", fmt.Errorf("received unexpected message type %T while waiting for a register request response", msg.GetMessage())
	}

	regRes := regResMsg.RegisterRes
	if regRes.GetStatus().GetType() != protos.Status_OK {
		return "", fmt.Errorf("server could not register client, reason: %s", regRes.GetStatus().GetErrorMessage())
	}

	return msg.GetServerUid(), nil
}

var _ Handler[*protos.ServerToSampler, *protos.SamplerToServer] = (*SamplerHandler)(nil)

type SamplerHandler struct {
	name            string
	resource        string
	recvServerReqCb func(*protos.ServerToSampler) (bool, *protos.SamplerToServer, error)
	initialConfig   *protos.ClientSamplerConfigUpdate
}

func NewSamplerHandler(name, resource string, recvServerReqCb func(*protos.ServerToSampler) (bool, *protos.SamplerToServer, error), initialConfig *protos.ClientSamplerConfigUpdate) Handler[*protos.ServerToSampler, *protos.SamplerToServer] {
	return &SamplerHandler{
		name:            name,
		resource:        resource,
		recvServerReqCb: recvServerReqCb,
		initialConfig:   initialConfig,
	}
}

func (ch SamplerHandler) grpcStream(ctx context.Context, client protos.ControlPlaneClient) (grpcStream[*protos.ServerToSampler, *protos.SamplerToServer], error) {
	return client.SamplerConn(ctx)
}

func (ch SamplerHandler) recvServerReq(req *protos.ServerToSampler) (bool, *protos.SamplerToServer, error) {
	return ch.recvServerReqCb(req)
}

func (ch SamplerHandler) toServerMsg(uid string) *protos.SamplerToServer {
	return &protos.SamplerToServer{
		Name:       ch.name,
		Resouce:    ch.resource,
		SamplerUid: uid,
		Timestamp:  timestamppb.New(time.Now()),
	}
}

func (ch SamplerHandler) regReqMsg(uid string) *protos.SamplerToServer {
	toServerMsg := ch.toServerMsg(uid)
	toServerMsg.Message = &protos.SamplerToServer_RegisterReq{
		RegisterReq: &protos.SamplerRegisterReq{
			InitialConfig: ch.initialConfig,
		},
	}

	return toServerMsg
}

func (SamplerHandler) validateRegResMsg(msg *protos.ServerToSampler) (string, error) {
	regResMsg, ok := msg.GetMessage().(*protos.ServerToSampler_RegisterRes)
	if !ok {
		return "", fmt.Errorf("received unexpected message type %T while waiting for a register request response", msg.GetMessage())
	}

	regRes := regResMsg.RegisterRes
	if regRes.GetStatus().GetType() != protos.Status_OK {
		return "", fmt.Errorf("server could not register client, reason: %s", regRes.GetStatus().GetErrorMessage())
	}

	return msg.GetServerUid(), nil
}
