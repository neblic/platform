package registry

import (
	"reflect"
	"sync"
	"testing"

	"github.com/neblic/platform/controlplane/data"
	"github.com/neblic/platform/controlplane/server/internal/registry/storage"
	"github.com/neblic/platform/logging"
)

func storageExample() storage.Storage[samplerIdentifier, *identifiedSamplerConfig] {
	storage := storage.NewMemory[samplerIdentifier, *identifiedSamplerConfig]()

	samplerIdentifier1 := samplerIdentifier{Resource: "resource1", Name: "name1", UID: "uid1"}
	storage.Set(samplerIdentifier1, &identifiedSamplerConfig{
		samplerIdentifier: samplerIdentifier1,
		SamplerConfig: &data.SamplerConfig{
			Streams: map[data.SamplerStreamUID]data.Stream{
				"stream1": {
					UID: "stream1",
					StreamRule: data.StreamRule{
						UID:  "streamRule1",
						Lang: data.SrlCel,
						Rule: "true",
					},
				},
			},
		},
	})

	samplerIdentifier2 := samplerIdentifier{Resource: "resource2", Name: "name1", UID: "uid1"}
	storage.Set(samplerIdentifier2, &identifiedSamplerConfig{
		samplerIdentifier: samplerIdentifier2,
		SamplerConfig: &data.SamplerConfig{
			Streams: map[data.SamplerStreamUID]data.Stream{},
		},
	})

	samplerIdentifier3 := samplerIdentifier{Resource: "resource1", Name: "name2", UID: "uid1"}
	storage.Set(samplerIdentifier3, &identifiedSamplerConfig{
		samplerIdentifier: samplerIdentifier3,
		SamplerConfig: &data.SamplerConfig{
			Streams: map[data.SamplerStreamUID]data.Stream{},
		},
	})

	samplerIdentifier4 := samplerIdentifier{Resource: "resource1", Name: "name1", UID: "uid2"}
	storage.Set(samplerIdentifier4, &identifiedSamplerConfig{
		samplerIdentifier: samplerIdentifier4,
		SamplerConfig: &data.SamplerConfig{
			Streams: map[data.SamplerStreamUID]data.Stream{},
		},
	})

	return storage
}

func loggerExample() logging.Logger {
	logger, _ := logging.NewZapDev()
	return logger
}

func TestConfigDB_Get(t *testing.T) {
	type fields struct {
		storage storage.Storage[samplerIdentifier, *identifiedSamplerConfig]
		logger  logging.Logger
		mutex   *sync.RWMutex
	}
	type args struct {
		samplerUID      data.SamplerUID
		samplerName     string
		samplerResource string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *data.SamplerConfig
	}{
		{
			name: "get valid entry",
			fields: fields{
				storage: storageExample(),
				logger:  loggerExample(),
				mutex:   new(sync.RWMutex),
			},
			args: args{
				samplerUID:      "uid1",
				samplerName:     "name1",
				samplerResource: "resource1",
			},
			want: &data.SamplerConfig{
				Streams: map[data.SamplerStreamUID]data.Stream{
					"stream1": {
						UID: "stream1",
						StreamRule: data.StreamRule{
							UID:  "streamRule1",
							Lang: data.SrlCel,
							Rule: "true",
						},
					},
				},
			},
		},
		{
			name: "get invalid entry",
			fields: fields{
				storage: storageExample(),
				logger:  loggerExample(),
				mutex:   new(sync.RWMutex),
			},
			args: args{
				samplerUID:      "uid1000",
				samplerName:     "name1",
				samplerResource: "resource1",
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &ConfigDB{
				storage: tt.fields.storage,
				logger:  tt.fields.logger,
				mutex:   tt.fields.mutex,
			}
			if got := s.Get(tt.args.samplerUID, tt.args.samplerName, tt.args.samplerResource); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ConfigDB.Get() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfigDB_Set(t *testing.T) {
	type fields struct {
		storage storage.Storage[samplerIdentifier, *identifiedSamplerConfig]
		logger  logging.Logger
		mutex   *sync.RWMutex
	}
	type args struct {
		samplerUID      data.SamplerUID
		samplerName     string
		samplerResource string
		config          *data.SamplerConfig
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "Set nonexisting entry with valid config",
			fields: fields{
				storage: storageExample(),
				logger:  loggerExample(),
				mutex:   new(sync.RWMutex),
			},
			args: args{
				samplerUID:      "uid1000",
				samplerName:     "name1",
				samplerResource: "resource1",
				config: &data.SamplerConfig{
					Streams: map[data.SamplerStreamUID]data.Stream{
						"stream2": {
							UID: "stream2",
							StreamRule: data.StreamRule{
								UID:  "streamRule2",
								Lang: data.SrlCel,
								Rule: "true",
							},
						},
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &ConfigDB{
				storage: tt.fields.storage,
				logger:  tt.fields.logger,
				mutex:   tt.fields.mutex,
			}
			if err := s.Set(tt.args.samplerUID, tt.args.samplerName, tt.args.samplerResource, tt.args.config); (err != nil) != tt.wantErr {
				t.Errorf("ConfigDB.Set() error = %v, wantErr %v", err, tt.wantErr)
			}

			// Check internal state
			identifiedSamplerConfig := newIdentifiedSamplerConfig(tt.args.samplerUID, tt.args.samplerName, tt.args.samplerResource, tt.args.config)
			samplerIdentifier := identifiedSamplerConfig.samplerIdentifier
			value, err := s.storage.Get(samplerIdentifier)
			if err != nil {
				t.Errorf("ConfigDB.Set() storage does not contain the data index")
			}

			if !reflect.DeepEqual(value.SamplerConfig, tt.args.config) {
				t.Errorf("ConfigDB.Set() invalid storage data = %v, want %v", value.SamplerConfig, tt.args.config)
			}
		})
	}
}

func TestConfigDB_Delete(t *testing.T) {
	type fields struct {
		storage storage.Storage[samplerIdentifier, *identifiedSamplerConfig]
		logger  logging.Logger
		mutex   *sync.RWMutex
	}
	type args struct {
		samplerUID      data.SamplerUID
		samplerName     string
		samplerResource string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "delete nonexisting entry",
			fields: fields{
				storage: storageExample(),
				logger:  loggerExample(),
				mutex:   new(sync.RWMutex),
			},
			args: args{
				samplerUID:      "uid1000",
				samplerName:     "name1",
				samplerResource: "resource1",
			},
			wantErr: true,
		},
		{
			name: "delete existing entry",
			fields: fields{
				storage: storageExample(),
				logger:  loggerExample(),
				mutex:   new(sync.RWMutex),
			},
			args: args{
				samplerUID:      "uid1",
				samplerName:     "name1",
				samplerResource: "resource1",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &ConfigDB{
				storage: tt.fields.storage,
				logger:  tt.fields.logger,
				mutex:   tt.fields.mutex,
			}
			if err := s.Delete(tt.args.samplerUID, tt.args.samplerName, tt.args.samplerResource); (err != nil) != tt.wantErr {
				t.Errorf("ConfigDB.Delete() error = %v, wantErr %v", err, tt.wantErr)
			}

			// Check internal state
			identifiedSamplerConfig := newIdentifiedSamplerConfig(tt.args.samplerUID, tt.args.samplerName, tt.args.samplerResource, nil)
			samplerIdentifier := identifiedSamplerConfig.samplerIdentifier
			_, err := s.storage.Get(samplerIdentifier)
			if err != storage.ErrUnknownKey {
				t.Errorf("ConfigDB.Delete() storage does contain the deleted data index or an unexpected error happened")
			}
		})
	}
}
