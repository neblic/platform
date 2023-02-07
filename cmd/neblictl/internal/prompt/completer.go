package internal

import (
	"strings"

	"github.com/c-bata/go-prompt"
	"github.com/neblic/platform/cmd/neblictl/internal/interpoler"
)

func Suggestions(nodes interpoler.CommandNodes, command *interpoler.TokanizedCommand) []prompt.Suggest {
	// Interpolate command
	result := interpoler.Interpolate(command, nodes)

	// Generate full list of suggestions. Populae the suggestion prefix if a partial command/parameter was provided
	allSuggestions := []string{}
	suggestionPrefix := ""
	if result.Error != nil {
		switch result.Error.Error {

		// An empty command was provided (a command that does not have any part). Use the full list of commands to extract
		// the suggestions
		case interpoler.ErrEmptyCommand:
			for _, node := range nodes {
				allSuggestions = append(allSuggestions, node.Name)
			}

		// An invalid command was found. The error details contain the command value (empty string if nothing was introduced)
		case interpoler.ErrInvalidCommand:

			// In case of having a trailing space or some remaining parts for the command that were not processed,
			// it means there is some garbage at the end of the command and no suggestion can be provided
			if command.HasTrailingSpace || len(result.Remaining.Parts) > 0 {
				break
			}

			// In case of having a target, it means the last valid part was the target itself, the invalid command received
			// will need to match one of the target subcommands.
			// If the target is nil, it means we are dealing with the first part of the command, the full list of commands will
			// be used to generate the suggestions
			allNodes := nodes
			if result.Target != nil {
				allNodes = result.Target.Subcommands
			}

			for _, node := range allNodes {
				allSuggestions = append(allSuggestions, node.Name)
			}
			suggestionPrefix = result.Error.Details

		// A valid partial command was found. In case of having a trailing space, the user is starting the next part
		// of the command, provide him with the list of subcommands as suggestions
		case interpoler.ErrPartialCommand:
			if command.HasTrailingSpace {
				for _, node := range result.Target.Subcommands {
					allSuggestions = append(allSuggestions, node.Name)
				}
			}

		// A valid command was found, but at least one of the parameters was not provided. In case of having a trailing space,
		// the last parameter function has to be run to know the suggestions.
		case interpoler.ErrParameterNotProvided:
			if command.HasTrailingSpace {
				parameter, _ := result.Parameters.GetLast()

				if parameter.Completer != nil {
					allSuggestions = parameter.Completer(result.Parameters)
				}
			}

		// An invalid parameter was found. The error details contain the parameter value (empty string if nothing was introduced)
		case interpoler.ErrInvalidParameter:
			parameter, _ := result.Parameters.GetLast()

			if parameter.Completer != nil {
				allSuggestions = parameter.Completer(result.Parameters)
			}
			suggestionPrefix = result.Error.Details
		}
	}

	// Use suggestion prefix to filter out the suggestions that don't match. If an empty suggestion prefix is provided
	// all the suggestions will be valid
	suggestions := []prompt.Suggest{}
	for _, suggestion := range allSuggestions {
		if !strings.HasPrefix(suggestion, suggestionPrefix) {
			continue
		}
		suggestions = append(suggestions, prompt.Suggest{Text: suggestion})
	}
	return suggestions
}
