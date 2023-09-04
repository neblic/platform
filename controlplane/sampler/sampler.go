package sampler

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/neblic/platform/controlplane/control"
	"github.com/neblic/platform/controlplane/internal/stream"
	"github.com/neblic/platform/controlplane/protos"
	"github.com/neblic/platform/logging"
)

type Sampler struct {
	data *control.Sampler
	opts *options

	samplerStream *stream.Stream[*protos.ServerToSampler, *protos.SamplerToServer]
	events        chan Event

	logger logging.Logger
}

func New(name, resource string, samplerOptions ...Option) *Sampler {
	opts := newDefaultSamplerOptions()
	for _, opt := range samplerOptions {
		opt.apply(opts)
	}

	uid := control.SamplerUID(uuid.NewString())
	p := &Sampler{
		data: control.NewSampler(name, resource, uid),
		opts: opts,
	}

	p.logger = opts.logger.With("role", "sampler", "sampler_uid", string(uid))
	// TODO
	// Instead of creating a stream (therefore a new network connection) in each initialization, allow reusing the same one.
	// To do so, we would need to decouple the grpc client inside the stream from the stream connection so it is just
	// initialized once and allow the creation of new connections reusing the grpc client.
	p.samplerStream = stream.New(
		string(p.data.UID),
		opts.streamOpts,
		stream.NewSamplerHandler(p.data.Name, p.data.Resource, p.recvServerReqCb),
		p.logger,
	)

	return p
}

func (p *Sampler) Connect(serverAddr string) error {
	return p.samplerStream.Connect(serverAddr)
}

func (p *Sampler) State() State {
	return NewStateFromStreamState(p.samplerStream.State())
}

func (p *Sampler) sendEvent(ev Event) {
	if p.events != nil {
		p.events <- ev
	}
}

func (p *Sampler) Events() chan Event {
	if p.events == nil {
		p.events = make(chan Event)
	}

	// forward stream events
	go func() {
	loop:
		for {
			streamState, more := <-p.samplerStream.StateChanges()
			if !more {
				break loop
			}
			p.sendEvent(StateUpdate{State: NewStateFromStreamState(streamState)})
		}
	}()

	return p.events
}

func (p *Sampler) Name() string {
	return p.data.Name
}

func (p *Sampler) UID() control.SamplerUID {
	return p.data.UID
}

func (p *Sampler) Config() control.SamplerConfig {
	return p.data.Config
}

func (p *Sampler) recvServerReqCb(serverToSamplerReq *protos.ServerToSampler) (bool, *protos.SamplerToServer, error) {
	var (
		samplerToServerRes *protos.SamplerToServer
		err                error
	)

	p.logger.Debug(fmt.Sprintf("Processing %T request", serverToSamplerReq.Message))

	switch msg := serverToSamplerReq.GetMessage().(type) {
	case *protos.ServerToSampler_ConfReq:
		if samplerToServerRes, err = p.handleConfigurationRequest(msg.ConfReq); err != nil {
			return true, nil, err
		}
	default:
		return false, nil, nil
	}

	p.logger.Debug(fmt.Sprintf("Replying %T response", samplerToServerRes.Message))

	return true, samplerToServerRes, nil
}

// handleConfigurationRequest sets the sampler configuration (e.g. its rule rules)
// the configuration replaces the previous configuration but he operation is idempotent
// so it should only apply the differences between the new and previous configuration
func (p *Sampler) handleConfigurationRequest(req *protos.ServerSamplerConfReq) (*protos.SamplerToServer, error) {
	samplerConfig := control.NewSamplerConfigFromProto(req.GetSamplerConfig())

	p.data.Config = samplerConfig
	p.sendEvent(ConfigUpdate{
		Config: samplerConfig,
	})

	res := p.samplerStream.ToServerMsg()
	res.Message = &protos.SamplerToServer_ConfRes{
		ConfRes: &protos.ServerSamplerConfRes{
			Status: &protos.Status{
				Type: protos.Status_OK,
			},
		},
	}

	return res, nil
}

func (p *Sampler) UpdateStats(ctx context.Context, stats control.SamplerSamplingStats) error {
	p.data.SamplingStats = stats

	msg := p.samplerStream.ToServerMsg()
	msg.Message = &protos.SamplerToServer_SamplerStatsMsg{
		SamplerStatsMsg: &protos.SamplerStatsMsg{
			SamplingStats: stats.ToProto(),
		},
	}

	if err := p.samplerStream.SendToS(ctx, msg); err != nil {
		return fmt.Errorf("error sending updated stats to server: %s", err)
	}

	return nil
}

func (p *Sampler) Close(timeout time.Duration) error {
	if err := p.samplerStream.Close(timeout); err != nil {
		return fmt.Errorf("error closing stream: %w", err)
	}

	return nil
}
