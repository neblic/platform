package controlplane

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/neblic/platform/controlplane/client"
	"github.com/neblic/platform/controlplane/data"
)

type Client struct {
	mutex *sync.RWMutex

	internal       *client.Client
	registeredOnce bool
	currentState   client.State

	samplers    map[resourceAndSampler]*data.Sampler
	samplerUIDs map[resourceAndSampler]data.SamplerUID
}

func NewClient(uuid string, address string, opts ...client.Option) (*Client, error) {
	client := &Client{
		mutex:       new(sync.RWMutex),
		internal:    client.New(uuid, opts...),
		samplers:    map[resourceAndSampler]*data.Sampler{},
		samplerUIDs: map[resourceAndSampler]data.SamplerUID{},
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

	c.samplers = map[resourceAndSampler]*data.Sampler{}
	c.samplerUIDs = map[resourceAndSampler]data.SamplerUID{}
	for _, sampler := range samplers {
		key := resourceAndSampler{
			resource: sampler.Resource,
			sampler:  sampler.Name,
		}
		c.samplers[key] = sampler
		c.samplerUIDs[key] = sampler.UID
	}

	return nil
}

func (c *Client) getSamplers(ctx context.Context, cached bool) (map[resourceAndSampler]*data.Sampler, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	var err error
	if !cached {
		err = c.pullSamplerConfigs(ctx)
	}
	return c.samplers, err
}

func (c *Client) getSampler(ctx context.Context, name, resource string, cached bool) (*data.Sampler, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	samplerConfigs, err := c.getSamplers(ctx, cached)
	return samplerConfigs[resourceAndSampler{resource, name}], err
}

func (c *Client) setSamplerConfig(ctx context.Context, name, resource string, update *data.SamplerConfigUpdate) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	return c.internal.ConfigureSampler(ctx, name, resource, "", update)
}
