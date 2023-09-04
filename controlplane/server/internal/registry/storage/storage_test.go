package storage

import (
	"reflect"
	"testing"
)

func initializeStorage(storage Storage[TestKey, *TestValue]) {
	err := storage.Set(TestKey{Key: "key1"}, &TestValue{Value: "value1"})
	if err != nil {
		panic(err)
	}
	err = storage.Set(TestKey{Key: "key3"}, &TestValue{Value: "value3"})
	if err != nil {
		panic(err)
	}
}

type TestKey struct {
	Key string
}

func (k TestKey) ToString() string {
	return k.Key
}

func (k *TestKey) FromString(str string) {
	k.Key = str
}

type TestValue struct {
	Value string
}

func (k TestValue) ToString() string {
	return k.Value
}

func (k *TestValue) FromString(str string) {
	k.Value = str
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
				key: TestKey{Key: "key1"},
			},
			want:    &TestValue{Value: "value1"},
			wantErr: nil,
		},
		{
			name: "get value that does not exists",
			args: args{
				key: TestKey{Key: "key2"},
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

func StorageRangeSuite(t *testing.T, storageProvider func() Storage[TestKey, *TestValue]) {

	tests := []struct {
		name    string
		want    map[TestKey]*TestValue
		wantErr error
	}{
		{
			name: "successful range",
			want: map[TestKey]*TestValue{
				{Key: "key1"}: {Value: "value1"},
				{Key: "key3"}: {Value: "value3"},
			},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			s := storageProvider()
			got := map[TestKey]*TestValue{}
			err := s.Range(func(key TestKey, value *TestValue) {
				got[key] = value
			})
			if err != tt.wantErr {
				t.Errorf("Storage.Range() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Storage.Range() = %v, want %v", got, tt.want)
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
				key:   TestKey{Key: "key2"},
				value: &TestValue{Value: "value2"},
			},
			want:    &TestValue{Value: "value2"},
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
				key: TestKey{Key: "key1"},
			},
			wantErr: nil,
		},
		{
			name: "delete non existing key",
			args: args{
				key: TestKey{Key: "key2"},
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
