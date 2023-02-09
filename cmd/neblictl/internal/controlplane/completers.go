package controlplane

import (
	"context"
	"sort"

	"github.com/neblic/platform/cmd/neblictl/internal/interpoler"
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
	// Get all samplers
	samplers, _ := c.controlPlaneClient.getSamplers(ctx, true)

	// Store resources in a map to remove duplicates
	resourcesMap := map[string]bool{"*": true}
	for sampler := range samplers {
		resourcesMap[sampler.resource] = true
	}

	// Construct list of samplers
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
	// Get options
	resourceParameter, ok := parameters.Get("resource")

	// Store samplers in a map to remove duplicates
	samplersMap := map[string]bool{"*": true}
	samplersFull, _ := c.controlPlaneClient.getSamplers(ctx, true)
	for resourceAndSamplerEntry := range samplersFull {
		if !ok || (ok && resourceAndSamplerEntry.resource == resourceParameter.Value) {
			samplersMap[resourceAndSamplerEntry.sampler] = true
		}
	}

	// Construct list of samplers
	samplers := []string{}
	for sampler := range samplersMap {
		samplers = append(samplers, sampler)
	}
	sort.Strings(samplers)

	return samplers
}

func (c *Completers) ListRules(ctx context.Context, parameters interpoler.ParametersWithValue) []string {
	// Get options
	resourceParameter, _ := parameters.Get("resource")
	samplerParameter, _ := parameters.Get("sampler")

	sampler, _ := c.controlPlaneClient.getSampler(ctx, samplerParameter.Value, resourceParameter.Value, true)
	if sampler == nil {
		return []string{}
	}

	var rules []string
	for _, samplingRule := range sampler.Config.SamplingRules {
		rules = append(rules, samplingRule.Rule)
	}

	// Construct output
	return rules
}
