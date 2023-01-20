package sampler

import (
	"fmt"

	"github.com/neblic/platform/controlplane/data"
	"github.com/neblic/platform/controlplane/protos"
	internalsampler "github.com/neblic/platform/controlplane/server/internal/defs/sampler"
	"github.com/neblic/platform/controlplane/server/internal/protocol/stream"
	"github.com/neblic/platform/controlplane/server/internal/registry"
	"github.com/neblic/platform/logging"
)

type Sampler struct {
	samplerReg *registry.Sampler

	registeredOnce bool
	stream         *stream.Stream[*protos.SamplerToServer, *protos.ServerToSampler]

	logger logging.Logger
}

func New(logger logging.Logger, serverUID string, samplerReg *registry.Sampler, opts *stream.Options) *Sampler {
	if logger == nil {
		logger = logging.NewNopLogger()
	}

	p := &Sampler{
		samplerReg: samplerReg,
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
		err := p.samplerReg.UpdateSamplerStats(
			data.SamplerUID(clientToServerMsg.GetSamplerUid()),
			data.NewSamplerSamplingStatsFromProto(msg.SamplerStatsMsg.GetSamplingStats()),
		)

		return true, nil, err
	default:
		return false, nil, nil
	}
}

func (p *Sampler) streamStateChangeCb(state internalsampler.State, name, resource string, uid data.SamplerUID) error {
	switch state {
	case internalsampler.Registered:
		if err := p.samplerReg.Register(uid, name, resource, p); err != nil {
			return err
		}

		if !p.registeredOnce {
			p.logger = p.logger.With("sampler_uid", uid, "sampler_name", name, "sampler_resource", resource)
		}

		p.logger.Debug("sampler registered")

		p.registeredOnce = true
	case internalsampler.Unregistered:
		if err := p.samplerReg.Deregister(uid); err != nil {
			return fmt.Errorf("error deregistering client, uid: %s: %w", uid, err)
		}

		p.logger.Debug("sampler unregistered")
	}

	return nil
}
