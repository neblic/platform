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
	clients map[data.ClientUID]*internalclient.Client

	configs *ConfigDB

	logger logging.Logger
	m      sync.Mutex
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

func (c *Client) getClient(uid data.ClientUID) (*internalclient.Client, error) {
	foundClient, ok := c.clients[uid]
	if !ok {
		return nil, fmt.Errorf("%w, uid: %s", ErrUnkownClient, uid)
	}

	return foundClient, nil
}

func (c *Client) GetSamplerConfig(uid data.SamplerUID, name, resource string) *data.SamplerConfig {
	c.m.Lock()
	defer c.m.Unlock()

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
	c.m.Lock()
	defer c.m.Unlock()

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

	return nil
}

func (c *Client) Deregister(UID data.ClientUID) error {
	c.m.Lock()
	defer c.m.Unlock()

	_, err := c.getClient(UID)
	if errors.Is(err, ErrUnkownClient) {
		c.logger.Debug("deregistering unknown client, nothing to do", "client_uid", UID)

		return nil
	} else if err != nil {
		return err
	}

	delete(c.clients, UID)

	return nil
}

func (c *Client) UpdateSamplerConfig(uid data.SamplerUID, name, resource string, update data.SamplerConfigUpdate) error {
	c.m.Lock()
	defer c.m.Unlock()

	updatedConfig := c.configs.Get(uid, name, resource)
	if updatedConfig == nil {
		updatedConfig = data.NewSamplerConfig()
	}

	// LimiterIn
	if update.Reset.LimiterIn {
		updatedConfig.LimiterIn = nil
	}
	if update.LimiterIn != nil {
		updatedConfig.LimiterIn = update.LimiterIn
	}

	// SamplingIn
	if update.Reset.SamplingIn {
		updatedConfig.SamplingIn = nil
	}
	if update.SamplingIn != nil {
		updatedConfig.SamplingIn = update.SamplingIn
	}

	// Streams
	if update.Reset.Streams {
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

	// LimiterOut
	if update.Reset.LimiterOut {
		updatedConfig.LimiterOut = nil
	}
	if update.LimiterOut != nil {
		updatedConfig.LimiterIn = update.LimiterOut
	}

	if err := c.configs.Set(uid, name, resource, updatedConfig); err != nil {
		c.logger.Error(fmt.Sprintf("Error setting configuration: %s", err.Error()))
	}

	return nil
}

func (c *Client) DeleteSamplerConfig(uid data.SamplerUID, name, resource string) error {
	err := c.configs.Delete(uid, name, resource)
	if err != nil {
		c.logger.Error(err.Error())
	}

	return nil
}
