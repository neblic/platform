package registry

import (
	"errors"
	"fmt"
	"sync"

	data "github.com/neblic/platform/controlplane/data"
	internalclient "github.com/neblic/platform/controlplane/server/internal/defs/client"
	"github.com/neblic/platform/logging"
)

var (
	ErrUnkownClient = errors.New("unknown client")
)

type Client struct {
	clients  map[data.ClientUID]*internalclient.Client
	configs  *ConfigDB
	eventsCh chan Event

	logger logging.Logger
	sync.Mutex
}

func NewClient(logger logging.Logger, opts *Options) (*Client, error) {
	if logger == nil {
		logger = logging.NewNopLogger()
	}

	configDB, err := NewConfigDB(logger, opts)
	if err != nil {
		return nil, fmt.Errorf("error initializing configuration database: %w", err)
	}

	return &Client{
		clients: make(map[data.ClientUID]*internalclient.Client),
		configs: configDB,
		logger:  logger,
	}, nil
}

func (c *Client) Events() chan Event {
	if c.eventsCh == nil {
		// eventsCh needs to be buffer to avoid a deadlock caused by methods accessing the registry in response to its events
		c.eventsCh = make(chan Event, 10)
	}

	return c.eventsCh
}

func (c *Client) getClient(uid data.ClientUID) (*internalclient.Client, error) {
	foundClient, ok := c.clients[uid]
	if !ok {
		return nil, fmt.Errorf("%w, uid: %s", ErrUnkownClient, uid)
	}

	return foundClient, nil
}

func (c *Client) GetClient(uid data.ClientUID) (*internalclient.Client, error) {
	c.Lock()
	defer c.Unlock()

	return c.getClient(uid)
}

func (c *Client) GetSamplerConfig(uid data.SamplerUID, name, resource string) *data.SamplerConfig {
	config := c.configs.Get(uid, name, resource)
	if config == nil {
		// if no config for the specific uid, try to get a more generic config
		config = c.configs.Get("", name, resource)

		// if no config exists, return empty config
		if config == nil {
			config = &data.SamplerConfig{}
		}
	}

	return config
}

func (c *Client) Register(uid data.ClientUID) error {
	c.Lock()
	defer c.Unlock()

	knownClient, err := c.getClient(uid)
	if err != nil && !errors.Is(err, ErrUnkownClient) {
		return err
	} else if err == nil {
		if knownClient.State == internalclient.Registered {
			c.logger.Error("reregistering an already registered client", "client_id", uid)
		}
	}

	client := internalclient.New(uid)
	client.State = internalclient.Registered

	c.clients[uid] = client

	c.sendEvent(&ClientEvent{
		Action: ClientRegistered,
		UID:    uid,
	})

	return nil
}

func (c *Client) Deregister(UID data.ClientUID) error {
	c.Lock()
	defer c.Unlock()

	_, err := c.getClient(UID)
	if errors.Is(err, ErrUnkownClient) {
		c.logger.Debug("deregistering unknown client, nothing to do", "client_uid", UID)

		return nil
	} else if err != nil {
		return err
	}

	delete(c.clients, UID)

	c.sendEvent(&ClientEvent{
		Action: ClientDeregistered,
		UID:    UID,
	})

	return nil
}

func (c *Client) UpdateSamplerConfig(uid data.SamplerUID, name, resource string, update data.SamplerConfigUpdate) error {
	c.Lock()
	defer c.Unlock()
	c.logger.Debug("updating sampler config", "uid", uid, "name", name, "resource", resource, "update", update)

	updatedConfig := c.configs.Get(uid, name, resource)
	if updatedConfig == nil {
		updatedConfig = data.NewSamplerConfig()
	}

	if update.LimiterIn != nil {
		updatedConfig.LimiterIn = update.LimiterIn
	}

	if update.SamplingIn != nil {
		updatedConfig.SamplingIn = update.SamplingIn
	}

	if len(update.StreamUpdates) > 0 {
		updatedConfig.Streams = make(map[data.SamplerStreamUID]data.Stream)
	}
	for _, rule := range update.StreamUpdates {
		switch rule.Op {
		case data.StreamRuleUpsert:
			updatedConfig.Streams[rule.Stream.UID] = rule.Stream
		case data.StreamRuleDelete:
			delete(updatedConfig.Streams, rule.Stream.UID)
		default:
			c.logger.Error(fmt.Sprintf("received unkown sampling rule update operation: %d", rule.Op))
		}
	}

	if update.LimiterOut != nil {
		updatedConfig.LimiterIn = update.LimiterOut
	}

	var action ConfigAction
	if updatedConfig.IsEmpty() {
		action = ConfigDeleted
		if err := c.configs.Delete(uid, name, resource); err != nil {
			c.logger.Error(fmt.Sprintf("Error deleting configuration: %s", err.Error()))
		}
	} else {
		action = ConfigUpdated
		if err := c.configs.Set(uid, name, resource, updatedConfig); err != nil {
			c.logger.Error(fmt.Sprintf("Error setting configuration: %s", err.Error()))

		}
	}

	c.sendEvent(&ConfigEvent{
		Action:          action,
		SamplerName:     name,
		SamplerResource: resource,
		SamplerUID:      uid,
	})

	return nil
}

func (c *Client) DeleteSamplerConfig(uid data.SamplerUID, name, resource string) error {
	c.Lock()
	defer c.Unlock()

	err := c.configs.Delete(uid, name, resource)
	if err != nil {
		c.logger.Error(err.Error())
	}

	c.sendEvent(&ConfigEvent{
		Action:          ConfigDeleted,
		SamplerName:     name,
		SamplerResource: resource,
		SamplerUID:      uid,
	})

	return nil
}

func (c *Client) sendEvent(ev Event) {
	if c.eventsCh != nil {
		c.eventsCh <- ev
	}
}
