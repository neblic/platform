package storage

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

type Disk[K comparable, V any] struct {
	mutex    sync.RWMutex
	fullPath string
}

func NewDisk[K comparable, V any](path string, name string) (*Disk[K, V], error) {
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

func (d *Disk[K, V]) marshalKey(key K) []byte {
	// Marshal key
	bytes, err := json.Marshal(key)
	if err != nil {
		panic(fmt.Sprintf("could not marshal key: %v", err))
	}

	// Keys are stored as base64, encode marshaled key
	encodedBytes := make([]byte, hex.EncodedLen(len(bytes)))
	hex.Encode(encodedBytes, bytes)
	if err != nil {
		panic(fmt.Sprintf("could not decode key from a base64 string: %v", err))
	}

	return encodedBytes
}

func (d *Disk[K, V]) unmarshalKey(keyBytes []byte) K {
	// Keys are stored as base64, decode it
	decodedBytes := make([]byte, hex.DecodedLen(len(keyBytes)))
	_, err := hex.Decode(decodedBytes, keyBytes)
	if err != nil {
		panic(fmt.Sprintf("could not decode key from a base64 string: %v", err))
	}

	// Unmarshal decoded bytes
	var key K
	err = json.Unmarshal(decodedBytes, (*K)(&key))
	if err != nil {
		panic(fmt.Sprintf("could not unmarshal key: %v", err))
	}

	return key
}

func (d *Disk[K, V]) marshalValue(value V) []byte {
	// Marshal value
	bytes, err := json.Marshal(value)
	if err != nil {
		panic(fmt.Sprintf("could not marshal value: %v", err))
	}

	return bytes
}

func (d *Disk[K, V]) unmarshalValue(b []byte) V {
	var value V
	err := json.Unmarshal(b, (*V)(&value))
	if err != nil {
		panic(fmt.Sprintf("could not unmarshal value: %v", err))
	}

	return value
}

func (d *Disk[K, V]) keyPath(key K) string {
	return filepath.Join(d.fullPath, string(d.marshalKey(key)))
}

func (d *Disk[K, V]) Get(key K) (V, error) {
	var emptyValue V

	d.mutex.RLock()
	defer d.mutex.RUnlock()

	data, err := os.ReadFile(d.keyPath(key))
	if err != nil {
		if os.IsNotExist(err) {
			return emptyValue, ErrUnknownKey
		}
		return emptyValue, fmt.Errorf("could not read data from disk: %v", err)
	}

	value := d.unmarshalValue(data)

	return value, nil
}

func (d *Disk[K, V]) Range(fn func(key K, value V)) error {
	entries, err := os.ReadDir(d.fullPath)
	if err != nil {
		return fmt.Errorf("could not read all data from disk: %w", err)
	}

	var multiErr error
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		// Read data from file
		content, err := os.ReadFile(filepath.Join(d.fullPath, entry.Name()))
		if err != nil {
			multiErr = errors.Join(multiErr, err)
		}

		// Decode key
		key := d.unmarshalKey([]byte(entry.Name()))

		// Decode value
		value := d.unmarshalValue(content)

		fn(key, value)
	}

	return multiErr
}

func (d *Disk[K, V]) Set(key K, value V) error {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	keyBytes := d.marshalKey(key)

	// Perform a two step write to avoid populating malformed configuration to disk in case
	// of unexpected stop of the service at mid write. In order to achieve that, configuration
	// is writen to a tmporary file and then moved to the final place.
	tmpFile, err := os.CreateTemp(d.fullPath, string(keyBytes))
	if err != nil {
		return fmt.Errorf("could not create a temporary file: %v", err)
	}

	valueBytes := d.marshalValue(value)
	if _, err := tmpFile.Write(valueBytes); err != nil {
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

func (d *Disk[K, V]) Delete(key K) error {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	err := os.Remove(d.keyPath(key))
	if err != nil {
		if os.IsNotExist(err) {
			return ErrUnknownKey
		}
		return fmt.Errorf("could not delete data from disk: %v", err)
	}

	return nil
}
