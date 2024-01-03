package storage

import (
	"log"
	"os"
	"testing"
)

func diskStorageProvider() Storage {
	file, err := os.CreateTemp("", "storagetest")
	if err != nil {
		log.Fatal(err)
	}

	diskStorage, err := NewDisk(file.Name())
	if err != nil {
		panic(err)
	}

	initializeStorage(diskStorage)

	return diskStorage
}

func TestDisk_Get(t *testing.T) {
	StorageGetSamplerSuite(t, diskStorageProvider)
}

func TestDisk_Range(t *testing.T) {
	StorageRangeSamplersSuite(t, diskStorageProvider)
}

func TestDisk_Set(t *testing.T) {
	StorageSetSamplerSuite(t, diskStorageProvider)
}

func TestDisk_Delete(t *testing.T) {
	StorageDeleteSamplerSuite(t, diskStorageProvider)
}
