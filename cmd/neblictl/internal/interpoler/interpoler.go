package interpoler

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

func stringContains(list []string, value string) bool {
	for _, v := range list {
		if v == value {
			return true
		}
	}
	return false
}

func Interpolate(ctx context.Context, command *TokanizedCommand, nodes CommandNodes) *InterpolateResult {
	// If the command parts is empty return an empty command error, this can only happen
	// the first time the interpolate function is called
	if command.Len() == 0 {
		return NewInterpolateResult().
			WithError(newInterpolateError(ErrEmptyCommand))
	}

	// Split the command parts between the current part and the remaining ones
	part, remainingCommand := command.Cut()

	// Try to match the current command part with an entry of the full list of commands
	for _, node := range nodes {
		if node.Name == part {
			validParams := make(map[string]Parameter)
			setParams := make(ParametersWithValue)

			for _, parameter := range node.Parameters {
				validParams[parameter.Name] = parameter

				// set default value as the parameter value if the parameter is optional and has a default value
				// it will be overwritten if the parameter is set in the command
				// or removed is doesn't have a proper value
				if parameter.Optional && parameter.Default != "" {
					setParams[parameter.Name] = &ParameterWithValue{
						Parameter:    parameter,
						Value:        parameter.Default,
						DefaultValue: true,
					}
				}
			}

			// get list of parameters set
			for i := 0; i < len(remainingCommand.Tokens); i += 1 {
				if strings.HasPrefix(remainingCommand.Tokens[i].Value, "--") && i+1 < len(remainingCommand.Tokens) {
					parameterName := remainingCommand.Tokens[i].Value[2:]
					if parameter, ok := validParams[parameterName]; ok {
						value := remainingCommand.Tokens[i+1].Value

						validValue := true
						if parameter.Completer != nil {
							validValues := parameter.Completer(ctx, setParams)
							if !stringContains(validValues, value) {
								validValue = false
							}
						}

						setParams[parameterName] = &ParameterWithValue{
							Parameter:    parameter,
							Value:        value,
							InvalidValue: !validValue,
						}
					}
				}
			}

			var (
				i      int = 0
				result *InterpolateResult
			)
		loop:
			for ; i < len(remainingCommand.Tokens); i += 1 {
				// if it doesn't start with `--` it is not a valid parameter name
				if !strings.HasPrefix(remainingCommand.Tokens[i].Value, "--") {
					result = NewInterpolateResult().
						WithTarget(node).
						WithParameters(setParams).
						WithError(newInterpolateError(ErrInvalidParameterName).WithDetails(remainingCommand.Tokens[i].Value))
					break loop
				}

				// if it doesn't match any of the command parameters, it is not a valid parameter name or it is incomplete
				parameterName := remainingCommand.Tokens[i].Value[2:]
				parameter, ok := validParams[parameterName]
				if !ok {
					result = NewInterpolateResult().
						WithTarget(node).
						WithParameters(setParams).
						WithError(newInterpolateError(ErrInvalidParameterName).WithDetails(remainingCommand.Tokens[i].Value))
					break loop
				}

				// if there is not another token next, the parameter value is missing
				if i+1 >= len(remainingCommand.Tokens) {
					result = NewInterpolateResult().
						WithTarget(node).
						WithTargetParameter(&parameter).
						WithParameters(setParams).
						WithError(newInterpolateError(ErrEmptyParamaterValue))
					break loop
				}

				// point cursor to parameter value
				i += 1

				// if the parameter value doesn't match a valid value, it is incomplete
				if parameter.Completer != nil {
					validValues := parameter.Completer(ctx, setParams)
					if !stringContains(validValues, remainingCommand.Tokens[i].Value) {
						result = NewInterpolateResult().
							WithTarget(node).
							WithTargetParameter(&parameter).
							WithParameters(setParams).
							WithError(newInterpolateError(ErrInvalidParameterValue).WithDetails(remainingCommand.Tokens[i].Value))
						break loop
					}
				}
			}

			if result == nil {
				result = NewInterpolateResult().
					WithTarget(node).
					WithParameters(setParams)
			}

			// if it is a parameter unset error, the cursor needs to have a trailing space to show completions unless the user has already started writting
			if i < len(remainingCommand.Tokens) {
				expectedAutocompletionCursorPos := remainingCommand.Tokens[i].Pos + len(remainingCommand.Tokens[i].Value)
				if errors.Is(result.Error.Error, ErrEmptyParamaterValue) {
					expectedAutocompletionCursorPos += 1
				}
				if expectedAutocompletionCursorPos != remainingCommand.CursorPos {
					result = result.WithMisplacedCursor()
				}
			}

			// check that all required parameters are set
			var missingReqParams []string
			for _, cmdParams := range node.Parameters {
				if !cmdParams.Optional {
					if _, ok := setParams[cmdParams.Name]; !ok {
						missingReqParams = append(missingReqParams, fmt.Sprintf("--%s", cmdParams.Name))
					}
				}
			}
			result = result.WithMissingRequiredParameters(missingReqParams)

			return result
		}
	}

	return NewInterpolateResult().
		WithRemaining(remainingCommand).
		WithError(newInterpolateError(ErrInvalidCommand).WithDetails(command.Tokens[0].Value))
}
