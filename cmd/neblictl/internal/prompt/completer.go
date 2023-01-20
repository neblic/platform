package internal

import (
	"strings"

	"github.com/c-bata/go-prompt"
	"github.com/neblic/platform/cmd/neblictl/internal/interpoler"
)

func Suggestions(commands []*interpoler.Command, commandParts []string) []prompt.Suggest {
	// Interpolate command
	result := interpoler.Interpolate(commandParts, commands)

	// Generate full list of suggestions. Populae the suggestion prefix if a partial command/parameter was provided
	allSuggestions := []string{}
	suggestionPrefix := ""
	if result.Error != nil {
		switch result.Error.Error {

		// An empty command was provided (a command that does not have any part). Use the full list of commands to extract
		// the suggestions
		case interpoler.ErrEmptyCommand:
			for _, command := range commands {
				allSuggestions = append(allSuggestions, command.Name)
			}

		// An invalid command was found. The error details contain the command value (empty string if nothing was introduced)
		case interpoler.ErrInvalidCommand:
			// In case of having a target, it means the last valid part was the target itself, the invalid command received
			// will need to match one of the target subcommands.
			// If the target is nil, it means we are dealing with the first part of the command, the full list of commands will
			// be used to generate the suggestions
			allCommands := commands
			if result.Target != nil {
				allCommands = result.Target.Subcommands
			}

			for _, command := range allCommands {
				allSuggestions = append(allSuggestions, command.Name)
			}
			suggestionPrefix = result.Error.Details

		// A valid command or parameter was just processed and the next one was not started, nothing to complete.
		case interpoler.ErrPartialCommand, interpoler.ErrParameterNotProvided:

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
