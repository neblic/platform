package storage

import (
	"io/ioutil"
	"log"
	"reflect"
	"testing"
)

type TestData struct {
	Data string
}

func buildDiskStorage() *Disk[TestData] {
	dir, err := ioutil.TempDir("", "storagetest")
	if err != nil {
		log.Fatal(err)
	}
	diskStorage, err := NewDisk[TestData](dir, "test")
	if err != nil {
		panic(err)
	}

	err = diskStorage.Set("key1", TestData{Data: "value1"})
	if err != nil {
		panic(err)
	}
	err = diskStorage.Set("key3", TestData{Data: "value3"})
	if err != nil {
		panic(err)
	}

	return diskStorage
}

func TestDisk_Get(t *testing.T) {

	type args struct {
		key string
	}
	tests := []struct {
		name    string
		args    args
		want    TestData
		wantErr error
	}{
		{
			name: "get value that exists",
			args: args{
				key: "key1",
			},
			want:    TestData{Data: "value1"},
			wantErr: nil,
		},
		{
			name: "get value that does not exists",
			args: args{
				key: "key2",
			},
			want:    TestData{},
			wantErr: ErrUnknownKey,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := buildDiskStorage()
			got, err := s.Get(tt.args.key)
			if err != tt.wantErr {
				t.Errorf("Storage.Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Storage.Get() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDisk_Set(t *testing.T) {

	type args struct {
		key   string
		value TestData
	}
	tests := []struct {
		name    string
		args    args
		want    TestData
		wantErr bool
	}{
		{
			name: "test setting non existing value",
			args: args{
				key:   "key2",
				value: TestData{Data: "value2"},
			},
			want:    TestData{Data: "value2"},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := buildDiskStorage()
			err := s.Set(tt.args.key, tt.args.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("Storage.Set() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			value, err := s.Get(tt.args.key)
			if err != nil {
				t.Errorf("Storage.Set() error validating set data = %v", err)
			}
			if !reflect.DeepEqual(value, tt.want) {
				t.Errorf("Storage.Set() set data did not produce the expected result, got %v, want %v", value, tt.want)
			}
		})
	}
}

func TestDisk_Delete(t *testing.T) {

	type args struct {
		key string
	}
	tests := []struct {
		name    string
		args    args
		wantErr error
	}{
		{
			name: "delete existing key",
			args: args{
				key: "key1",
			},
			wantErr: nil,
		},
		{
			name: "delete non existing key",
			args: args{
				key: "key2",
			},
			wantErr: ErrUnknownKey,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := buildDiskStorage()
			err := s.Delete(tt.args.key)
			if err != tt.wantErr {
				t.Errorf("Storage.Delete() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			_, err = s.Get(tt.args.key)
			if err != ErrUnknownKey {
				t.Errorf("Storage.Delete() element not deleted")
			}
		})
	}
}

func TestDisk_Range(t *testing.T) {

	tests := []struct {
		name    string
		want    map[string]TestData
		wantErr bool
	}{
		{
			name:    "test ranging 2 values",
			want:    map[string]TestData{"key1": {Data: "value1"}, "key3": {Data: "value3"}},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create callback function
			got := map[string]TestData{}
			callback := func(key string, value TestData) {
				got[key] = value
			}

			s := buildDiskStorage()
			err := s.Range(callback)
			if (err != nil) != tt.wantErr {
				t.Errorf("Storage.Range() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Storage.Range() got %v, want %v", got, tt.want)
			}

		})
	}
}
