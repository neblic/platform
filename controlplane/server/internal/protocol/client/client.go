package client

import (
	"fmt"

	"github.com/neblic/platform/controlplane/data"
	"github.com/neblic/platform/controlplane/protos"
	internalclient "github.com/neblic/platform/controlplane/server/internal/defs/client"
	"github.com/neblic/platform/controlplane/server/internal/protocol/stream"
	"github.com/neblic/platform/controlplane/server/internal/registry"
	"github.com/neblic/platform/logging"
)

type Client struct {
	clientReg  *registry.Client  // It is allowed to modify the client registry
	samplerReg *registry.Sampler // Client handlers can't modify the sampler registry, but can read it

	registeredOnce bool
	stream         *stream.Stream[*protos.ClientToServer, *protos.ServerToClient]

	logger logging.Logger
}

func New(
	logger logging.Logger,
	serverUID string,
	clientReg *registry.Client,
	samplerReg *registry.Sampler,
	opts *stream.Options) *Client {

	if logger == nil {
		logger = logging.NewNopLogger()
	}

	c := &Client{
		clientReg:  clientReg,
		samplerReg: samplerReg,
	}

	c.logger = logger.With("role", "server/client")
	c.stream = stream.New[*protos.ClientToServer, *protos.ServerToClient](
		c.logger, serverUID, opts,
		stream.NewClientHandler(c.recvToServerReqCb, c.streamStateChangeCb),
	)

	return c
}

func (c *Client) HandleStream(s protos.ControlPlane_ClientConnServer) error {
	return c.stream.Handle(s)
}

// recvToServerReqCb implements how to handle requests initiated by the client
func (c *Client) recvToServerReqCb(clientToServerReq *protos.ClientToServer) (bool, *protos.ServerToClient, error) {
	var (
		serverToClientRes *protos.ServerToClient
		err               error
	)

	c.logger.Debug(fmt.Sprintf("Processing %T request", clientToServerReq.Message))

	switch msg := clientToServerReq.Message.(type) {
	case *protos.ClientToServer_ListSamplersReq:
		serverToClientRes, err = c.handleListSamplersReq(msg.ListSamplersReq)
		if err != nil {
			return true, nil, err
		}
	case *protos.ClientToServer_SamplerConfReq:
		serverToClientRes, err = c.handleSamplerConfReq(msg.SamplerConfReq)
		if err != nil {
			return true, nil, err
		}

	default:
		return false, nil, nil
	}

	c.logger.Debug(fmt.Sprintf("Replying %T response", serverToClientRes.Message))

	return true, serverToClientRes, nil
}

// streamStateChangeCb implements the logic to execute when there is a stream state change
func (c *Client) streamStateChangeCb(state internalclient.State, uid data.ClientUID) error {
	switch state {
	case internalclient.Registered:
		if err := c.clientReg.Register(uid); err != nil {
			return err
		}

		if !c.registeredOnce {
			c.logger = c.logger.With("client_uid", uid)
		}

		c.logger.Debug("Client registered")

		c.registeredOnce = true
	case internalclient.Unregistered:
		if err := c.clientReg.Deregister(uid); err != nil {
			return fmt.Errorf("error deregistering client, uid: %s: %w", uid, err)
		}

		c.logger.Debug("Client deregistered")
	default:
		c.logger.Error(fmt.Sprintf("Received unknown state change %s", state))
	}

	return nil
}
