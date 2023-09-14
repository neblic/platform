package controlplane

import (
	"context"
	"sort"

	"github.com/neblic/platform/cmd/neblictl/internal/interpoler"
	"github.com/neblic/platform/controlplane/control"
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
	streamUIDValue := "*"
	streamUIDParameter, streamUIDParameterOk := parameters.Get("stream-uid")
	if streamUIDParameterOk && streamUIDParameter.Filter {
		streamUIDValue = streamUIDParameter.Value
	}

	samplers, _ := c.controlPlaneClient.getSamplers(ctx, "*", samplerParameter.Value, streamUIDValue, true)

	// Store resources in a map to remove duplicates
	resourcesMap := map[string]bool{"*": true}
	for _, sampler := range samplers {
		resourcesMap[sampler.Resource] = true
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
	streamUIDValue := "*"
	streamUIDParameter, streamUIDParameterOk := parameters.Get("stream-uid")
	if streamUIDParameterOk && streamUIDParameter.Filter {
		streamUIDValue = streamUIDParameter.Value
	}

	samplers, _ := c.controlPlaneClient.getSamplers(ctx, resourceParameter.Value, "*", streamUIDValue, true)

	// Store resources in a map to remove duplicates
	samplersMap := map[string]bool{"*": true}
	for _, sampler := range samplers {
		samplersMap[sampler.Name] = true
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

	samplers, _ := c.controlPlaneClient.getSamplers(ctx, resourceParameter.Value, samplerParameter.Value, "*", true)
	if len(samplers) == 0 {
		return []string{}
	}

	uidsSet := map[string]struct{}{}
	for _, sampler := range samplers {
		for _, stream := range sampler.Config.Streams {
			if digestUIDParameterOk && digestUIDParameter.Filter {
				for _, digest := range sampler.Config.Digests {
					if digest.UID == control.SamplerDigestUID(digestUIDParameter.Value) {
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
	streamUIDValue := "*"
	streamUIDParameter, streamUIDParameterOk := parameters.Get("stream-uid")
	if streamUIDParameterOk {
		streamUIDValue = streamUIDParameter.Value
	}

	samplers, _ := c.controlPlaneClient.getSamplers(ctx, resourceParameter.Value, samplerParameter.Value, streamUIDValue, true)

	// Store resources in a map to remove duplicates
	digestsUIDMap := make(map[string]bool)
	for _, sampler := range samplers {
		for _, digest := range sampler.Config.Digests {
			digestsUIDMap[string(digest.UID)] = true
		}
	}

	// Create deduplicated list of digest uids
	digestUids := []string{}
	for digestUid := range digestsUIDMap {
		digestUids = append(digestUids, digestUid)
	}
	sort.Strings(digestUids)

	return digestUids
}

func (c *Completers) ListSampleType(ctx context.Context, parameters interpoler.ParametersWithValue) []string {
	sampleTypes := []string{}
	for _, sampleType := range control.ValidSampleTypes {
		sampleTypes = append(sampleTypes, sampleType.String())
	}
	return sampleTypes
}

func (c *Completers) ListEventsUID(ctx context.Context, parameters interpoler.ParametersWithValue) []string {
	resourceParameter, _ := parameters.Get("resource-name")
	samplerParameter, _ := parameters.Get("sampler-name")

	streamUIDValue := "*"
	streamUIDParameter, streamUIDParameterOk := parameters.Get("stream-uid")
	if streamUIDParameterOk {
		streamUIDValue = streamUIDParameter.Value
	}

	samplers, _ := c.controlPlaneClient.getSamplers(ctx, resourceParameter.Value, samplerParameter.Value, streamUIDValue, true)

	// Store resources in a map to remove duplicates
	eventsUIDMap := make(map[string]bool)
	for _, sampler := range samplers {
		for _, event := range sampler.Config.Events {
			eventsUIDMap[string(event.UID)] = true
		}
	}

	// Create deduplicated list of event uids
	eventsUID := []string{}
	for resoure := range eventsUIDMap {
		eventsUID = append(eventsUID, resoure)
	}
	sort.Strings(eventsUID)

	return eventsUID
}
