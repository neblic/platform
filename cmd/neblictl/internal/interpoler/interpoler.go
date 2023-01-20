package interpoler

func Interpolate(commandParts []string, commands []*Command) *InterpolateResult {
	// If the command parts is empty return an empty command error, this can only happen
	// the first time the interpolate function is called
	if len(commandParts) == 0 {
		return NewInterpolateResult().
			WithError(newInterpolateError(ErrEmptyCommand))
	}

	// Split the command parts between the current part and the remaining ones
	part, remainingParts := commandParts[0], commandParts[1:]

	// Try to match the current command part with and entry of the full list of commands
	for _, cmd := range commands {
		if cmd.Name == part {

			// A command was matched, process the command parameters if exist
			parameters := ParametersWithValue{}
			for _, parameter := range cmd.Parameters {
				// In case a parameter is expected but the input does not contain any other part, return an error indicating
				// a parameter was expected but not provided.
				if len(remainingParts) == 0 {
					parameters = append(parameters, &ParameterWithValue{Parameter: parameter, Value: ""})

					return NewInterpolateResult().
						WithTarget(cmd).
						WithParameters(parameters).
						WithError(newInterpolateError(ErrParameterNotProvided).WithDetails(parameter.Name))
				}

				// Split the remaining parts to get the parameter value and the remaining parts
				part, remainingParts = remainingParts[0], remainingParts[1:]

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
							WithTarget(cmd).
							WithParameters(parameters).
							WithError(newInterpolateError(ErrInvalidParameter).WithDetails(part))
					}
				}
			}

			// Basic case
			if len(remainingParts) == 0 {
				result := NewInterpolateResult().
					WithTarget(cmd).
					WithParameters(parameters)

				if len(result.Target.Subcommands) > 0 {
					result.WithError(newInterpolateError(ErrPartialCommand))
				}

				return result
			}

			// Recursive case
			result := Interpolate(remainingParts, cmd.Subcommands).
				PreappendParameters(parameters)

			// In case of the result not having a target, set the current command as target.
			// Otherwise add it to the parents
			if result.Target == nil {
				result.Target = cmd
			} else {
				result.PreappendParent(cmd)
			}

			return result
		}
	}

	return NewInterpolateResult().
		WithError(newInterpolateError(ErrInvalidCommand).WithDetails(commandParts[0]))
}
