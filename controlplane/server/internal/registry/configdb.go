package registry

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"sync"

	"github.com/neblic/platform/controlplane/data"
	"github.com/neblic/platform/controlplane/server/internal/registry/storage"
	"github.com/neblic/platform/logging"
)

type samplerIdentifier struct {
	Resource string
	Name     string
	UID      string
}

func (si samplerIdentifier) Hash() string {
	hasher := sha1.New()
	hasher.Write([]byte(si.Resource))
	hasher.Write([]byte(si.Name))
	hasher.Write([]byte(si.UID))
	hash := string(hex.EncodeToString(hasher.Sum(nil)))
	return hash
}

type identifiedSamplerConfig struct {
	samplerIdentifier
	*data.SamplerConfig
}

func newIdentifiedSamplerConfig(samplerUID data.SamplerUID, samplerName, samplerResource string, config *data.SamplerConfig) *identifiedSamplerConfig {
	samplerIdentifier := samplerIdentifier{
		Resource: samplerResource,
		Name:     samplerName,
		UID:      string(samplerUID),
	}

	identifiedSamplerConfig := &identifiedSamplerConfig{
		samplerIdentifier: samplerIdentifier,
		SamplerConfig:     config,
	}

	return identifiedSamplerConfig
}

type ConfigDB struct {
	storage storage.Storage[samplerIdentifier, *identifiedSamplerConfig]

	logger logging.Logger
	mutex  *sync.RWMutex
}

func NewConfigDB(logger logging.Logger, opts *Options) (*ConfigDB, error) {
	// Initialize storage
	var storageInstance storage.Storage[samplerIdentifier, *identifiedSamplerConfig]
	switch opts.StorageType {
	case MemoryStorage:
		storageInstance = storage.NewMemory[samplerIdentifier, *identifiedSamplerConfig]()
	case DiskStorage:
		var err error
		storageInstance, err = storage.NewDisk[samplerIdentifier, *identifiedSamplerConfig](opts.Path, "config")
		if err != nil {
			return nil, fmt.Errorf("error initializing client disk storage: %v", err)
		}
	}

	c := &ConfigDB{
		storage: storageInstance,

		logger: logger,
		mutex:  new(sync.RWMutex),
	}

	return c, nil
}

func (s *ConfigDB) Get(samplerUID data.SamplerUID, samplerName, samplerResource string) *data.SamplerConfig {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	samplerIdentifier := samplerIdentifier{
		Resource: samplerResource,
		Name:     samplerName,
		UID:      string(samplerUID),
	}

	config, err := s.storage.Get(samplerIdentifier)
	if err != nil {
		if err != storage.ErrUnknownKey {
			s.logger.Error(err.Error())
		}
		return nil
	}

	return config.SamplerConfig
}

func (s *ConfigDB) Set(samplerUID data.SamplerUID, samplerName, samplerResource string, config *data.SamplerConfig) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	identifiedSamplerConfig := newIdentifiedSamplerConfig(samplerUID, samplerName, samplerResource, config)
	samplerIdentifier := identifiedSamplerConfig.samplerIdentifier

	// Set config to the storage
	err := s.storage.Set(samplerIdentifier, identifiedSamplerConfig)

	return err
}

func (s *ConfigDB) Delete(samplerUID data.SamplerUID, samplerName, samplerResource string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	identifiedSamplerConfig := newIdentifiedSamplerConfig(samplerUID, samplerName, samplerResource, nil)
	samplerIdentifier := identifiedSamplerConfig.samplerIdentifier

	// Remove config from storage
	err := s.storage.Delete(samplerIdentifier)

	return err
}
