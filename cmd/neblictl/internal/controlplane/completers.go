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
	streamNameValue := "*"
	streamNameParameter, streamNameParameterOk := parameters.Get("stream-name")
	if streamNameParameterOk && streamNameParameter.Filter {
		streamNameValue = streamNameParameter.Value
	}

	samplers, _ := c.controlPlaneClient.getSamplers(ctx, "*", samplerParameter.Value, streamNameValue, true)

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
	streamNameValue := "*"
	streamNameParameter, streamNameParameterOk := parameters.Get("stream-name")
	if streamNameParameterOk && streamNameParameter.Filter {
		streamNameValue = streamNameParameter.Value
	}

	samplers, _ := c.controlPlaneClient.getSamplers(ctx, resourceParameter.Value, "*", streamNameValue, true)

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

func (c *Completers) ListStreamsName(ctx context.Context, parameters interpoler.ParametersWithValue) []string {
	resourceParameter, _ := parameters.Get("resource-name")
	samplerParameter, _ := parameters.Get("sampler-name")
	digestNameParameter, digestNameParameterOk := parameters.Get("digest-name")

	samplers, _ := c.controlPlaneClient.getSamplers(ctx, resourceParameter.Value, samplerParameter.Value, "*", true)
	if len(samplers) == 0 {
		return []string{}
	}

	namesSet := map[string]struct{}{}
	for _, sampler := range samplers {
		for _, stream := range sampler.Config.Streams {
			if digestNameParameterOk && digestNameParameter.Filter {
				for _, digest := range sampler.Config.Digests {
					if digest.Name == digestNameParameter.Value {
						namesSet[stream.Name] = struct{}{}
					}
				}
			} else {
				namesSet[string(stream.Name)] = struct{}{}
			}
		}
	}

	var names []string
	for name := range namesSet {
		names = append(names, name)
	}
	sort.Strings(names)

	// Construct output
	return names
}

func (c *Completers) ListDigestsName(ctx context.Context, parameters interpoler.ParametersWithValue) []string {
	resourceParameter, _ := parameters.Get("resource-name")
	samplerParameter, _ := parameters.Get("sampler-name")
	streamNameValue := "*"
	streamNameParameter, streamNameParameterOk := parameters.Get("stream-name")
	if streamNameParameterOk {
		streamNameValue = streamNameParameter.Value
	}

	samplers, _ := c.controlPlaneClient.getSamplers(ctx, resourceParameter.Value, samplerParameter.Value, streamNameValue, true)

	// Store resources in a map to remove duplicates
	digestsNameMap := make(map[string]bool)
	for _, sampler := range samplers {
		for _, digest := range sampler.Config.Digests {
			digestsNameMap[string(digest.Name)] = true
		}
	}

	// Create deduplicated list of digest names
	digestNames := []string{}
	for digestName := range digestsNameMap {
		digestNames = append(digestNames, digestName)
	}
	sort.Strings(digestNames)

	return digestNames
}

func (c *Completers) ListSampleType(ctx context.Context, parameters interpoler.ParametersWithValue) []string {
	return []string{control.RawSampleType.String()}
}

func (c *Completers) ListEventsName(ctx context.Context, parameters interpoler.ParametersWithValue) []string {
	resourceParameter, _ := parameters.Get("resource-name")
	samplerParameter, _ := parameters.Get("sampler-name")

	streamNameValue := "*"
	streamNameParameter, streamNameParameterOk := parameters.Get("stream-name")
	if streamNameParameterOk {
		streamNameValue = streamNameParameter.Value
	}

	samplers, _ := c.controlPlaneClient.getSamplers(ctx, resourceParameter.Value, samplerParameter.Value, streamNameValue, true)

	// Store resources in a map to remove duplicates
	eventsNameMap := make(map[string]bool)
	for _, sampler := range samplers {
		for _, event := range sampler.Config.Events {
			eventsNameMap[string(event.Name)] = true
		}
	}

	// Create deduplicated list of event names
	eventsName := []string{}
	for eventName := range eventsNameMap {
		eventsName = append(eventsName, eventName)
	}
	sort.Strings(eventsName)

	return eventsName
}
