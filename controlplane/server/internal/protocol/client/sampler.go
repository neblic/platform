package client

import (
	"fmt"

	"github.com/neblic/platform/controlplane/control"
	"github.com/neblic/platform/controlplane/protos"
	"github.com/neblic/platform/controlplane/server/internal/defs"
)

func (c *Client) handleListSamplersReq(_ *protos.ClientListSamplersReq) (*protos.ServerToClient, error) {

	var protoSamplers []*protos.Sampler
	c.samplerRegistry.RangeSamplers(func(sampler *defs.Sampler) (carryon bool) {
		samplingStats := control.SamplerSamplingStats{}
		for _, instance := range sampler.Instances {
			samplingStats.SamplesEvaluated += instance.Stats.SamplesEvaluated
			samplingStats.SamplesExported += instance.Stats.SamplesExported
			samplingStats.SamplesDigested += instance.Stats.SamplesDigested

		}
		protoSamplers = append(protoSamplers, &protos.Sampler{
			Uid:            "",
			Resource:       sampler.Resource,
			Name:           sampler.Name,
			Tags:           sampler.Tags.ToProto(),
			Capabilities:   sampler.Capabilities.ToProto(),
			Config:         sampler.Config.ToProto(),
			SamplingStats:  samplingStats.ToProto(),
			CollectorStats: sampler.CollectorStats.ToProto(),
		})

		// We want to carry on until all the registered instances have been processed
		return true
	})

	serverToClientRes := c.stream.FromServerMsg()
	serverToClientRes.Message = &protos.ServerToClient_ListSamplersRes{
		ListSamplersRes: &protos.ClientListSamplersRes{
			Samplers: protoSamplers,
			Status: &protos.Status{
				Type: protos.Status_OK,
			},
		},
	}

	return serverToClientRes, nil
}

func (c *Client) handleSamplerConfReq(req *protos.ClientSamplerConfReq) (*protos.ServerToClient, error) {
	if req.GetSamplerConfigUpdate() == nil {
		if err := c.samplerRegistry.DeleteSamplerConfig(
			req.GetSamplerResource(),
			req.GetSamplerName(),
		); err != nil {
			return nil, fmt.Errorf("error deleting sampler configuration: %w", err)
		}
	} else {
		update := control.NewSamplerConfigUpdateFromProto(req.GetSamplerConfigUpdate())

		err := update.IsValid()
		if err != nil {
			serverToClientRes := c.stream.FromServerMsg()
			serverToClientRes.Message = &protos.ServerToClient_SamplerConfRes{
				SamplerConfRes: &protos.ClientSamplerConfRes{
					Status: &protos.Status{
						Type:         protos.Status_BAD_REQUEST,
						ErrorMessage: err.Error(),
					},
				},
			}

			return serverToClientRes, nil
		}

		if err := c.samplerRegistry.UpdateSamplerConfig(
			req.GetSamplerResource(),
			req.GetSamplerName(),
			update,
		); err != nil {
			return nil, fmt.Errorf("error updating sampler configuration: %w", err)
		}
	}

	serverToClientRes := c.stream.FromServerMsg()
	serverToClientRes.Message = &protos.ServerToClient_SamplerConfRes{
		SamplerConfRes: &protos.ClientSamplerConfRes{
			Status: &protos.Status{
				Type: protos.Status_OK,
			},
		},
	}

	return serverToClientRes, nil
}
