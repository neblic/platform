package controlplane

import (
	"github.com/neblic/platform/cmd/neblictl/internal/interpoler"
)

type Commands struct {
	Commands interpoler.CommandNodes
}

func NewCommands(controlPlaneExecutors *Executors, controlPlaneCompleters *Completers) *Commands {
	return &Commands{
		Commands: interpoler.CommandNodes{
			{
				Name:        "resources:list",
				Description: "List all resources",
				Executor:    controlPlaneExecutors.ListResources,
			},
			{
				Name:        "samplers:list",
				Description: "List all samplers",
				Executor:    controlPlaneExecutors.ListSamplers,
			},
			{
				Name:        "samplers:limiterout:set",
				Description: "Sets the maximum number of samples exported per second by a sampler",
				Executor:    controlPlaneExecutors.SamplerLimiterOutSet,
				Parameters: []interpoler.Parameter{
					{
						Name:        "limit",
						Description: "Maximum number of samples per second exported",
					},
					{
						Name:        "resource",
						Description: "Filter streams by resource",
						Completer:   controlPlaneCompleters.ListResources,
						Optional:    true,
						Default:     "*",
					},
					{
						Name:        "sampler",
						Description: "Filter streams by sampler",
						Completer:   controlPlaneCompleters.ListSamplers,
						Optional:    true,
						Default:     "*",
					},
				},
			},
			{
				Name:        "samplers:limiterout:unset",
				Description: "Unsets the maximum number of samples per second exported by a sampler",
				Executor:    controlPlaneExecutors.SamplerLimiterOutUnset,
				Parameters: []interpoler.Parameter{
					{
						Name:        "resource",
						Description: "Filter streams by resource",
						Completer:   controlPlaneCompleters.ListResources,
						Optional:    true,
						Default:     "*",
					},
					{
						Name:        "sampler",
						Description: "Filter streams by sampler",
						Completer:   controlPlaneCompleters.ListSamplers,
						Optional:    true,
						Default:     "*",
					},
				},
			},
			{
				Name:        "streams:list",
				Description: "List streams",
				Executor:    controlPlaneExecutors.ListStreams,
				Parameters: []interpoler.Parameter{
					{
						Name:        "resource",
						Description: "Filter streams by resource",
						Completer:   controlPlaneCompleters.ListResources,
						Optional:    true,
						Default:     "*",
					},
					{
						Name:        "sampler",
						Description: "Filter streams by sampler",
						Completer:   controlPlaneCompleters.ListSamplers,
						Optional:    true,
						Default:     "*",
					},
				},
			},
			{
				Name:        "streams:create",
				Description: "Create streams",
				ExtendedDescription: `* It it possible to create multiple streams targeting different samplers at once. 
* If the uid is not specified, a random one will be generated.
* All the created streams will have the same UID`,
				Parameters: []interpoler.Parameter{
					{
						Name:        "rule",
						Description: "CEL rule that will select the stream elements",
					},
					{
						Name:        "uid",
						Description: "Stream uid",
						Optional:    true,
						DoNotFilter: true,
					},
					{
						Name:        "resource",
						Description: "Filter streams by resource",
						Completer:   controlPlaneCompleters.ListResources,
						Optional:    true,
						Default:     "*",
					},
					{
						Name:        "sampler",
						Description: "Filter streams by sampler",
						Completer:   controlPlaneCompleters.ListSamplers,
						Optional:    true,
						Default:     "*",
					},
				},
				Executor: controlPlaneExecutors.CreateStreams,
			},
			{
				Name:                "streams:update",
				Description:         "Update streams",
				ExtendedDescription: `* It it possible to update multiple streams targeting different samplers at once.`,
				Parameters: []interpoler.Parameter{
					{
						Name:        "uid",
						Description: "Stream uid",
						Completer:   controlPlaneCompleters.ListStreamsUID,
					},
					{
						Name:        "updated-rule",
						Description: "Updated CEL rule",
					},
					{
						Name:        "resource",
						Description: "Filter streams by resource",
						Completer:   controlPlaneCompleters.ListResources,
						Optional:    true,
						Default:     "*",
					},
					{
						Name:        "sampler",
						Description: "Filter streams by sampler",
						Completer:   controlPlaneCompleters.ListSamplers,
						Optional:    true,
						Default:     "*",
					},
				},
				Executor: controlPlaneExecutors.UpdateStreams,
			},
			{
				Name:                "streams:delete",
				Description:         "Delete streams",
				ExtendedDescription: `* It it possible to delete multiple streams targeting different samplers at once.`,
				Parameters: []interpoler.Parameter{
					{
						Name:        "uid",
						Description: "Stream uid",
						Completer:   controlPlaneCompleters.ListStreamsUID,
					},
					{
						Name:        "resource",
						Description: "Filter streams by resource",
						Completer:   controlPlaneCompleters.ListResources,
						Optional:    true,
						Default:     "*",
					},
					{
						Name:        "sampler",
						Description: "Filter streams by sampler",
						Completer:   controlPlaneCompleters.ListSamplers,
						Optional:    true,
						Default:     "*",
					},
				},
				Executor: controlPlaneExecutors.DeleteStreams,
			},
		},
	}
}
