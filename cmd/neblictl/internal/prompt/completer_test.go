package internal

import (
	"reflect"
	"testing"

	"github.com/c-bata/go-prompt"
	"github.com/neblic/platform/cmd/neblictl/internal/interpoler"
)

var commands = interpoler.CommandNodes{
	{
		Name:        "list",
		Description: "List elements",
		Subcommands: interpoler.CommandNodes{
			{
				Name:        "samplers",
				Description: "List samplers",
			},
			{
				Name:        "rules",
				Description: "List rules",
				Parameters: []interpoler.Parameter{
					{
						Name:        "sampler",
						Description: "Name of an already configured sampler",
						Completer: func(funcOptions interpoler.ParametersWithValue) []string {
							return []string{"p1", "p2"}
						},
					},
				},
			},
		},
	},
	{
		Name:        "create",
		Description: "Create a sampling rule for a specific sampler",
		Parameters: []interpoler.Parameter{
			{
				Name:        "sampler",
				Description: "Name of an already configured sampler",
				Completer: func(funcOptions interpoler.ParametersWithValue) []string {
					return []string{"p1", "p2"}
				},
			},
			{
				Name:        "sampling_rule",
				Description: "Sampling rule, format TBD",
				Completer: func(funcOptions interpoler.ParametersWithValue) []string {
					return []string{}
				},
			},
		},
	},
	{
		Name:        "update",
		Description: "Update sampling rule for a specific sampler",
		Parameters: []interpoler.Parameter{
			{
				Name:        "sampler",
				Description: "Name of an already configured sampler",
				Completer: func(funcOptions interpoler.ParametersWithValue) []string {
					return []string{"p1", "p2"}
				},
			},
			{
				Name:        "old_sampling_rule",
				Description: "Old sampling rule, format TBD",
				Completer: func(funcOptions interpoler.ParametersWithValue) []string {
					samplerParameter, _ := funcOptions.Get("sampler")
					switch samplerParameter.Value {
					case "p1":
						return []string{"p1s1", "p1s2"}
					case "p2":
						return []string{}
					default:
						panic("sampler does not exist")
					}
				},
			},
			{
				Name:        "new_sampling_rule",
				Description: "New sampling rule, format TBD",
				Completer: func(funcOptions interpoler.ParametersWithValue) []string {
					return []string{}
				},
			},
		},
	},
}

func TestSuggestions(t *testing.T) {
	type args struct {
		commands         []*interpoler.CommandNode
		tokanizedCommand *interpoler.TokanizedCommand
	}
	tests := []struct {
		name string
		args args
		want []prompt.Suggest
	}{
		{
			name: "Empty command",
			args: args{
				commands:         commands,
				tokanizedCommand: interpoler.NewTokanizedCommand([]string{}, false),
			},
			want: []prompt.Suggest{{Text: "list"}, {Text: "create"}, {Text: "update"}},
		},
		{
			name: "Partial command",
			args: args{
				commands:         commands,
				tokanizedCommand: interpoler.NewTokanizedCommand([]string{"li"}, false),
			},
			want: []prompt.Suggest{{Text: "list"}},
		},
		{
			name: "Suggest subcommand",
			args: args{
				commands:         commands,
				tokanizedCommand: interpoler.NewTokanizedCommand([]string{"list", ""}, false),
			},
			want: []prompt.Suggest{{Text: "samplers"}, {Text: "rules"}},
		},
		{
			name: "Suggest partial subcommand",
			args: args{
				commands:         commands,
				tokanizedCommand: interpoler.NewTokanizedCommand([]string{"list", "sa"}, false),
			},
			want: []prompt.Suggest{{Text: "samplers"}},
		},
		{
			name: "Suggest partial subcommand with trailing space",
			args: args{
				commands:         commands,
				tokanizedCommand: interpoler.NewTokanizedCommand([]string{"list", "sa"}, true),
			},
			want: []prompt.Suggest{},
		},
		{
			name: "Suggest partial subcommand with trailing space",
			args: args{
				commands:         commands,
				tokanizedCommand: interpoler.NewTokanizedCommand([]string{"list", "sa", "a"}, false),
			},
			want: []prompt.Suggest{},
		},
		{
			name: "Suggest parameter value",
			args: args{
				commands:         commands,
				tokanizedCommand: interpoler.NewTokanizedCommand([]string{"list", "rules", ""}, false),
			},
			want: []prompt.Suggest{{Text: "p1"}, {Text: "p2"}},
		},
		{
			name: "Suggest partial parameter value",
			args: args{
				commands:         commands,
				tokanizedCommand: interpoler.NewTokanizedCommand([]string{"list", "rules", "p"}, false),
			},
			want: []prompt.Suggest{{Text: "p1"}, {Text: "p2"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Suggestions(tt.args.commands, tt.args.tokanizedCommand); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Suggestions() = %v, want %v", got, tt.want)
			}
		})
	}
}
