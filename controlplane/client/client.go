package client

import (
	"context"
	"fmt"
	"time"

	"github.com/neblic/platform/controlplane/control"
	"github.com/neblic/platform/controlplane/internal/stream"
	"github.com/neblic/platform/controlplane/protos"
	"github.com/neblic/platform/logging"
)

type Client struct {
	uid  control.ClientUID
	opts *options

	clientStream *stream.Stream[*protos.ServerToClient, *protos.ClientToServer]
	stateChanges chan State

	logger logging.Logger
}

func New(uid string, clientOptions ...Option) *Client {
	opts := newDefaultOptions()
	for _, opt := range clientOptions {
		opt.apply(opts)
	}

	c := &Client{
		uid:  control.ClientUID(uid),
		opts: opts,
	}

	c.logger = opts.logger.With("role", "client", "client_uid", string(uid))
	c.clientStream = stream.New(
		string(c.uid),
		opts.streamOpts,
		stream.NewClientHandler(c.recvServerReqCb),
		c.logger,
	)

	return c
}

func (c *Client) Connect(serverAddr string) error {
	return c.clientStream.Connect(serverAddr)
}

func (c *Client) State() State {
	return NewStateFromStreamState(c.clientStream.State())
}

func (c *Client) StateChanges() chan State {
	if c.stateChanges == nil {
		c.stateChanges = make(chan State)
	}

	go func() {
	loop:
		for {
			streamState, more := <-c.clientStream.StateChanges()
			if !more {
				break loop
			}

			c.stateChanges <- NewStateFromStreamState(streamState)
		}
	}()

	return c.stateChanges
}

func (c *Client) Errors() chan error {
	// for now, we can forward errors directly from the stream to the user,
	// if we ever need to send errors from the client itself in addition to stream errors,
	// we will need to create a new channel
	return c.clientStream.Errors()
}

func (c *Client) UID() control.ClientUID {
	return c.uid
}

func (c *Client) ListSamplers(ctx context.Context) ([]*control.Sampler, error) {
	req := c.clientStream.ToServerMsg()
	req.Message = &protos.ClientToServer_ListSamplersReq{
		ListSamplersReq: &protos.ClientListSamplersReq{},
	}

	c.logger.Debug(fmt.Sprintf("Sending %T request", req.Message))

	res, err := c.clientStream.SendReqToS(ctx, req)
	if err != nil {
		return nil, err
	}

	listSamplersRes, ok := res.GetMessage().(*protos.ServerToClient_ListSamplersRes)
	if !ok {
		return nil, fmt.Errorf("received unexpected list samplers response type %T", res.GetMessage())
	}

	status := listSamplersRes.ListSamplersRes.GetStatus()
	if status.GetType() != protos.Status_OK {
		return nil, fmt.Errorf("error getting samplers list: %s", status.GetErrorMessage())
	}

	var samplers []*control.Sampler
	for _, protosSampler := range listSamplersRes.ListSamplersRes.GetSamplers() {
		samplers = append(samplers, control.NewSamplerFromProto(protosSampler))
	}

	return samplers, nil
}

// ConfigureSampler sends a configuration to a sampler
func (c *Client) ConfigureSampler(ctx context.Context, samplerResource, samplerName string, update *control.SamplerConfigUpdate) error {
	req := c.clientStream.ToServerMsg()
	req.Message = &protos.ClientToServer_SamplerConfReq{
		SamplerConfReq: &protos.ClientSamplerConfReq{
			SamplerName:         samplerName,
			SamplerResource:     samplerResource,
			SamplerConfigUpdate: update.ToProto(),
		},
	}

	c.logger.Debug(fmt.Sprintf("Sending %T request", req.Message))

	res, err := c.clientStream.SendReqToS(ctx, req)
	if err != nil {
		return err
	}

	samplerConfRes, ok := res.GetMessage().(*protos.ServerToClient_SamplerConfRes)
	if !ok {
		return fmt.Errorf("received unexpected configure rule rules response type %T", res.GetMessage())
	}

	status := samplerConfRes.SamplerConfRes.GetStatus()
	if status.GetType() != protos.Status_OK {
		return fmt.Errorf("%s", status.GetErrorMessage())
	}

	return nil
}

func (c *Client) Close(timeout time.Duration) error {
	if err := c.clientStream.Close(timeout); err != nil {
		return fmt.Errorf("error closing stream: %w", err)
	}

	return nil
}

func (c *Client) recvServerReqCb(serverMsg *protos.ServerToClient) (bool, *protos.ClientToServer, error) {
	switch serverMsg.GetMessage().(type) {
	case *protos.ServerToClient_SamplerStatsMsg:
		return true, nil, nil
	default:
		return false, nil, nil
	}
}
