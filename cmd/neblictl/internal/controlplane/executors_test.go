package controlplane

import (
	"testing"
)

func TestExecutors_doesResourceAndSamplerMatch(t *testing.T) {
	type fields struct {
		controlPlaneClient *Client
	}
	type args struct {
		resourceParameter       string
		samplerParameter        string
		resourceAndSamplerEntry resourceAndSampler
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "Match specific resource and sampler",
			fields: fields{
				controlPlaneClient: nil,
			},
			args: args{
				resourceParameter:       "resource",
				samplerParameter:        "sampler",
				resourceAndSamplerEntry: resourceAndSampler{"resource", "sampler"},
			},
			want: true,
		},
		{
			name: "Match specific resource and any sampler",
			fields: fields{
				controlPlaneClient: nil,
			},
			args: args{
				resourceParameter:       "resource",
				samplerParameter:        "*",
				resourceAndSamplerEntry: resourceAndSampler{"resource", "sampler"},
			},
			want: true,
		},
		{
			name: "Match any resource and specific sampler",
			fields: fields{
				controlPlaneClient: nil,
			},
			args: args{
				resourceParameter:       "*",
				samplerParameter:        "sampler",
				resourceAndSamplerEntry: resourceAndSampler{"resource", "sampler"},
			},
			want: true,
		},
		{
			name: "Match any resource and any sampler",
			fields: fields{
				controlPlaneClient: nil,
			},
			args: args{
				resourceParameter:       "*",
				samplerParameter:        "*",
				resourceAndSamplerEntry: resourceAndSampler{"resource", "sampler"},
			},
			want: true,
		},
		{
			name: "NOT matching resource and NOT matching sampler",
			fields: fields{
				controlPlaneClient: nil,
			},
			args: args{
				resourceParameter:       "invalid_resource",
				samplerParameter:        "invalid_parameter",
				resourceAndSamplerEntry: resourceAndSampler{"resource", "sampler"},
			},
			want: false,
		},
		{
			name: "NOT matching resource and any sampler",
			fields: fields{
				controlPlaneClient: nil,
			},
			args: args{
				resourceParameter:       "invalid_resource",
				samplerParameter:        "*",
				resourceAndSamplerEntry: resourceAndSampler{"resource", "sampler"},
			},
			want: false,
		},
		{
			name: "NOT matching resource and matching sampler",
			fields: fields{
				controlPlaneClient: nil,
			},
			args: args{
				resourceParameter:       "invalid_resource",
				samplerParameter:        "sampler",
				resourceAndSamplerEntry: resourceAndSampler{"resource", "sampler"},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &Executors{
				controlPlaneClient: tt.fields.controlPlaneClient,
			}
			if got := e.doesResourceAndSamplerMatch(tt.args.resourceParameter, tt.args.samplerParameter, tt.args.resourceAndSamplerEntry); got != tt.want {
				t.Errorf("Executors.doesResourceAndSamplerMatch() = %v, want %v", got, tt.want)
			}
		})
	}
}
