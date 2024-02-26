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

func newSamplerCapabilities() control.Capabilities {
	return control.Capabilities{
		Stream: control.StreamCapabilities{
			Enabled: true,
		},
		LimiterIn: control.LimiterCapabilities{
			Enabled: false,
		},
		SamplingIn: control.SamplingCapabilities{
			Enabled: false,
			Types:   []control.SamplingType{},
		},
		LimiterOut: control.LimiterCapabilities{
			Enabled: false,
		},
		Digest: control.DigestCapabilities{
			Enabled: false,
			Types:   []control.DigestType{},
		},
	}
}

func initializeStorage(storage Storage) error {
	err := storage.SetSampler(SamplerEntry{
		Resource:     "resource1",
		Name:         "sampler1",
		Config:       newSamplerConfig("stream1"),
		Capabilities: newSamplerCapabilities(),
	})
	if err != nil {
		return err
	}
	err = storage.SetSampler(SamplerEntry{
		Resource:     "resource3",
		Name:         "sampler3",
		Config:       newSamplerConfig("stream3"),
		Capabilities: newSamplerCapabilities(),
	})
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
		want    SamplerEntry
		wantErr error
	}{
		{
			name: "get value that exists",
			args: args{
				resource: "resource1",
				sampler:  "sampler1",
			},
			want: SamplerEntry{
				Resource:     "resource1",
				Name:         "sampler1",
				Config:       newSamplerConfig("stream1"),
				Capabilities: newSamplerCapabilities(),
			},
			wantErr: nil,
		},
		{
			name: "get value that does not exists",
			args: args{
				resource: "resource2",
				sampler:  "sampler2",
			},
			want:    SamplerEntry{},
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
			if !reflect.DeepEqual(got.Capabilities, tt.want.Capabilities) {
				t.Errorf("Storage.GetSampler() = %v, want %v", got.Capabilities, tt.want.Capabilities)
			}
		})
	}
}

func StorageRangeSamplersSuite(t *testing.T, storageProvider func() Storage) {

	tests := []struct {
		name    string
		want    map[defs.SamplerIdentifier]SamplerEntry
		wantErr error
	}{
		{
			name: "successful range",
			want: map[defs.SamplerIdentifier]SamplerEntry{
				{Resource: "resource1", Name: "sampler1"}: {
					Resource:     "resource1",
					Name:         "sampler1",
					Config:       newSamplerConfig("stream1"),
					Capabilities: newSamplerCapabilities(),
				},
				{Resource: "resource3", Name: "sampler3"}: {
					Resource:     "resource3",
					Name:         "sampler3",
					Config:       newSamplerConfig("stream3"),
					Capabilities: newSamplerCapabilities(),
				},
			},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			s := storageProvider()
			got := map[defs.SamplerIdentifier]SamplerEntry{}
			err := s.RangeSamplers(func(entry SamplerEntry) {
				got[defs.NewSamplerIdentifier(entry.Resource, entry.Name)] = entry
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
		resource     string
		sampler      string
		config       control.SamplerConfig
		capabilities control.Capabilities
	}
	tests := []struct {
		name    string
		args    args
		want    SamplerEntry
		wantErr bool
	}{
		{
			name: "test setting non existing value",
			args: args{
				resource:     "resource2",
				sampler:      "sampler2",
				config:       newSamplerConfig("stream2"),
				capabilities: newSamplerCapabilities(),
			},
			want: SamplerEntry{
				Resource:     "resource2",
				Name:         "sampler2",
				Config:       newSamplerConfig("stream2"),
				Capabilities: newSamplerCapabilities(),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := storageProvider()
			err := s.SetSampler(SamplerEntry{
				Resource:     tt.args.resource,
				Name:         tt.args.sampler,
				Config:       tt.args.config,
				Capabilities: tt.args.capabilities,
			})
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
