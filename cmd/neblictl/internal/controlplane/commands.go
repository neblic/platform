package controlplane

import (
	"github.com/neblic/platform/cmd/neblictl/internal/interpoler"
)

type Commands struct {
	Commands []*interpoler.Command
}

func NewCommands(controlPlaneExecutors *Executors, controlPlaneCompleters *Completers) *Commands {
	return &Commands{
		Commands: []*interpoler.Command{
			{
				Name:        "list",
				Description: "List elements",
				Subcommands: []*interpoler.Command{
					{
						Name:        "resources",
						Description: "List resources",
						Executor:    controlPlaneExecutors.ListResources,
					},
					{
						Name:        "samplers",
						Description: "List samplers",
						Executor:    controlPlaneExecutors.ListSamplers,
					},
					{
						Name:        "rules",
						Description: "List rules",
						Parameters: []interpoler.Parameter{
							{
								Name:        "resource",
								Description: "Resource where the sampler have been defined",
								Completer:   controlPlaneCompleters.ListResources,
							},
							{
								Name:        "sampler",
								Description: "Name of an already configured sampler",
								Completer:   controlPlaneCompleters.ListSamplers,
							},
						},
						Executor: controlPlaneExecutors.ListRules,
					},
				},
			},
			{
				Name:        "create",
				Description: "Create sampling configuration for a specific resource and sampler",
				Subcommands: []*interpoler.Command{
					{
						Name:        "rule",
						Description: "Create sampling rule for a specific resource and sampler",
						Parameters: []interpoler.Parameter{
							{
								Name:        "resource",
								Description: "Resource where the sampler have been defined",
								Completer:   controlPlaneCompleters.ListResources,
							},
							{
								Name:        "sampler",
								Description: "Name of an already configured sampler",
								Completer:   controlPlaneCompleters.ListSamplers,
							},
							{
								Name:        "sampling_rule",
								Description: "Sampling rule, format TBD",
							},
						},
						Executor: controlPlaneExecutors.CreateRule,
					},
					{
						Name:        "rate",
						Description: "Create sampling rate for a specific resource and sampler",
						Parameters: []interpoler.Parameter{
							{
								Name:        "resource",
								Description: "Resource where the sampler have been defined",
								Completer:   controlPlaneCompleters.ListResources,
							},
							{
								Name:        "sampler",
								Description: "Name of an already configured sampler",
								Completer:   controlPlaneCompleters.ListSamplers,
							},
							{
								Name:        "limit",
								Description: "Maximum number of samples per second exported",
							},
						},
						Executor: controlPlaneExecutors.CreateRate,
					},
				},
			},
			{
				Name:        "update",
				Description: "Update sampling configuration for a specific resource and sampler",
				Subcommands: []*interpoler.Command{
					{
						Name:        "rule",
						Description: "Update sampling rule for a specific resource and sampler",
						Parameters: []interpoler.Parameter{
							{
								Name:        "resource",
								Description: "Resource where the sampler have been defined",
								Completer:   controlPlaneCompleters.ListResources,
							},
							{
								Name:        "sampler",
								Description: "Name of an already configured sampler",
								Completer:   controlPlaneCompleters.ListSamplers,
							},
							{
								Name:        "old_sampling_rule",
								Description: "Old sampling rule, format TBD",
							},
							{
								Name:        "new_sampling_rule",
								Description: "New sampling rule, format TBD",
							},
						},
						Executor: controlPlaneExecutors.UpdateRule,
					},
					{
						Name:        "rate",
						Description: "Update sampling rate for a specific resource and sampler",
						Parameters: []interpoler.Parameter{
							{
								Name:        "resource",
								Description: "Resource where the sampler have been defined",
								Completer:   controlPlaneCompleters.ListResources,
							},
							{
								Name:        "sampler",
								Description: "Name of an already configured sampler",
								Completer:   controlPlaneCompleters.ListSamplers,
							},
							{
								Name:        "limit",
								Description: "Maximum number of samples per second exported",
							},
						},
						Executor: controlPlaneExecutors.UpdateRate,
					},
				},
			},
			{
				Name:        "delete",
				Description: "Update sampling configuration for a specific resource and sampler",
				Subcommands: []*interpoler.Command{
					{
						Name:        "rule",
						Description: "Update sampling rule for a specific resource and sampler",
						Parameters: []interpoler.Parameter{
							{
								Name:        "resource",
								Description: "Resource where the sampler have been defined",
								Completer:   controlPlaneCompleters.ListResources,
							},
							{
								Name:        "sampler",
								Description: "Name of an already configured sampler",
								Completer:   controlPlaneCompleters.ListSamplers,
							},
							{
								Name:        "sampling_rule",
								Description: "Sampling rule, format TBD",
							},
						},
						Executor: controlPlaneExecutors.DeleteRule,
					},
					{
						Name:        "rate",
						Description: "Update sampling rate for a specific resource and sampler",
						Parameters: []interpoler.Parameter{
							{
								Name:        "resource",
								Description: "Resource where the sampler have been defined",
								Completer:   controlPlaneCompleters.ListResources,
							},
							{
								Name:        "sampler",
								Description: "Name of an already configured sampler",
								Completer:   controlPlaneCompleters.ListSamplers,
							},
						},
						Executor: controlPlaneExecutors.DeleteRate,
					},
				},
			},
		},
	}
}
