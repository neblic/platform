package client

import (
	"fmt"

	"github.com/neblic/platform/controlplane/data"
	"github.com/neblic/platform/controlplane/protos"
)

func (c *Client) handleListSamplersReq(_ *protos.ClientListSamplersReq) (*protos.ServerToClient, error) {
	samplers := c.samplerReg.GetRegisteredSamplers()
	var protoSamplers []*protos.Sampler
	for _, p := range samplers {
		protoSamplers = append(protoSamplers, p.Data.ToProto())
	}

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
		if err := c.clientReg.DeleteSamplerConfig(
			data.SamplerUID(req.GetSamplerUid()),
			req.GetSamplerName(),
			req.GetSamplerResource(),
		); err != nil {
			return nil, fmt.Errorf("error deleting sampler configuration: %w", err)
		}
	} else {
		update := data.NewSamplerConfigUpdateFromProto(req.GetSamplerConfigUpdate())

		if err := c.clientReg.UpdateSamplerConfig(
			data.SamplerUID(req.GetSamplerUid()),
			req.GetSamplerName(),
			req.GetSamplerResource(),
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
