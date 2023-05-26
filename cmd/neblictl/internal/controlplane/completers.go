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

// TODO: Generic sampler filter based on supplied parameters
func (c *Completers) ListResourcesUID(ctx context.Context, parameters interpoler.ParametersWithValue) []string {
	samplerParameter, _ := parameters.Get("sampler-name")
	streamUIDParameter, streamUIDParameterOk := parameters.Get("stream-uid")

	samplers, _ := c.controlPlaneClient.getSamplers(ctx, "*", samplerParameter.Value, true)

	// Store resources in a map to remove duplicates
	resourcesMap := map[string]bool{"*": true}
	for _, sampler := range samplers {
		if streamUIDParameterOk {
			for _, stream := range sampler.Config.Streams {
				if streamUIDParameterOk &&
					stream.UID == data.SamplerStreamUID(streamUIDParameter.Value) {
					resourcesMap[sampler.Resource] = true
				}
			}
		} else {
			resourcesMap[sampler.Resource] = true
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

// ListSamplersUID lists all the available samplers. If a resource parameter is provided, it will just
// return the samplers that are part of the resource
func (c *Completers) ListSamplersUID(ctx context.Context, parameters interpoler.ParametersWithValue) []string {
	resourceParameter, _ := parameters.Get("resource-name")
	streamUIDParameter, streamUIDParameterOk := parameters.Get("stream-uid")

	samplers, _ := c.controlPlaneClient.getSamplers(ctx, resourceParameter.Value, "*", true)

	// Store resources in a map to remove duplicates
	samplersMap := map[string]bool{"*": true}
	for _, sampler := range samplers {
		if streamUIDParameterOk {
			for _, stream := range sampler.Config.Streams {
				if stream.UID == data.SamplerStreamUID(streamUIDParameter.Value) {
					samplersMap[sampler.Name] = true
				}
			}
		} else {
			samplersMap[sampler.Name] = true
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
	resourceParameter, _ := parameters.Get("resource-name")
	samplerParameter, _ := parameters.Get("sampler-name")
	digestUIDParameter, digestUIDParameterOk := parameters.Get("digest-uid")

	samplers, _ := c.controlPlaneClient.getSamplers(ctx, resourceParameter.Value, samplerParameter.Value, true)
	if len(samplers) == 0 {
		return []string{}
	}

	uidsSet := map[string]struct{}{}
	for _, sampler := range samplers {
		for _, stream := range sampler.Config.Streams {
			if digestUIDParameterOk {
				for _, digest := range sampler.Config.Digests {
					if digest.UID == data.SamplerDigestUID(digestUIDParameter.Value) {
						uidsSet[string(stream.UID)] = struct{}{}
					}
				}
			} else {
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

func (c *Completers) ListDigestsUID(ctx context.Context, parameters interpoler.ParametersWithValue) []string {
	resourceParameter, _ := parameters.Get("resource-name")
	samplerParameter, _ := parameters.Get("sampler-name")
	streamUIDParameter, streamUIDParameterOk := parameters.Get("stream-uid")

	samplers, _ := c.controlPlaneClient.getSamplers(ctx, resourceParameter.Value, samplerParameter.Value, true)

	// Store resources in a map to remove duplicates
	digestsUIDMap := make(map[string]bool)
	for _, sampler := range samplers {
		for _, digest := range sampler.Config.Digests {
			for _, stream := range sampler.Config.Streams {
				if streamUIDParameterOk && stream.UID == data.SamplerStreamUID(streamUIDParameter.Value) {
					digestsUIDMap[string(digest.UID)] = true
				}
			}
		}
	}

	digestsUID := []string{}
	for resoure := range digestsUIDMap {
		digestsUID = append(digestsUID, resoure)
	}
	sort.Strings(digestsUID)

	return digestsUID
}
