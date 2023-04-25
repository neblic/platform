package controlplane

import (
	"context"
	"sort"

	"github.com/neblic/platform/cmd/neblictl/internal/interpoler"
	"github.com/neblic/platform/controlplane/data"
)

type Completers struct {
	controlPlaneClient *Client
}

func NewCompleters(controlPlaneClient *Client) *Completers {
	return &Completers{
		controlPlaneClient: controlPlaneClient,
	}
}

func (c *Completers) ListResources(ctx context.Context, parameters interpoler.ParametersWithValue) []string {
	resourceParameter, _ := parameters.Get("resource")
	samplerParameter, _ := parameters.Get("sampler")
	streamUIDParameter, streamUIDParameterOk := parameters.Get("uid")

	samplers, _ := c.controlPlaneClient.getSamplers(ctx, resourceParameter.Value, samplerParameter.Value, true)

	// Store resources in a map to remove duplicates
	resourcesMap := map[string]bool{"*": true}
	for _, sampler := range samplers {
		if !streamUIDParameterOk || streamUIDParameter.DoNotFilter {
			resourcesMap[sampler.Resource] = true
		} else {
			for _, stream := range sampler.Config.Streams {
				if streamUIDParameterOk && stream.UID == data.SamplerStreamUID(streamUIDParameter.Value) {
					resourcesMap[sampler.Resource] = true
				}
			}
		}
	}

	// Construct list of resources
	resources := []string{}
	for resoure := range resourcesMap {
		resources = append(resources, resoure)
	}
	sort.Strings(resources)

	return resources
}

// ListSamplers lists all the available samplers. If a resource parameter is provided, it will just
// return the samplers that are part of the resource
func (c *Completers) ListSamplers(ctx context.Context, parameters interpoler.ParametersWithValue) []string {
	resourceParameter, _ := parameters.Get("resource")
	samplerParameter, _ := parameters.Get("sampler")
	streamUIDParameter, streamUIDParameterOk := parameters.Get("uid")

	samplers, _ := c.controlPlaneClient.getSamplers(ctx, resourceParameter.Value, samplerParameter.Value, true)

	// Store resources in a map to remove duplicates
	samplersMap := map[string]bool{"*": true}
	for _, sampler := range samplers {
		if !streamUIDParameterOk || streamUIDParameter.DoNotFilter {
			samplersMap[sampler.Name] = true
		} else {
			for _, stream := range sampler.Config.Streams {
				if streamUIDParameterOk && stream.UID == data.SamplerStreamUID(streamUIDParameter.Value) {
					samplersMap[sampler.Name] = true
				}
			}
		}
	}

	// Construct list of sampler names
	samplersName := []string{}
	for resoure := range samplersMap {
		samplersName = append(samplersName, resoure)
	}
	sort.Strings(samplersName)

	return samplersName
}

func (c *Completers) ListStreamsUID(ctx context.Context, parameters interpoler.ParametersWithValue) []string {
	resourceParameter, _ := parameters.Get("resource")
	samplerParameter, _ := parameters.Get("sampler")
	streamUIDParameter, streamUIDParameterOk := parameters.Get("uid")

	samplers, _ := c.controlPlaneClient.getSamplers(ctx, resourceParameter.Value, samplerParameter.Value, true)
	if len(samplers) == 0 {
		return []string{}
	}

	uidsSet := map[string]struct{}{}
	for _, sampler := range samplers {
		for _, stream := range sampler.Config.Streams {
			if !streamUIDParameterOk || streamUIDParameter.DoNotFilter || (streamUIDParameterOk && stream.UID == data.SamplerStreamUID(streamUIDParameter.Value)) {
				uidsSet[string(stream.UID)] = struct{}{}
			}
		}
	}
	var uids []string
	for uid := range uidsSet {
		uids = append(uids, uid)
	}
	sort.Strings(uids)

	// Construct output
	return uids
}
