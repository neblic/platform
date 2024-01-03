package storage

import (
	"reflect"
	"testing"

	"github.com/neblic/platform/controlplane/control"
	"github.com/neblic/platform/controlplane/server/internal/defs"
)

func newSamplerConfig(stream control.SamplerStreamUID) control.SamplerConfig {
	return control.SamplerConfig{
		Streams: map[control.SamplerStreamUID]control.Stream{stream: {UID: stream}},
		LimiterIn: &control.LimiterConfig{
			Limit: 0,
		},
		SamplingIn: &control.SamplingConfig{
			SamplingType: 0,
			DeterministicSampling: control.DeterministicSamplingConfig{
				SampleRate:             0,
				SampleEmptyDeterminant: false,
			},
		},
		LimiterOut: &control.LimiterConfig{
			Limit: 0,
		},
		Digests: map[control.SamplerDigestUID]control.Digest{},
		Events:  map[control.SamplerEventUID]control.Event{},
	}
}

func initializeStorage(storage Storage) error {
	err := storage.SetSampler("resource1", "sampler1", newSamplerConfig("stream1"))
	if err != nil {
		return err
	}
	err = storage.SetSampler("resource3", "sampler3", newSamplerConfig("stream3"))
	if err != nil {
		return err
	}

	return nil
}

func StorageGetSamplerSuite(t *testing.T, storageProvider func() Storage) {

	type args struct {
		resource string
		sampler  string
	}
	tests := []struct {
		name    string
		args    args
		want    control.SamplerConfig
		wantErr error
	}{
		{
			name: "get value that exists",
			args: args{
				resource: "resource1",
				sampler:  "sampler1",
			},
			want:    newSamplerConfig("stream1"),
			wantErr: nil,
		},
		{
			name: "get value that does not exists",
			args: args{
				resource: "resource2",
				sampler:  "sampler2",
			},
			want:    control.SamplerConfig{},
			wantErr: ErrUnknownSampler,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := storageProvider()
			err := initializeStorage(s)
			if err != nil {
				t.Errorf("Storage.GetSampler() error initializing storage = %v", err)
				return
			}

			got, err := s.GetSampler(tt.args.resource, tt.args.sampler)
			if err != tt.wantErr {
				t.Errorf("Storage.GetSampler() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Storage.GetSampler() = %v, want %v", got, tt.want)
			}
		})
	}
}

func StorageRangeSamplersSuite(t *testing.T, storageProvider func() Storage) {

	tests := []struct {
		name    string
		want    map[defs.SamplerIdentifier]control.SamplerConfig
		wantErr error
	}{
		{
			name: "successful range",
			want: map[defs.SamplerIdentifier]control.SamplerConfig{
				{Resource: "resource1", Name: "sampler1"}: newSamplerConfig("stream1"),
				{Resource: "resource3", Name: "sampler3"}: newSamplerConfig("stream3"),
			},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			s := storageProvider()
			got := map[defs.SamplerIdentifier]control.SamplerConfig{}
			err := s.RangeSamplers(func(resource string, sampler string, config control.SamplerConfig) {
				got[defs.NewSamplerIdentifier(resource, sampler)] = config
			})
			if err != tt.wantErr {
				t.Errorf("Storage.RangeSamplers() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Storage.RangeSamplers() = %v, want %v", got, tt.want)
			}
		})
	}
}

func StorageSetSamplerSuite(t *testing.T, storageProvider func() Storage) {

	type args struct {
		resource string
		sampler  string
		config   control.SamplerConfig
	}
	tests := []struct {
		name    string
		args    args
		want    control.SamplerConfig
		wantErr bool
	}{
		{
			name: "test setting non existing value",
			args: args{
				resource: "resource2",
				sampler:  "sampler2",
				config:   newSamplerConfig("stream2"),
			},
			want:    newSamplerConfig("stream2"),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := storageProvider()
			err := s.SetSampler(tt.args.resource, tt.args.sampler, tt.args.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("Storage.SetSampler() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			value, err := s.GetSampler(tt.args.resource, tt.args.sampler)
			if err != nil {
				t.Errorf("Storage.SetSampler() error validating set data = %v", err)
			}
			if !reflect.DeepEqual(value, tt.want) {
				t.Errorf("Storage.SetSampler() set data did not produce the expected result, got %v, want %v", value, tt.want)
			}
		})
	}
}

func StorageDeleteSamplerSuite(t *testing.T, storageProvider func() Storage) {

	type args struct {
		resource string
		sampler  string
	}
	tests := []struct {
		name    string
		args    args
		wantErr error
	}{
		{
			name: "delete existing key",
			args: args{
				resource: "resource1",
				sampler:  "sampler1",
			},
			wantErr: nil,
		},
		{
			name: "delete non existing key",
			args: args{
				resource: "resource2",
				sampler:  "sampler2",
			},
			wantErr: ErrUnknownSampler,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := storageProvider()
			err := s.DeleteSampler(tt.args.resource, tt.args.sampler)
			if err != tt.wantErr {
				t.Errorf("Storage.Delete() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			_, err = s.GetSampler(tt.args.resource, tt.args.sampler)
			if err != ErrUnknownSampler {
				t.Errorf("Storage.Delete() element not deleted")
			}
		})
	}
}
