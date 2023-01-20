package stream

import (
	"fmt"
	"time"

	data "github.com/neblic/platform/controlplane/data"
	"github.com/neblic/platform/controlplane/protos"
	internalclient "github.com/neblic/platform/controlplane/server/internal/defs/client"
	internalsampler "github.com/neblic/platform/controlplane/server/internal/defs/sampler"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type grpcStream[T ToS, F FromS] interface {
	Send(F) error
	Recv() (T, error)
}

type Handler[T ToS, F FromS] interface {
	recvReqToS(T) (bool, F, error)

	startRegistration(T) (string, F, error)
	finishRegistration(T) error
	deregister(string) error

	fromServerMsg(string) F
}

type ClientRecvFromServerReqFunc func(*protos.ClientToServer) (bool, *protos.ServerToClient, error)
type ClientStateChangeFunc func(internalclient.State, data.ClientUID) error

type ClientHandler struct {
	rectToServerReqCb ClientRecvFromServerReqFunc
	stateChangeCb     ClientStateChangeFunc
}

var _ Handler[*protos.ClientToServer, *protos.ServerToClient] = (*ClientHandler)(nil)

func NewClientHandler(
	rectToServerReqCb ClientRecvFromServerReqFunc,
	stateChangeCb ClientStateChangeFunc,
) *ClientHandler {
	return &ClientHandler{
		rectToServerReqCb: rectToServerReqCb,
		stateChangeCb:     stateChangeCb,
	}
}

func (ch ClientHandler) recvReqToS(req *protos.ClientToServer) (bool, *protos.ServerToClient, error) {
	return ch.rectToServerReqCb(req)
}

func (ch ClientHandler) startRegistration(req *protos.ClientToServer) (string, *protos.ServerToClient, error) {
	var st *protos.Status

	_, ok := req.GetMessage().(*protos.ClientToServer_RegisterReq)
	if !ok {
		st = &protos.Status{
			Type:         protos.Status_UNKNOWN,
			ErrorMessage: fmt.Sprintf("received unexpected message type %T while waiting for a register request", req.GetMessage()),
		}
	} else {
		st = &protos.Status{
			Type: protos.Status_OK,
		}
	}

	regRes := ch.fromServerMsg(req.GetClientUid())
	regRes.Message = &protos.ServerToClient_RegisterRes{
		RegisterRes: &protos.ClientRegisterRes{
			Status: st,
		},
	}

	return req.GetClientUid(), regRes, nil
}

func (ch ClientHandler) finishRegistration(req *protos.ClientToServer) error {
	if err := ch.stateChangeCb(internalclient.Registered, data.ClientUID(req.GetClientUid())); err != nil {
		return err
	}

	return nil
}

func (ch ClientHandler) deregister(uid string) error {
	return ch.stateChangeCb(internalclient.Unregistered, data.ClientUID(uid))
}

func (ClientHandler) fromServerMsg(uid string) *protos.ServerToClient {
	return &protos.ServerToClient{
		Timestamp: timestamppb.New(time.Now()),
		ServerUid: uid,
	}
}

type SamplerRecvFromServerReqFunc func(*protos.SamplerToServer) (bool, *protos.ServerToSampler, error)
type SamplerStateChangeFunc func(internalsampler.State, string, string, data.SamplerUID) error

type SamplerHandler struct {
	rectToServerReqCb SamplerRecvFromServerReqFunc
	stateChangeCb     SamplerStateChangeFunc
}

var _ Handler[*protos.SamplerToServer, *protos.ServerToSampler] = (*SamplerHandler)(nil)

func NewSamplerHandler(
	rectToServerReqCb SamplerRecvFromServerReqFunc,
	stateChangeCb SamplerStateChangeFunc,
) *SamplerHandler {
	return &SamplerHandler{
		rectToServerReqCb: rectToServerReqCb,
		stateChangeCb:     stateChangeCb,
	}
}

func (ph SamplerHandler) recvReqToS(req *protos.SamplerToServer) (bool, *protos.ServerToSampler, error) {
	return ph.rectToServerReqCb(req)
}

func (ph SamplerHandler) startRegistration(req *protos.SamplerToServer) (string, *protos.ServerToSampler, error) {
	var st *protos.Status

	_, ok := req.GetMessage().(*protos.SamplerToServer_RegisterReq)
	if !ok {
		st = &protos.Status{
			Type:         protos.Status_UNKNOWN,
			ErrorMessage: fmt.Sprintf("received unexpected message type %T while waiting for a register request", req.GetMessage()),
		}
	} else {
		st = &protos.Status{
			Type: protos.Status_OK,
		}
	}

	regRes := ph.fromServerMsg(req.GetSamplerUid())
	regRes.Message = &protos.ServerToSampler_RegisterRes{
		RegisterRes: &protos.SamplerRegisterRes{
			Status: st,
		},
	}

	return req.GetSamplerUid(), regRes, nil
}

func (ph SamplerHandler) finishRegistration(req *protos.SamplerToServer) error {
	if err := ph.stateChangeCb(internalsampler.Registered, req.GetRegisterReq().GetSamplerName(), req.GetRegisterReq().GetResource(), data.SamplerUID(req.GetSamplerUid())); err != nil {
		return err
	}

	return nil
}

func (ph SamplerHandler) deregister(uid string) error {
	return ph.stateChangeCb(internalsampler.Unregistered, "", "", data.SamplerUID(uid))
}

func (SamplerHandler) fromServerMsg(uid string) *protos.ServerToSampler {
	return &protos.ServerToSampler{
		Timestamp: timestamppb.New(time.Now()),
		ServerUid: uid,
	}
}
