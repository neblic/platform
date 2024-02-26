package storage

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"sync"

	"github.com/google/renameio/v2"
	"gopkg.in/yaml.v3"
)

type ConfigDocument struct {
	Samplers []SamplerEntry
}

type Disk struct {
	mutex sync.RWMutex
	path  string
}

func NewDisk(path string) (*Disk, error) {

	err := os.MkdirAll(filepath.Dir(path), os.ModePerm)
	if err != nil {
		return nil, fmt.Errorf("could not create disk storage directory: %v", err)
	}

	_, err = os.Stat(path)
	if errors.Is(err, os.ErrNotExist) {
		err := writeConfigDocument(path, &ConfigDocument{})
		if err != nil {
			return nil, err
		}
	}

	return &Disk{
		mutex: sync.RWMutex{},
		path:  path,
	}, nil
}

func readConfigDocument(path string) (*ConfigDocument, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("could not read configuration from disk: %v", err)
	}

	configDocument := &ConfigDocument{}
	err = yaml.Unmarshal(data, configDocument)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal configuration: %v", err)
	}

	return configDocument, nil
}

func writeConfigDocument(path string, configDocument *ConfigDocument) error {
	data, err := yaml.Marshal(configDocument)
	if err != nil {
		return fmt.Errorf("could not marshal configuration: %v", err)
	}

	err = renameio.WriteFile(path, data, 0666)
	if err != nil {
		return fmt.Errorf("could not write configuration to disk: %v", err)
	}

	return nil
}

func findSampler(samplers []SamplerEntry, resource string, sampler string) int {
	return slices.IndexFunc(samplers, func(entry SamplerEntry) bool {
		return entry.Resource == resource && entry.Name == sampler
	})
}

func (d *Disk) RangeSamplers(fn func(entry SamplerEntry)) error {
	d.mutex.RLock()
	defer d.mutex.RUnlock()

	// Read data
	configDocument, err := readConfigDocument(d.path)
	if err != nil {
		return err
	}

	// Range samplers
	for _, sampler := range configDocument.Samplers {
		fn(sampler)
	}

	return nil
}

func (d *Disk) GetSampler(resource string, sampler string) (SamplerEntry, error) {
	d.mutex.RLock()
	defer d.mutex.RUnlock()

	// Read data
	configDocument, err := readConfigDocument(d.path)
	if err != nil {
		return SamplerEntry{}, err
	}

	// Find sampler
	index := findSampler(configDocument.Samplers, resource, sampler)
	if index == -1 {
		return SamplerEntry{}, ErrUnknownSampler
	}

	return configDocument.Samplers[index], nil
}

func (d *Disk) SetSampler(entry SamplerEntry) error {
	d.mutex.RLock()
	defer d.mutex.RUnlock()

	// Read data
	configDocument, err := readConfigDocument(d.path)
	if err != nil {
		return err
	}

	// Find sampler and replace config
	index := findSampler(configDocument.Samplers, entry.Resource, entry.Name)
	if index == -1 {
		configDocument.Samplers = append(configDocument.Samplers, entry)
	} else {
		configDocument.Samplers[index].Config = entry.Config
		configDocument.Samplers[index].Capabilities = entry.Capabilities
	}

	// Write data
	err = writeConfigDocument(d.path, configDocument)

	return err
}

func (d *Disk) DeleteSampler(resource string, sampler string) error {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	// Read data
	configDocument, err := readConfigDocument(d.path)
	if err != nil {
		return err
	}

	// Find sampler and replace config
	index := findSampler(configDocument.Samplers, resource, sampler)
	if index == -1 {
		return ErrUnknownSampler
	}

	// Delete entry
	configDocument.Samplers = append(configDocument.Samplers[:index], configDocument.Samplers[index+1:]...)

	// Write data
	err = writeConfigDocument(d.path, configDocument)

	return err
}
