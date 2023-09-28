package controlplane

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/neblic/platform/controlplane/client"
	"github.com/neblic/platform/controlplane/control"
)

type Client struct {
	mutex *sync.RWMutex

	internal       *client.Client
	registeredOnce bool
	currentState   client.State

	samplers map[resourceAndSampler]*control.Sampler
}

func NewClient(uuid string, address string, opts ...client.Option) (*Client, error) {
	client := &Client{
		mutex:    new(sync.RWMutex),
		internal: client.New(uuid, opts...),
		samplers: map[resourceAndSampler]*control.Sampler{},
	}

	notifyFailedFirstRegistration := make(chan error)
	go client.handleErrors(notifyFailedFirstRegistration)

	notifyFirstRegistration := make(chan struct{})
	go client.handleStateChanges(notifyFirstRegistration)

	err := client.internal.Connect(address)
	if err != nil {
		return nil, fmt.Errorf("control plane server could not be reached: %v", err)
	}

	// wait until client registered
	select {
	case err := <-notifyFailedFirstRegistration:
		return nil, err
	case <-notifyFirstRegistration:
	}

	// Perfrom initial pulling of config
	ctx, _ := context.WithTimeout(context.Background(), time.Second*30)
	err = client.pullSamplerConfigs(ctx)

	return client, err
}

func (c *Client) handleErrors(notifyFailedFirstRegistration chan error) {
	for {
		err := <-c.internal.Errors()
		if errors.Is(err, client.ErrRegistrationFailure) {
			if !c.registeredOnce && notifyFailedFirstRegistration != nil {
				notifyFailedFirstRegistration <- err
				continue
			}
		}
	}
}

func (c *Client) handleStateChanges(notifyFirstRegistration chan struct{}) {
	for {
		c.currentState = <-c.internal.StateChanges()
		if c.currentState == client.Registered {
			if notifyFirstRegistration != nil {
				notifyFirstRegistration <- struct{}{}
			}
			c.registeredOnce = true
		}
	}
}

func (c *Client) pullSamplerConfigs(ctx context.Context) error {
	// Get all samplers
	samplers, err := c.internal.ListSamplers(ctx)
	if err != nil {
		return err
	}

	c.samplers = map[resourceAndSampler]*control.Sampler{}
	for _, sampler := range samplers {
		key := resourceAndSampler{
			resource: sampler.Resource,
			sampler:  sampler.Name,
		}
		c.samplers[key] = sampler
	}

	return nil
}

func (c *Client) getAllSamplers(ctx context.Context, cached bool) (map[resourceAndSampler]*control.Sampler, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	var err error
	if !cached {
		err = c.pullSamplerConfigs(ctx)
	}
	return c.samplers, err
}

func doesResourceAndSamplerMatch(resourceParameter, samplerParameter string, resourceAndSamplerEntry resourceAndSampler) bool {
	acceptAllResources := resourceParameter == "*"
	acceptAllSamplers := samplerParameter == "*"

	// This part of the logic explores all four possible combinations.
	// 1) ASSUMPTION: resourceParameter==* and SamplerParameter==*
	allResourcesAndAllSamplers := acceptAllResources && acceptAllSamplers
	// 2) ASSUMPTION: resourceParameter==* and samplerParameter==<sampler> CONDITION resourceAndSamplerEntry.sampler==<sampler>
	allResourcesAndMatchingSampler := acceptAllResources && resourceAndSamplerEntry.sampler == samplerParameter
	// 3) ASSUMPTION resourceAndSamplerEntry.resrouce==<resource> and resourceParameter==* CONDITION resourceParameter==<resource>
	matchingResourceAndAllSamplers := acceptAllSamplers && resourceAndSamplerEntry.resource == resourceParameter
	// 4) ASSUMPTION resourceParameter==<resource> and samplerParameter==<sampler> CONDITION resourceAndSamplerEntry.resource==<resource> and resourceAndSamplerEntry.sampler==<sampler>
	matchingResourceAndMatchingSampler := resourceAndSamplerEntry.resource == resourceParameter && resourceAndSamplerEntry.sampler == samplerParameter

	return allResourcesAndAllSamplers || allResourcesAndMatchingSampler || matchingResourceAndAllSamplers || matchingResourceAndMatchingSampler

}

func (c *Client) getSamplers(ctx context.Context, resourceParameter string, samplerParameter string, streamNameParameter string, cached bool) (map[resourceAndSampler]*control.Sampler, error) {
	samplers, err := c.getAllSamplers(ctx, cached)

	// Iterate over all samplers and select the ones matching the input
	resourceAndSamplers := map[resourceAndSampler]*control.Sampler{}
	for resourceAndSamplerEntry, samplerData := range samplers {
		if doesResourceAndSamplerMatch(resourceParameter, samplerParameter, resourceAndSamplerEntry) {

			// Check if stream with the provided name exists
			var ok bool
			for _, stream := range samplerData.Config.Streams {
				if stream.Name == streamNameParameter {
					ok = true
					break
				}
			}

			if streamNameParameter != "" && streamNameParameter != "*" && !ok {
				// stream parameter contains a valid name that the current sampler does not contain. Skip it
				continue
			}

			resourceAndSamplers[resourceAndSampler{resourceAndSamplerEntry.resource, resourceAndSamplerEntry.sampler}] = samplerData
		}
	}

	// Return error if no matching sampler was found
	if len(resourceAndSamplers) == 0 {
		var err error
		if resourceParameter == "*" || samplerParameter == "*" {
			err = fmt.Errorf("could not find any sampler matching the criteria")
		} else {
			err = fmt.Errorf("sampler does not exist")
		}
		return resourceAndSamplers, err
	}

	return resourceAndSamplers, err
}

func (c *Client) setSamplerConfig(ctx context.Context, name, resource string, update *control.SamplerConfigUpdate) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	return c.internal.ConfigureSampler(ctx, resource, name, update)
}
