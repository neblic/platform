package controlplane

import (
	"context"

	"github.com/neblic/platform/cmd/neblictl/internal/interpoler"
)

type Commands struct {
	Commands interpoler.CommandNodes
}

func NewCommands(controlPlaneExecutors *Executors, controlPlaneCompleters *Completers) *Commands {
	return &Commands{
		Commands: interpoler.CommandNodes{

			// resources
			{
				Name:        "resources:list",
				Description: "List all resources",
				Executor:    controlPlaneExecutors.ResourcesList,
				Parameters: []interpoler.Parameter{
					{
						Name:        "resource-name",
						Description: "Filter by resource",
						Completer:   controlPlaneCompleters.ListResourcesUID,
						Filter:      true,
						Optional:    true,
						Default:     "*",
					},
					{
						Name:        "sampler-name",
						Description: "Filter by sampler",
						Completer:   controlPlaneCompleters.ListSamplersUID,
						Filter:      true,
						Optional:    true,
						Default:     "*",
					},
				},
			},

			// samplers
			{
				Name:        "samplers:list",
				Description: "List all samplers",
				Executor:    controlPlaneExecutors.SamplersList,
				Parameters: []interpoler.Parameter{
					{
						Name:        "resource-name",
						Description: "Filter by resource",
						Completer:   controlPlaneCompleters.ListResourcesUID,
						Filter:      true,
						Optional:    true,
						Default:     "*",
					},
					{
						Name:        "sampler-name",
						Description: "Filter by sampler",
						Completer:   controlPlaneCompleters.ListSamplersUID,
						Filter:      true,
						Optional:    true,
						Default:     "*",
					},
				},
			},

			// samplers:config:list
			{
				Name:        "samplers:list:config",
				Description: "List all samplers configurations",
				Executor:    controlPlaneExecutors.SamplersListConfig,
				Parameters: []interpoler.Parameter{
					{
						Name:        "resource-name",
						Description: "Filter by resource",
						Completer:   controlPlaneCompleters.ListResourcesUID,
						Filter:      true,
						Optional:    true,
						Default:     "*",
					},
					{
						Name:        "sampler-name",
						Description: "Filter by sampler",
						Completer:   controlPlaneCompleters.ListSamplersUID,
						Filter:      true,
						Optional:    true,
						Default:     "*",
					},
				},
			},

			// samplers:limiterin
			{
				Name:        "samplers:limiterin:set",
				Description: "Sets the maximum number of samples processed per second by a sampler",
				Executor:    controlPlaneExecutors.SamplersLimiterInSet,
				Parameters: []interpoler.Parameter{
					{
						Name:        "limit",
						Description: "Maximum number of samples per second processed",
					},
					{
						Name:        "resource-name",
						Description: "Filter by resource",
						Completer:   controlPlaneCompleters.ListResourcesUID,
						Filter:      true,
						Optional:    true,
						Default:     "*",
					},
					{
						Name:        "sampler-name",
						Description: "Filter by sampler",
						Completer:   controlPlaneCompleters.ListSamplersUID,
						Filter:      true,
						Optional:    true,
						Default:     "*",
					},
				},
			},
			{
				Name:        "samplers:limiterin:unset",
				Description: "Unsets the maximum number of samples per second processed by a sampler",
				Executor:    controlPlaneExecutors.SamplersLimiterInUnset,
				Parameters: []interpoler.Parameter{
					{
						Name:        "resource-name",
						Description: "Filter by resource",
						Completer:   controlPlaneCompleters.ListResourcesUID,
						Filter:      true,
						Optional:    true,
						Default:     "*",
					},
					{
						Name:        "sampler-name",
						Description: "Filter by sampler",
						Completer:   controlPlaneCompleters.ListSamplersUID,
						Filter:      true,
						Optional:    true,
						Default:     "*",
					},
				},
			},

			// samplers:samplerin:deterministic
			{
				Name:        "samplers:samplerin:set:deterministic",
				Description: "Sets a deterministic samplerin configuration",
				Executor:    controlPlaneExecutors.SamplersSamplerInSetDeterministic,
				Parameters: []interpoler.Parameter{
					{
						Name:        "sample_rate",
						Description: "Deterministic sampling sample rate. 1 means all samples are sampled",
					},
					{
						Name:        "sample_empty_determinant",
						Description: "Boolean value to determine if samples with an empty determinant should be sampled",
						Completer: func(ctx context.Context, funcOptions interpoler.ParametersWithValue) []string {
							return []string{"true", "false"}
						},
						Optional: true,
						Default:  "false",
					},
					{
						Name:        "resource-name",
						Description: "Filter by resource",
						Completer:   controlPlaneCompleters.ListResourcesUID,
						Filter:      true,
						Optional:    true,
						Default:     "*",
					},
					{
						Name:        "sampler-name",
						Description: "Filter by sampler",
						Completer:   controlPlaneCompleters.ListSamplersUID,
						Filter:      true,
						Optional:    true,
						Default:     "*",
					},
				},
			},
			{
				Name:        "samplers:samplerin:unset",
				Description: "Unsets any samplerin configuration set",
				Executor:    controlPlaneExecutors.SamplersSamplerInUnset,
				Parameters: []interpoler.Parameter{
					{
						Name:        "resource-name",
						Description: "Filter by resource",
						Completer:   controlPlaneCompleters.ListResourcesUID,
						Filter:      true,
						Optional:    true,
						Default:     "*",
					},
					{
						Name:        "sampler-name",
						Description: "Filter by sampler",
						Completer:   controlPlaneCompleters.ListSamplersUID,
						Filter:      true,
						Optional:    true,
						Default:     "*",
					},
				},
			},

			// samplers:limiterout
			{
				Name:        "samplers:limiterout:set",
				Description: "Sets the maximum number of samples exported per second by a sampler",
				Executor:    controlPlaneExecutors.SamplersLimiterOutSet,
				Parameters: []interpoler.Parameter{
					{
						Name:        "limit",
						Description: "Maximum number of samples per second exported",
					},
					{
						Name:        "resource-name",
						Description: "Filter by resource",
						Completer:   controlPlaneCompleters.ListResourcesUID,
						Filter:      true,
						Optional:    true,
						Default:     "*",
					},
					{
						Name:        "sampler-name",
						Description: "Filter by sampler",
						Completer:   controlPlaneCompleters.ListSamplersUID,
						Filter:      true,
						Optional:    true,
						Default:     "*",
					},
				},
			},
			{
				Name:        "samplers:limiterout:unset",
				Description: "Unsets the maximum number of samples per second exported by a sampler",
				Executor:    controlPlaneExecutors.SamplersLimiterOutUnset,
				Parameters: []interpoler.Parameter{
					{
						Name:        "resource-name",
						Description: "Filter by resource",
						Completer:   controlPlaneCompleters.ListResourcesUID,
						Filter:      true,
						Optional:    true,
						Default:     "*",
					},
					{
						Name:        "sampler-name",
						Description: "Filter by sampler",
						Completer:   controlPlaneCompleters.ListSamplersUID,
						Filter:      true,
						Optional:    true,
						Default:     "*",
					},
				},
			},

			// streams
			{
				Name:        "streams:list",
				Description: "List streams",
				Executor:    controlPlaneExecutors.StreamsList,
				Parameters: []interpoler.Parameter{
					{
						Name:        "resource-name",
						Description: "Filter by resource",
						Completer:   controlPlaneCompleters.ListResourcesUID,
						Filter:      true,
						Optional:    true,
						Default:     "*",
					},
					{
						Name:        "sampler-name",
						Description: "Filter by sampler",
						Completer:   controlPlaneCompleters.ListSamplersUID,
						Filter:      true,
						Optional:    true,
						Default:     "*",
					},
					{
						Name:        "stream-name",
						Description: "Filter by stream name",
						Filter:      true,
						Completer:   controlPlaneCompleters.ListStreamsName,
						Optional:    true,
						Default:     "",
					},
				},
			},
			{
				Name:        "streams:create",
				Description: "Create streams",
				ExtendedDescription: `* It it possible to create multiple streams targeting different samplers at once.
* All the created streams will have the same name`,
				Parameters: []interpoler.Parameter{
					{
						Name:        "rule",
						Description: "CEL rule that will select the stream elements",
					},
					{
						Name:        "stream-name",
						Description: "Stream name",
					},
					{
						Name:        "resource-name",
						Description: "Filter by resource",
						Completer:   controlPlaneCompleters.ListResourcesUID,
						Filter:      true,
						Optional:    true,
						Default:     "*",
					},
					{
						Name:        "sampler-name",
						Description: "Filter by sampler",
						Completer:   controlPlaneCompleters.ListSamplersUID,
						Filter:      true,
						Optional:    true,
						Default:     "*",
					},
					{
						Name:        "export-raw",
						Description: "Export raw samples",
						Completer: func(ctx context.Context, funcOptions interpoler.ParametersWithValue) []string {
							return []string{"true", "false"}
						},
						Optional: true,
						Default:  "false",
					},
					{
						Name:        "keyed",
						Description: "Define as a keyed stream. During the computation of an event in a keyed stream, each different key will have its own state",
						Completer: func(ctx context.Context, funcOptions interpoler.ParametersWithValue) []string {
							return []string{"true", "false"}
						},
						Optional: true,
						Default:  "false",
					},
					{
						Name:        "keyed-ttl",
						Description: "If the stream is keyed, this parameter controls the time to live for each key. Follows golang duration format",
						Optional:    true,
						Default:     "1h",
					},
					{
						Name:        "keyed-max-keys",
						Description: "If the stream is keyed, this parameter controls the maximum number of keys to keep",
						Optional:    true,
						Default:     "1000",
					},
					{
						Name:        "max-sample-size",
						Description: "Samples larger than this size will be dropped",
						Optional:    true,
						Default:     "10240",
					},
				},
				Executor: controlPlaneExecutors.StreamsCreate,
			},
			{
				Name:                "streams:update",
				Description:         "Update streams",
				ExtendedDescription: `* It it possible to update multiple streams targeting different samplers at once.`,
				Parameters: []interpoler.Parameter{
					{
						Name:        "stream-name",
						Description: "Stream name",
						Completer:   controlPlaneCompleters.ListStreamsName,
					},
					{
						Name:        "updated-rule",
						Description: "Updated CEL rule",
					},
					{
						Name:        "resource-name",
						Description: "Filter by resource",
						Completer:   controlPlaneCompleters.ListResourcesUID,
						Filter:      true,
						Optional:    true,
						Default:     "*",
					},
					{
						Name:        "sampler-name",
						Description: "Filter by sampler",
						Completer:   controlPlaneCompleters.ListSamplersUID,
						Filter:      true,
						Optional:    true,
						Default:     "*",
					},
					{
						Name:        "export-raw",
						Description: "Export raw samples",
						Completer: func(ctx context.Context, funcOptions interpoler.ParametersWithValue) []string {
							return []string{"true", "false"}
						},
						Optional: true,
						Default:  "false",
					},
					{
						Name:        "keyed",
						Description: "Define as a keyed stream. During the computation of an event in a keyed stream, each different key will have its own state",
						Completer: func(ctx context.Context, funcOptions interpoler.ParametersWithValue) []string {
							return []string{"true", "false"}
						},
						Optional: true,
						Default:  "false",
					},
					{
						Name:        "keyed-ttl",
						Description: "If the stream is keyed, this parameter controls the time to live for each key. Follows golang duration format",
						Optional:    true,
						Default:     "1h",
					},
					{
						Name:        "keyed-max-keys",
						Description: "If the stream is keyed, this parameter controls the maximum number of keys to keep",
						Optional:    true,
						Default:     "1000",
					},
					{
						Name:        "max-sample-size",
						Description: "Samples larger than this size will be dropped",
						Optional:    true,
						Default:     "10240",
					},
				},
				Executor: controlPlaneExecutors.StreamsUpdate,
			},
			{
				Name:                "streams:delete",
				Description:         "Delete streams",
				ExtendedDescription: `* It it possible to delete multiple streams targeting different samplers at once.`,
				Parameters: []interpoler.Parameter{
					{
						Name:        "stream-name",
						Description: "Stream name",
						Completer:   controlPlaneCompleters.ListStreamsName,
					},
					{
						Name:        "resource-name",
						Description: "Filter by resource",
						Completer:   controlPlaneCompleters.ListResourcesUID,
						Filter:      true,
						Optional:    true,
						Default:     "*",
					},
					{
						Name:        "sampler-name",
						Description: "Filter by sampler",
						Completer:   controlPlaneCompleters.ListSamplersUID,
						Filter:      true,
						Optional:    true,
						Default:     "*",
					},
				},
				Executor: controlPlaneExecutors.StreamsDelete,
			},

			// Digests
			{
				Name:        "digests:list",
				Description: "List configured digests",
				Parameters: []interpoler.Parameter{
					{
						Name:        "resource-name",
						Description: "Filter by resource",
						Completer:   controlPlaneCompleters.ListResourcesUID,
						Filter:      true,
						Optional:    true,
						Default:     "*",
					},
					{
						Name:        "sampler-name",
						Description: "Filter by sampler",
						Completer:   controlPlaneCompleters.ListSamplersUID,
						Filter:      true,
						Optional:    true,
						Default:     "*",
					},
				},
				Executor: controlPlaneExecutors.DigestsList,
			},
			{
				Name:        "digests:structure:create",
				Description: "Generate structure digests",
				Parameters: []interpoler.Parameter{
					{
						Name:        "digest-name",
						Description: "Digest name",
					},
					{
						Name:        "stream-name",
						Description: "Stream name",
						Completer:   controlPlaneCompleters.ListStreamsName,
					},
					{
						Name:        "flush-period",
						Description: "Digests generation period (in seconds)",
						Optional:    true,
						Default:     "60",
					},
					{
						Name:        "max-processed-fields",
						Description: "Maximum number of fields per sample to process",
						Optional:    true,
						Default:     "100",
					},
					{
						Name:        "resource-name",
						Description: "Filter by resource",
						Completer:   controlPlaneCompleters.ListResourcesUID,
						Filter:      true,
						Optional:    true,
						Default:     "*",
					},
					{
						Name:        "sampler-name",
						Description: "Filter by sampler",
						Completer:   controlPlaneCompleters.ListSamplersUID,
						Filter:      true,
						Optional:    true,
						Default:     "*",
					},
				},
				Executor: controlPlaneExecutors.DigestsStructureCreate,
			},
			{
				Name:        "digests:structure:update",
				Description: "Update structure digests",
				Parameters: []interpoler.Parameter{
					{
						Name:        "digest-name",
						Description: "Digest name",
						Completer:   controlPlaneCompleters.ListDigestsName,
						Filter:      true,
					},
					{
						Name:        "stream-name",
						Description: "Stream name",
						Completer:   controlPlaneCompleters.ListStreamsName,
					},
					{
						Name:        "flush-period",
						Description: "Digests generation period (in seconds)",
						Optional:    true,
						Default:     "60",
					},
					{
						Name:        "max-processed-fields",
						Description: "Maximum number of fields per sample to process",
						Optional:    true,
						Default:     "100",
					},
					{
						Name:        "resource-name",
						Description: "Filter by resource",
						Filter:      true,
						Completer:   controlPlaneCompleters.ListResourcesUID,
						Optional:    true,
						Default:     "*",
					},
					{
						Name:        "sampler-name",
						Description: "Filter by sampler",
						Filter:      true,
						Completer:   controlPlaneCompleters.ListSamplersUID,
						Optional:    true,
						Default:     "*",
					},
				},
				Executor: controlPlaneExecutors.DigestsStructureUpdate,
			},
			{
				Name:        "digests:value:create",
				Description: "Configure generation of value digests",
				Parameters: []interpoler.Parameter{
					{
						Name:        "digest-name",
						Description: "Digest name",
					},
					{
						Name:        "stream-name",
						Description: "Stream name",
						Completer:   controlPlaneCompleters.ListStreamsName,
					},
					{
						Name:        "flush-period",
						Description: "Digests generation period (in seconds)",
						Optional:    true,
						Default:     "60",
					},
					{
						Name:        "max-processed-fields",
						Description: "Maximum number of fields to process when processing a sample",
						Optional:    true,
						Default:     "100",
					},
					{
						Name:        "resource-name",
						Description: "Filter by resource",
						Completer:   controlPlaneCompleters.ListResourcesUID,
						Optional:    true,
						Default:     "*",
					},
					{
						Name:        "sampler-name",
						Description: "Filter by sampler",
						Completer:   controlPlaneCompleters.ListSamplersUID,
						Optional:    true,
						Default:     "*",
					},
				},
				Executor: controlPlaneExecutors.DigestsValueCreate,
			},
			{
				Name:        "digests:value:update",
				Description: "Update value digests",
				Parameters: []interpoler.Parameter{
					{
						Name:        "digest-name",
						Description: "Digest name",
						Completer:   controlPlaneCompleters.ListDigestsName,
					},
					{
						Name:        "stream-name",
						Description: "Stream name",
						Completer:   controlPlaneCompleters.ListStreamsName,
					},
					{
						Name:        "flush-period",
						Description: "Digests generation period (in seconds)",
						Optional:    true,
						Default:     "60",
					},
					{
						Name:        "max-processed-fields",
						Description: "Maximum number of fields to process when processing a sample",
						Optional:    true,
						Default:     "100",
					},
					{
						Name:        "resource-name",
						Description: "Filter by resource",
						Completer:   controlPlaneCompleters.ListResourcesUID,
						Optional:    true,
						Default:     "*",
					},
					{
						Name:        "sampler-name",
						Description: "Filter by sampler",
						Completer:   controlPlaneCompleters.ListSamplersUID,
						Optional:    true,
						Default:     "*",
					},
				},
				Executor: controlPlaneExecutors.DigestsValueUpdate,
			},
			{
				Name:        "digests:delete",
				Description: "Delete digest",
				Parameters: []interpoler.Parameter{
					{
						Name:        "digest-name",
						Description: "Digest name",
						Completer:   controlPlaneCompleters.ListDigestsName,
					},
					{
						Name:        "resource-name",
						Description: "Filter by resource",
						Completer:   controlPlaneCompleters.ListResourcesUID,
						Filter:      true,
						Optional:    true,
						Default:     "*",
					},
					{
						Name:        "sampler-name",
						Description: "Filter by sampler",
						Completer:   controlPlaneCompleters.ListSamplersUID,
						Filter:      true,
						Optional:    true,
						Default:     "*",
					},
				},
				Executor: controlPlaneExecutors.DigestsDelete,
			},

			// Events
			{
				Name:        "events:list",
				Description: "List configured events",
				Parameters: []interpoler.Parameter{
					{
						Name:        "resource-name",
						Description: "Filter by resource",
						Completer:   controlPlaneCompleters.ListResourcesUID,
						Filter:      true,
						Optional:    true,
						Default:     "*",
					},
					{
						Name:        "sampler-name",
						Description: "Filter by sampler",
						Completer:   controlPlaneCompleters.ListSamplersUID,
						Filter:      true,
						Optional:    true,
						Default:     "*",
					},
				},
				Executor: controlPlaneExecutors.EventsList,
			},
			{
				Name:        "events:create",
				Description: "Create events",
				Parameters: []interpoler.Parameter{
					{
						Name:        "event-name",
						Description: "event name",
					},
					{
						Name:        "stream-name",
						Description: "Stream name",
						Completer:   controlPlaneCompleters.ListStreamsName,
					},
					{
						Name:        "sample-type",
						Description: "Sample type",
						Completer:   controlPlaneCompleters.ListSampleType,
					},
					{
						Name:        "rule",
						Description: "CEL rule that will create events from elements in the the stream-name",
					},
					{
						Name:        "resource-name",
						Description: "Filter by resource",
						Completer:   controlPlaneCompleters.ListResourcesUID,
						Filter:      true,
						Optional:    true,
						Default:     "*",
					},
					{
						Name:        "sampler-name",
						Description: "Filter by sampler",
						Completer:   controlPlaneCompleters.ListSamplersUID,
						Filter:      true,
						Optional:    true,
						Default:     "*",
					},
					{
						Name:        "limit",
						Description: "Maximum number of events per second generated",
						Optional:    true,
						Default:     "10",
					},
					{
						Name:        "export-template",
						Description: "String template that will be interpolated when exporting the event",
						Optional:    true,
						Default:     "",
					},
				},
				Executor: controlPlaneExecutors.EventsCreate,
			},
			{
				Name:        "events:update",
				Description: "Update events",
				Parameters: []interpoler.Parameter{
					{
						Name:        "event-name",
						Description: "Event name",
						Completer:   controlPlaneCompleters.ListEventsName,
					},
					{
						Name:        "stream-name",
						Description: "Stream name",
						Completer:   controlPlaneCompleters.ListStreamsName,
					},
					{
						Name:        "sample-type",
						Description: "Sample type",
						Completer:   controlPlaneCompleters.ListSampleType,
					},
					{
						Name:        "rule",
						Description: "CEL rule that will create events from elements in the the stream-name",
					},
					{
						Name:        "resource-name",
						Description: "Filter by resource",
						Completer:   controlPlaneCompleters.ListResourcesUID,
						Filter:      true,
						Optional:    true,
						Default:     "*",
					},
					{
						Name:        "sampler-name",
						Description: "Filter by sampler",
						Completer:   controlPlaneCompleters.ListSamplersUID,
						Filter:      true,
						Optional:    true,
						Default:     "*",
					},
					{
						Name:        "limit",
						Description: "Maximum number of events per second generated",
						Optional:    true,
						Default:     "10",
					},
					{
						Name:        "export-template",
						Description: "String template that will be interpolated when exporting the event",
						Optional:    true,
						Default:     "",
					},
				},
				Executor: controlPlaneExecutors.EventsUpdate,
			},
			{
				Name:        "events:delete",
				Description: "Delete events",
				Parameters: []interpoler.Parameter{
					{
						Name:        "event-name",
						Description: "Event name",
						Completer:   controlPlaneCompleters.ListEventsName,
					},
					{
						Name:        "resource-name",
						Description: "Filter by resource",
						Completer:   controlPlaneCompleters.ListResourcesUID,
						Filter:      true,
						Optional:    true,
						Default:     "*",
					},
					{
						Name:        "sampler-name",
						Description: "Filter by sampler",
						Completer:   controlPlaneCompleters.ListSamplersUID,
						Filter:      true,
						Optional:    true,
						Default:     "*",
					},
				},
				Executor: controlPlaneExecutors.EventsDelete,
			},
		},
	}
}
