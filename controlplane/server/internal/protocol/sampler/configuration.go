package sampler

import (
	"fmt"

	data "github.com/neblic/platform/controlplane/data"
	"github.com/neblic/platform/controlplane/protos"
)

func (p *Sampler) Configure(samplerConfig *data.SamplerConfig) error {
	var protoSamplerConfig *protos.SamplerConfig
	if samplerConfig != nil {
		protoSamplerConfig = samplerConfig.ToProto()
	}

	req := &protos.ServerToSampler{
		Message: &protos.ServerToSampler_ConfReq{
			ConfReq: &protos.ServerSamplerConfReq{
				SamplerConfig: protoSamplerConfig,
			},
		},
	}

	p.logger.Debug(fmt.Sprintf("Sending %T request", req.Message))

	res, err := p.stream.Send(req)
	if err != nil {
		return fmt.Errorf("error configuring sampler: %w", err)
	}

	configRes, ok := res.GetMessage().(*protos.SamplerToServer_ConfRes)
	if !ok {
		return fmt.Errorf("error configuring sampler: unexpected response type %T", res.GetMessage())
	}

	status := configRes.ConfRes.GetStatus()
	if status.GetType() != protos.Status_OK {
		return fmt.Errorf("error configuring sampler: %s", status.GetErrorMessage())
	}

	return nil
}
