package stream

import (
	"fmt"
	"time"

	"github.com/neblic/platform/controlplane/control"
	"github.com/neblic/platform/controlplane/protos"
	"github.com/neblic/platform/controlplane/server/internal/defs"
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
type ClientStatusChangeFunc func(defs.Status, control.ClientUID) error

type ClientHandler struct {
	rectToServerReqCb ClientRecvFromServerReqFunc
	statusChangeCb    ClientStatusChangeFunc
}

var _ Handler[*protos.ClientToServer, *protos.ServerToClient] = (*ClientHandler)(nil)

func NewClientHandler(
	rectToServerReqCb ClientRecvFromServerReqFunc,
	statusChangeCb ClientStatusChangeFunc,
) *ClientHandler {
	return &ClientHandler{
		rectToServerReqCb: rectToServerReqCb,
		statusChangeCb:    statusChangeCb,
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
	if err := ch.statusChangeCb(defs.RegisteredStatus, control.ClientUID(req.GetClientUid())); err != nil {
		return err
	}

	return nil
}

func (ch ClientHandler) deregister(uid string) error {
	return ch.statusChangeCb(defs.UnregisteredStatus, control.ClientUID(uid))
}

func (ClientHandler) fromServerMsg(uid string) *protos.ServerToClient {
	return &protos.ServerToClient{
		Timestamp: timestamppb.New(time.Now()),
		ServerUid: uid,
	}
}

type SamplerRecvFromServerReqFunc func(*protos.SamplerToServer) (bool, *protos.ServerToSampler, error)
type SamplerStatusChangeFunc func(defs.Status, control.SamplerUID, *protos.SamplerToServer) error

type SamplerHandler struct {
	rectToServerReqCb SamplerRecvFromServerReqFunc
	statusChangeCb    SamplerStatusChangeFunc
}

var _ Handler[*protos.SamplerToServer, *protos.ServerToSampler] = (*SamplerHandler)(nil)

func NewSamplerHandler(
	rectToServerReqCb SamplerRecvFromServerReqFunc,
	statusChangeCb SamplerStatusChangeFunc,
) *SamplerHandler {
	return &SamplerHandler{
		rectToServerReqCb: rectToServerReqCb,
		statusChangeCb:    statusChangeCb,
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
	if err := ph.statusChangeCb(defs.RegisteredStatus, control.SamplerUID(req.GetSamplerUid()), req); err != nil {
		return err
	}

	return nil
}

func (ph SamplerHandler) deregister(uid string) error {
	return ph.statusChangeCb(defs.UnregisteredStatus, control.SamplerUID(uid), nil)
}

func (SamplerHandler) fromServerMsg(uid string) *protos.ServerToSampler {
	return &protos.ServerToSampler{
		Timestamp: timestamppb.New(time.Now()),
		ServerUid: uid,
	}
}
