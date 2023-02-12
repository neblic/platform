package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

type Disk[K Hasher, V any] struct {
	mutex    sync.RWMutex
	fullPath string
}

func NewDisk[K Hasher, V any](path string, name string) (*Disk[K, V], error) {
	fullPath := filepath.Join(path, name)

	err := os.MkdirAll(fullPath, os.ModePerm)
	if err != nil {
		return nil, fmt.Errorf("could not create storage directory: %v", err)
	}

	return &Disk[K, V]{
		mutex:    sync.RWMutex{},
		fullPath: fullPath,
	}, nil
}

func (d *Disk[K, V]) keyPath(key string) string {
	return filepath.Join(d.fullPath, key)
}

func (d *Disk[K, V]) Get(key K) (V, error) {
	var value V

	d.mutex.RLock()
	defer d.mutex.RUnlock()

	data, err := os.ReadFile(d.keyPath(key.Hash()))
	if err != nil {
		if os.IsNotExist(err) {
			return value, ErrUnknownKey
		}
		return value, fmt.Errorf("could not read data from disk: %v", err)
	}

	err = json.Unmarshal(data, (*V)(&value))
	if err != nil {
		return value, fmt.Errorf("could not deserialize data: %v", err)
	}

	return value, nil
}

func (d *Disk[K, V]) Set(key K, value V) error {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("could not serialize data: %v", err)
	}

	keyHash := key.Hash()

	// Perform a two step write to avoid populating malformed configuration to disk in case
	// of unexpected stop of the service at mid write. In order to achieve that, configuration
	// is writen to a tmporary file and then moved to the final place.
	tmpFile, err := os.CreateTemp(d.fullPath, keyHash)
	if err != nil {
		return fmt.Errorf("could not create a temporary file: %v", err)
	}

	if _, err := tmpFile.Write(data); err != nil {
		return fmt.Errorf("could not write configuration to a temporary file: %v", err)
	}

	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("could not close the temporary file: %v", err)
	}

	if err := os.Rename(tmpFile.Name(), d.keyPath(keyHash)); err != nil {
		return fmt.Errorf("could not rename the temporary file: %v", err)
	}

	return nil
}

func (d *Disk[K, V]) Delete(key K) error {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	err := os.Remove(d.keyPath(key.Hash()))
	if err != nil {
		if os.IsNotExist(err) {
			return ErrUnknownKey
		}
		return fmt.Errorf("could not delete data from disk: %v", err)
	}

	return nil
}
