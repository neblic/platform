package sampler

import (
	"fmt"

	"github.com/neblic/platform/controlplane/control"
	"github.com/neblic/platform/controlplane/protos"
	"github.com/neblic/platform/controlplane/server/internal/defs"
	"github.com/neblic/platform/controlplane/server/internal/protocol/stream"
	"github.com/neblic/platform/controlplane/server/internal/registry"
	"github.com/neblic/platform/logging"
)

type Sampler struct {
	samplerRegistry *registry.SamplerRegistry

	registeredOnce bool
	stream         *stream.Stream[*protos.SamplerToServer, *protos.ServerToSampler]

	logger logging.Logger
}

func New(logger logging.Logger, serverUID string, samplerRegistry *registry.SamplerRegistry, opts *stream.Options) *Sampler {
	if logger == nil {
		logger = logging.NewNopLogger()
	}

	p := &Sampler{
		samplerRegistry: samplerRegistry,
	}

	p.logger = logger.With("role", "server/sampler")
	p.stream = stream.New[*protos.SamplerToServer, *protos.ServerToSampler](
		p.logger, serverUID, opts,
		stream.NewSamplerHandler(p.recvToServerReqCb, p.streamStateChangeCb),
	)

	return p
}

func (p *Sampler) HandleStream(stream protos.ControlPlane_SamplerConnServer) error {
	return p.stream.Handle(stream)
}

func (p *Sampler) recvToServerReqCb(clientToServerMsg *protos.SamplerToServer) (bool, *protos.ServerToSampler, error) {
	switch msg := clientToServerMsg.GetMessage().(type) {
	case *protos.SamplerToServer_SamplerStatsMsg:
		err := p.samplerRegistry.UpdateStats(
			clientToServerMsg.Resouce,
			clientToServerMsg.Name,
			control.SamplerUID(clientToServerMsg.GetSamplerUid()),
			control.NewSamplerSamplingStatsFromProto(msg.SamplerStatsMsg.GetSamplingStats()),
		)

		return true, nil, err
	default:
		return false, nil, nil
	}
}

func (p *Sampler) streamStateChangeCb(state defs.Status, uid control.SamplerUID, req *protos.SamplerToServer) error {
	switch state {
	case defs.RegisteredStatus:
		initialConfig := control.NewSamplerConfig()
		initialConfig.Merge(control.NewSamplerConfigUpdateFromProto(req.GetRegisterReq().SamplerConfigUpdate))
		if err := p.samplerRegistry.Register(req.Resouce, req.Name, uid, p, *initialConfig); err != nil {
			return err
		}

		if !p.registeredOnce {
			p.logger = p.logger.With("sampler_uid", uid, "sampler_name", req.Name, "sampler_resource", req.Resouce)
		}

		p.logger.Debug("sampler registered")

		p.registeredOnce = true
	case defs.UnregisteredStatus:
		// Unregistered status could happen when a connection is closed. In that case, name and resource
		// parameters are empty.
		if err := p.samplerRegistry.Deregister(uid); err != nil {
			return fmt.Errorf("error deregistering client, uid: %s: %w", uid, err)
		}

		p.logger.Debug("sampler unregistered")
	}

	return nil
}
