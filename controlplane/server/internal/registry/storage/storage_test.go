package storage

import (
	"reflect"
	"testing"
)

func initializeStorage(storage Storage[TestKey, *TestValue]) {
	err := storage.Set(TestKey{String: "key1"}, &TestValue{String: "value1"})
	if err != nil {
		panic(err)
	}
	err = storage.Set(TestKey{String: "key3"}, &TestValue{String: "value3"})
	if err != nil {
		panic(err)
	}
}

type TestKey struct {
	String string
}

func (k TestKey) Hash() string {
	return k.String
}

type TestValue struct {
	String string
}

func StorageGetSuite(t *testing.T, storageProvider func() Storage[TestKey, *TestValue]) {

	type args struct {
		key TestKey
	}
	tests := []struct {
		name    string
		args    args
		want    *TestValue
		wantErr error
	}{
		{
			name: "get value that exists",
			args: args{
				key: TestKey{String: "key1"},
			},
			want:    &TestValue{String: "value1"},
			wantErr: nil,
		},
		{
			name: "get value that does not exists",
			args: args{
				key: TestKey{String: "key2"},
			},
			want:    nil,
			wantErr: ErrUnknownKey,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := storageProvider()
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

func StorageSetSuite(t *testing.T, storageProvider func() Storage[TestKey, *TestValue]) {

	type args struct {
		key   TestKey
		value *TestValue
	}
	tests := []struct {
		name    string
		args    args
		want    *TestValue
		wantErr bool
	}{
		{
			name: "test setting non existing value",
			args: args{
				key:   TestKey{String: "key2"},
				value: &TestValue{String: "value2"},
			},
			want:    &TestValue{String: "value2"},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := storageProvider()
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

func StorageDeleteSuite(t *testing.T, storageProvider func() Storage[TestKey, *TestValue]) {

	type args struct {
		key TestKey
	}
	tests := []struct {
		name    string
		args    args
		wantErr error
	}{
		{
			name: "delete existing key",
			args: args{
				key: TestKey{String: "key1"},
			},
			wantErr: nil,
		},
		{
			name: "delete non existing key",
			args: args{
				key: TestKey{String: "key2"},
			},
			wantErr: ErrUnknownKey,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := storageProvider()
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
