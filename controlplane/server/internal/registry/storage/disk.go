package storage

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
)

type Disk[T any] struct {
	mutex    sync.RWMutex
	fullPath string
}

func NewDisk[T any](path string, name string) (*Disk[T], error) {
	fullPath := filepath.Join(path, name)

	err := os.Mkdir(fullPath, os.ModePerm)
	if err != nil {
		return nil, fmt.Errorf("could not create storage directory: %v", err)
	}

	return &Disk[T]{
		mutex:    sync.RWMutex{},
		fullPath: fullPath,
	}, nil
}

func (d *Disk[T]) keyPath(key string) string {
	return filepath.Join(d.fullPath, key)
}

func (d *Disk[T]) Get(key string) (T, error) {
	var value T

	d.mutex.RLock()
	defer d.mutex.RUnlock()

	data, err := os.ReadFile(d.keyPath(key))
	if err != nil {
		if os.IsNotExist(err) {
			return value, ErrUnknownKey
		}
		return value, fmt.Errorf("could not read data from disk: %v", err)
	}

	err = json.Unmarshal(data, (*T)(&value))
	if err != nil {
		return value, fmt.Errorf("could not deserialize data: %v", err)
	}

	return value, nil
}

func (d *Disk[T]) Set(key string, value T) error {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("could not serialize data: %v", err)
	}

	// Perform a two step write to avoid populating malformed configuration to disk in case
	// of unexpected stop of the service at mid write. In order to achieve that, configuration
	// is writen to a tmporary file and then moved to the final place.
	tmpFile, err := os.CreateTemp(d.fullPath, key)
	if err != nil {
		return fmt.Errorf("could not create a temporary file: %v", err)
	}

	if _, err := tmpFile.Write(data); err != nil {
		return fmt.Errorf("could not write configuration to a temporary file: %v", err)
	}

	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("could not close the temporary file: %v", err)
	}

	if err := os.Rename(tmpFile.Name(), d.keyPath(key)); err != nil {
		return fmt.Errorf("could not rename the temporary file: %v", err)
	}

	return nil
}

func (d *Disk[T]) Delete(key string) error {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	err := os.Remove(d.keyPath(key))
	if err != nil {
		if os.IsNotExist(err) {
			return ErrUnknownKey
		}
		return fmt.Errorf("could not delete data from disk: %v", err)
	}

	return err
}

func (d *Disk[T]) Range(callback func(key string, value T)) error {
	d.mutex.RLock()
	defer d.mutex.RUnlock()

	files, err := ioutil.ReadDir(d.fullPath)
	if err != nil {
		return fmt.Errorf("could not read all the entries from disk: %v", err)
	}

	for _, f := range files {
		key := f.Name()
		value, err := d.Get(key)
		if err != nil {
			return err
		}

		callback(key, value)
	}

	return nil
}
