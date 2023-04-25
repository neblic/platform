package internal

import (
	"context"
	"fmt"
	"strings"

	"github.com/c-bata/go-prompt"
	"github.com/neblic/platform/cmd/neblictl/internal/interpoler"
)

func Suggestions(ctx context.Context, nodes interpoler.CommandNodes, command *interpoler.TokanizedCommand) []prompt.Suggest {
	// Interpolate command
	result := interpoler.Interpolate(ctx, command, nodes)

	// Generate full list of suggestions. Populae the suggestion prefix if a partial command/parameter was provided
	allSuggestions := []string{}
	suggestionPrefix := ""
	if result.Error != nil && !result.MisplacedCursor {
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
			if command.HasTrailingSpace || len(result.Remaining.Tokens) > 0 {
				break
			}

			for _, node := range nodes {
				allSuggestions = append(allSuggestions, node.Name)
			}
			suggestionPrefix = result.Error.Details

		// Invalid parameter name
		case interpoler.ErrInvalidParameterName:
			for _, parameter := range result.Target.Parameters {
				// filter out already provided parameters
				if !result.Parameters.IsSet(parameter.Name) {
					allSuggestions = append(allSuggestions, fmt.Sprintf("--%s", parameter.Name))
				}
			}
			suggestionPrefix = result.Error.Details

		// Missing parameter value
		case interpoler.ErrInvalidParameterValue:
			fallthrough
		case interpoler.ErrEmptyParamaterValue:
			if result.TargetParameter.Completer != nil {
				allSuggestions = result.TargetParameter.Completer(ctx, result.Parameters)
			}
			suggestionPrefix = result.Error.Details
		}
	} else {
		// if the user is still writting, show parameters suggestions
		if result.Error == nil && command.HasTrailingSpace {
			for _, parameter := range result.Target.Parameters {
				// filter out already provided parameters
				if !result.Parameters.IsSet(parameter.Name) {
					allSuggestions = append(allSuggestions, fmt.Sprintf("--%s", parameter.Name))
				}
			}
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
