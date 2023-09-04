package client

import (
	"fmt"

	"github.com/neblic/platform/controlplane/control"
	"github.com/neblic/platform/controlplane/protos"
	"github.com/neblic/platform/controlplane/server/internal/defs"
	"github.com/neblic/platform/controlplane/server/internal/protocol/stream"
	"github.com/neblic/platform/controlplane/server/internal/registry"
	"github.com/neblic/platform/logging"
)

type Client struct {
	clientRegistry  *registry.ClientRegistry
	samplerRegistry *registry.SamplerRegistry

	registeredOnce bool
	stream         *stream.Stream[*protos.ClientToServer, *protos.ServerToClient]

	logger logging.Logger
}

func New(
	logger logging.Logger,
	serverUID string,
	clientRegistry *registry.ClientRegistry,
	samplerRegistry *registry.SamplerRegistry,
	opts *stream.Options) *Client {

	if logger == nil {
		logger = logging.NewNopLogger()
	}

	c := &Client{
		clientRegistry:  clientRegistry,
		samplerRegistry: samplerRegistry,
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
func (c *Client) streamStateChangeCb(state defs.Status, uid control.ClientUID) error {
	switch state {
	case defs.RegisteredStatus:
		if err := c.clientRegistry.Register(uid); err != nil {
			return err
		}

		if !c.registeredOnce {
			c.logger = c.logger.With("client_uid", uid)
		}

		c.logger.Debug("Client registered")

		c.registeredOnce = true
	case defs.UnregisteredStatus:
		if err := c.clientRegistry.Deregister(uid); err != nil {
			return fmt.Errorf("error deregistering client, uid: %s: %w", uid, err)
		}

		c.logger.Debug("Client deregistered")
	default:
		c.logger.Error(fmt.Sprintf("Received unknown state change %s", state))
	}

	return nil
}
