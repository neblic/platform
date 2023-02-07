package interpoler

func Interpolate(command *TokanizedCommand, nodes CommandNodes) *InterpolateResult {
	// If the command parts is empty return an empty command error, this can only happen
	// the first time the interpolate function is called
	if command.Len() == 0 {
		return NewInterpolateResult().
			WithError(newInterpolateError(ErrEmptyCommand))
	}

	// Split the command parts between the current part and the remaining ones
	part, remainingCommand := command.Cut()

	// Try to match the current command part with and entry of the full list of commands
	for _, node := range nodes {
		if node.Name == part {

			// A command was matched, process the command parameters if exist
			parameters := ParametersWithValue{}
			for _, parameter := range node.Parameters {
				// In case a parameter is expected but the input does not contain any other part, return an error indicating
				// a parameter was expected but not provided.
				if remainingCommand.Len() == 0 {
					parameters = append(parameters, &ParameterWithValue{Parameter: parameter, Value: ""})

					return NewInterpolateResult().
						WithTarget(node).
						WithParameters(parameters).
						WithRemaining(remainingCommand).
						WithError(newInterpolateError(ErrParameterNotProvided).WithDetails(parameter.Name))
				}

				// Split the remaining parts to get the parameter value and the remaining parts
				part, remainingCommand = remainingCommand.Cut()

				// Append the parameter to the function options
				parameters = append(parameters, &ParameterWithValue{Parameter: parameter, Value: part})

				// If a completer function exists, use it to check the parameter value is valid. If it's not valid throw and error
				if parameter.Completer != nil {
					validParameterValue := false
					for _, entry := range parameter.Completer(parameters) {
						if part == entry {
							validParameterValue = true
							break
						}
					}
					if !validParameterValue {
						return NewInterpolateResult().
							WithTarget(node).
							WithParameters(parameters).
							WithRemaining(remainingCommand).
							WithError(newInterpolateError(ErrInvalidParameter).WithDetails(part))
					}
				}
			}

			// Basic case
			if remainingCommand.Len() == 0 {
				result := NewInterpolateResult().
					WithTarget(node).
					WithParameters(parameters)

				if len(result.Target.Subcommands) > 0 {
					result.WithError(newInterpolateError(ErrPartialCommand))
				}

				return result
			}

			// Recursive case
			result := Interpolate(remainingCommand, node.Subcommands).
				PreappendParameters(parameters)

			// In case of the result not having a target, set the current command as target.
			// Otherwise add it to the parents
			if result.Target == nil {
				result.Target = node
			} else {
				result.PreappendParent(node)
			}

			return result
		}
	}

	return NewInterpolateResult().
		WithRemaining(remainingCommand).
		WithError(newInterpolateError(ErrInvalidCommand).WithDetails(command.Parts[0]))
}
