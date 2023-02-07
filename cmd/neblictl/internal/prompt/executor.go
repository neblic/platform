package internal

import (
	"bytes"
	"fmt"

	"github.com/neblic/platform/cmd/neblictl/internal"
	"github.com/neblic/platform/cmd/neblictl/internal/interpoler"
)

var (
	ErrEmptyComand = fmt.Errorf("empty command")
)

func generateGlobalHelp(targets interpoler.CommandNodes, writer *internal.Writer) {
	writer.WriteString("Usage\n\t[command]\n\nCommands\n")
	for _, target := range targets {
		writer.WriteStringf("\t%s: %s\n", target.Name, target.Description)
	}
}

func generateTargetHelp(parents interpoler.CommandNodes, target *interpoler.CommandNode, err *interpoler.InterpolateError, writer *internal.Writer) {
	// Build full command
	commandBuffer := bytes.Buffer{}
	for i, parent := range parents {
		if i > 0 {
			commandBuffer.WriteString(" ")
		}
		commandBuffer.WriteString(parent.Name)
	}
	if commandBuffer.Len() > 0 {
		commandBuffer.WriteString(" ")
	}
	commandBuffer.WriteString(target.Name)
	for _, parameter := range target.Parameters {
		commandBuffer.WriteString(" [")
		commandBuffer.WriteString(parameter.Name)
		commandBuffer.WriteString("]")
	}
	if len(target.Subcommands) > 0 {
		commandBuffer.WriteString(" [subcommand]")
	}

	// Build help
	if err != nil {
		writer.WriteStringf("Error: %v\n\n", err.DetailedError())
	}

	// Add description
	writer.WriteStringf("%s\n\n", target.Description)

	// Add usage section
	writer.WriteString("Usage\n")
	writer.WriteStringf("\t%s\n", commandBuffer.String())

	// Add parameters section
	if len(target.Parameters) > 0 {
		writer.WriteString("\nParameters\n")
		for _, parameter := range target.Parameters {
			writer.WriteStringf("\t%s: %s\n", parameter.Name, parameter.Description)
		}
	}

	// Add subcommands section
	if len(target.Subcommands) > 0 {
		writer.WriteString("\nSubcommands\n")
		for _, subcommand := range target.Subcommands {
			writer.WriteStringf("\t%s: %s\n", subcommand.Name, subcommand.Description)
		}
	}
}

func Execute(nodes interpoler.CommandNodes, command *interpoler.TokanizedCommand, writer *internal.Writer) error {
	// Check the command parts is not empty
	if command.Len() == 0 {
		return ErrEmptyComand
	}

	isHelp := false

	// Check if a help command was received
	if command.Parts[0] == "help" {
		// Strip help part from the command and force print help
		command.Parts = command.Parts[1:]

		// If the command just contains help, print global help
		if command.Len() == 0 {
			generateGlobalHelp(nodes, writer)
			return nil
		}

		isHelp = true
	}

	// Find executor
	result := interpoler.Interpolate(command, nodes)
	if result.Error != nil {
		// In case of processing a help command, do not show the error
		err := result.Error
		if isHelp {
			err = nil
		}

		switch result.Error.Error {
		case interpoler.ErrInvalidCommand:
			if result.Target != nil {
				generateTargetHelp(result.Parents, result.Target, err, writer)
				return nil
			}
			generateGlobalHelp(nodes, writer)
			return nil
		case interpoler.ErrParameterNotProvided, interpoler.ErrPartialCommand, interpoler.ErrInvalidParameter:
			generateTargetHelp(result.Parents, result.Target, err, writer)
			return nil
		}
	}

	// Force showing help
	if isHelp {
		generateTargetHelp(result.Parents, result.Target, nil, writer)
		return nil
	}

	// Perfrom action
	return result.Target.Executor(result.Parameters, writer)
}
