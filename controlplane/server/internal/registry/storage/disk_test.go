package storage

import (
	"io/ioutil"
	"log"
	"testing"
)

func diskStorageProvider() Storage[TestKey, *TestValue] {
	dir, err := ioutil.TempDir("", "storagetest")
	if err != nil {
		log.Fatal(err)
	}
	diskStorage, err := NewDisk[TestKey, *TestValue](dir, "test")
	if err != nil {
		panic(err)
	}

	initializeStorage(diskStorage)

	return diskStorage
}

func TestDisk_Get(t *testing.T) {
	StorageGetSuite(t, diskStorageProvider)
}

func TestDisk_Set(t *testing.T) {
	StorageSetSuite(t, diskStorageProvider)
}

func TestDisk_Delete(t *testing.T) {
	StorageDeleteSuite(t, diskStorageProvider)
}
