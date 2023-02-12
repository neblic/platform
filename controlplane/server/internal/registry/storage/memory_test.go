package storage

import (
	"testing"
)

func memoryStorageProvider() Storage[TestKey, *TestValue] {
	memoryStorage := NewMemory[TestKey, *TestValue]()
	initializeStorage(memoryStorage)

	return memoryStorage
}

func TestMemory_Get(t *testing.T) {
	StorageGetSuite(t, memoryStorageProvider)
}

func TestMemory_Set(t *testing.T) {
	StorageSetSuite(t, memoryStorageProvider)
}

func TestMemory_Delete(t *testing.T) {
	StorageDeleteSuite(t, memoryStorageProvider)
}
