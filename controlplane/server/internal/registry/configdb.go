package registry

import (
	"fmt"
	"strings"
	"sync"

	"github.com/neblic/platform/controlplane/data"
	"github.com/neblic/platform/controlplane/server/internal/registry/storage"
	"github.com/neblic/platform/logging"
)

type storeKey struct {
	samplerName     string
	samplerResource string
}

type storeValue struct {
	config map[data.SamplerUID]*data.SamplerConfig
}

type ConfigDB struct {
	db      map[storeKey]*storeValue
	storage storage.Storage[*data.SamplerConfig]

	logger logging.Logger
	sync.Mutex
}

func NewConfigDB(store storage.Storage[*data.SamplerConfig], logger logging.Logger) (*ConfigDB, error) {
	c := &ConfigDB{
		db:      make(map[storeKey]*storeValue),
		storage: store,

		logger: logger,
	}

	c.loadFromStorage()

	return c, nil
}

func (s *ConfigDB) storageKey(samplerUID data.SamplerUID, samplerName, samplerResource string) string {
	return fmt.Sprintf("%s#%s#%s", samplerName, samplerResource, samplerUID)
}

func (s *ConfigDB) parseKey(key string) (data.SamplerUID, string, string, error) {
	parts := strings.Split(key, "#")
	if len(parts) != 3 {
		return "", "", "", fmt.Errorf("couldn't parse configuration key: %s", key)
	}

	return data.SamplerUID(parts[0]), parts[1], parts[2], nil
}

func (s *ConfigDB) loadFromStorage() {
	s.storage.Range(func(key string, config *data.SamplerConfig) {
		uid, name, resource, err := s.parseKey(key)
		if err != nil {
			s.logger.Error(fmt.Sprintf("Error loading configuration from storage: %s", err))
		} else {
			s.set(uid, name, resource, config)
		}
	})
}

func (s *ConfigDB) Get(samplerUID data.SamplerUID, samplerName, samplerResource string) *data.SamplerConfig {
	s.Lock()
	defer s.Unlock()

	key := storeKey{samplerName, samplerResource}

	value, ok := s.db[key]
	if !ok {
		return nil
	}

	config, ok := value.config[samplerUID]
	if !ok {
		return nil
	}

	return config
}

func (s *ConfigDB) persist(samplerUID data.SamplerUID, samplerName, samplerResource string, config *data.SamplerConfig) error {
	if config != nil {
		if err := s.storage.Set(s.storageKey(samplerUID, samplerName, samplerResource), config); err != nil {
			return fmt.Errorf("error persisting config to disk: %w", err)
		}
	} else {
		if err := s.storage.Delete(s.storageKey(samplerUID, samplerName, samplerResource)); err != nil {
			return fmt.Errorf("error persisting config to disk: %w", err)
		}
	}

	return nil
}

func (s *ConfigDB) set(samplerUID data.SamplerUID, samplerName, samplerResource string, config *data.SamplerConfig) {
	key := storeKey{samplerName, samplerResource}
	value, ok := s.db[key]
	if !ok {
		value = &storeValue{
			config: make(map[data.SamplerUID]*data.SamplerConfig),
		}

		s.db[key] = value
	}

	value.config[samplerUID] = config
}

func (s *ConfigDB) Set(samplerUID data.SamplerUID, samplerName, samplerResource string, config *data.SamplerConfig) error {
	s.Lock()
	defer s.Unlock()

	if err := s.persist(samplerUID, samplerName, samplerResource, config); err != nil {
		return err
	}

	s.set(samplerUID, samplerName, samplerResource, config)

	return nil
}

func (s *ConfigDB) Delete(samplerUID data.SamplerUID, samplerName, samplerResource string) error {
	s.Lock()
	defer s.Unlock()

	key := storeKey{samplerName, samplerResource}

	value, ok := s.db[key]
	if !ok {
		return nil
	}

	if err := s.persist(samplerUID, samplerName, samplerResource, nil); err != nil {
		return err
	}

	delete(value.config, samplerUID)

	if len(value.config) == 0 {
		delete(s.db, key)
	}

	return nil
}
