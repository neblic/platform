package internal

import (
	"context"
	"reflect"
	"testing"

	"github.com/c-bata/go-prompt"
	"github.com/neblic/platform/cmd/neblictl/internal/interpoler"
)

var commands = interpoler.CommandNodes{
	{
		Name: "streams:create",
		Parameters: []interpoler.Parameter{
			{
				Name:        "sampler",
				Description: "Filter streams by sampler",
				Completer: func(_ context.Context, funcOptions interpoler.ParametersWithValue) []string {
					return []string{"sampler1", "sampler2"}
				},
			},
			{
				Name:        "uid",
				Description: "Stream uid",
			},
			{
				Name:        "rule",
				Description: "CEL rule that will select the stream elements",
			},
		},
	},
	{
		Name: "streams:update",
		Parameters: []interpoler.Parameter{
			{
				Name:        "sampler",
				Description: "Name of an already configured sampler",
				Completer: func(_ context.Context, funcOptions interpoler.ParametersWithValue) []string {
					return []string{"sampler1", "sampler2"}
				},
			},
			{
				Name:        "uid",
				Description: "Stream uid",
				Completer: func(_ context.Context, funcOptions interpoler.ParametersWithValue) []string {
					uidParamter, ok := funcOptions.Get("uid")
					if !ok {
						return []string{"sampler1s1", "sampler1s2", "sampler2s1"}
					}

					switch uidParamter.Value {
					case "sampler1":
						return []string{"sampler1s1", "sampler1s2"}
					case "sampler2":
						return []string{"sampler2s1"}
					default:
						panic("sampler does not exist")
					}
				},
			},
			{
				Name:        "updated-rule",
				Description: "Updated CEL rule",
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
				tokanizedCommand: interpoler.NewTokanizedCommand([]interpoler.Token{}, false, 0),
			},
			want: []prompt.Suggest{{Text: "streams:create"}, {Text: "streams:update"}},
		},
		{
			name: "Partial command",
			args: args{
				commands:         commands,
				tokanizedCommand: interpoler.NewTokanizedCommand([]interpoler.Token{{Value: "streams"}}, false, 0),
			},
			want: []prompt.Suggest{{Text: "streams:create"}, {Text: "streams:update"}},
		},
		{
			name: "Suggest parameter value",
			args: args{
				commands:         commands,
				tokanizedCommand: interpoler.NewTokanizedCommand([]interpoler.Token{{Value: "streams:create"}, {Value: "--sampler"}, {Value: "", Pos: 10}}, false, 10),
			},
			want: []prompt.Suggest{{Text: "sampler1"}, {Text: "sampler2"}},
		},
		{
			name: "Suggest partial parameter value",
			args: args{
				commands:         commands,
				tokanizedCommand: interpoler.NewTokanizedCommand([]interpoler.Token{{Value: "streams:update"}, {Value: "--uid"}, {Value: "sampler1", Pos: 6}}, false, 14),
			},
			want: []prompt.Suggest{{Text: "sampler1s1"}, {Text: "sampler1s2"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Suggestions(context.Background(), tt.args.commands, tt.args.tokanizedCommand); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Suggestions() = %v, want %v", got, tt.want)
			}
		})
	}
}
