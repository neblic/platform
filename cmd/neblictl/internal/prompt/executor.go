package internal

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/neblic/platform/cmd/neblictl/internal"
	"github.com/neblic/platform/cmd/neblictl/internal/interpoler"
)

var (
	ErrEmptyComand = fmt.Errorf("empty command")
	namespaceHelp  = map[string]string{
		"resources": "A resource identifies a service or a group of services that are part of the same logical application.",
		"samplers":  "A sampler is a component that collects samples from a resource.",
		"streams":   "A stream is a sequence of samples collected from a resource by a sampler.",
	}
)

func generateGlobalHelp(targets interpoler.CommandNodes, writer *internal.Writer) {
	currentNamespace := ""

	writer.WriteString("Usage:\n\t[namespace:command]\n\nCommands:")
	for _, target := range targets {
		namespaceCmd := strings.SplitN(target.Name, ":", 2)
		if len(namespaceCmd) != 2 {
			panic(fmt.Sprintf("Invalid command name: %s ", target.Name))
		}

		if currentNamespace != namespaceCmd[0] {
			currentNamespace = namespaceCmd[0]
			writer.WriteStringf("\n\t %s: %s\n", currentNamespace, namespaceHelp[currentNamespace])
		}

		writer.WriteStringf("\t\to %s: %s\n", namespaceCmd[1], target.Description)
	}
}

func generateTargetHelp(parents interpoler.CommandNodes, target *interpoler.CommandNode, errMsg error, writer *internal.Writer) {
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
	commandBuffer.WriteString(" ")
	for _, parameter := range target.Parameters {
		if parameter.Optional {
			commandBuffer.WriteString("[")
		}
		commandBuffer.WriteString(fmt.Sprintf("--%s <>", parameter.Name))
		if parameter.Optional {
			commandBuffer.WriteString("]")
		}
		commandBuffer.WriteString(" ")
	}

	// Build help
	if errMsg != nil {
		writer.WriteStringf("Error: %v\n\n", errMsg)
	}

	// Add description
	writer.WriteStringf("%s\n\n", target.Description)
	if len(target.ExtendedDescription) > 0 {
		writer.WriteStringf("%s\n\n", target.ExtendedDescription)
	}

	// Add usage section
	writer.WriteString("Usage\n")
	writer.WriteStringf("\t%s\n", commandBuffer.String())

	// Add parameters section
	if len(target.Parameters) > 0 {
		writer.WriteString("\nParameters\n")
		for _, parameter := range target.Parameters {
			writer.WriteStringf("\t%s: %s", parameter.Name, parameter.Description)
			if parameter.Optional {
				writer.WriteString(" (optional)")
			}
			writer.WriteString("\n")
		}
	}
}

func Execute(ctx context.Context, nodes interpoler.CommandNodes, command *interpoler.TokanizedCommand, writer *internal.Writer) error {
	// Check the command parts is not empty
	if command.Len() == 0 {
		return ErrEmptyComand
	}

	isHelp := false

	// Check if a help command was received
	if command.Tokens[0].Value == "help" {
		// Strip help part from the command and force print help
		command.Tokens = command.Tokens[1:]

		// If the command just contains help, print global help
		if command.Len() == 0 {
			generateGlobalHelp(nodes, writer)
			return nil
		}

		isHelp = true
	}

	// Find executor
	result := interpoler.Interpolate(ctx, command, nodes)

	if len(result.MissingRequiredParameters) > 0 {
		generateTargetHelp(result.Parents, result.Target, fmt.Errorf("missing required parameters: %s", strings.Join(result.MissingRequiredParameters, " ")), writer)
		return nil
	}

	if result.Error != nil {
		// In case of processing a help command, do not show the error
		err := result.Error
		if isHelp {
			err = nil
		}

		switch result.Error.Error {
		case interpoler.ErrInvalidCommand:
			fallthrough
		case interpoler.ErrInvalidParameterName:
			if result.Target != nil {
				generateTargetHelp(result.Parents, result.Target, err.DetailedError(), writer)
				return nil
			}
			generateGlobalHelp(nodes, writer)
			return nil
		case interpoler.ErrInvalidParameterValue:
			fallthrough
		case interpoler.ErrEmptyParamaterValue:
			generateTargetHelp(result.Parents, result.Target, err.DetailedError(), writer)
			return nil
		}
	}

	// Force showing help
	if isHelp {
		generateTargetHelp(result.Parents, result.Target, nil, writer)
		return nil
	}

	// Perfrom action
	err := result.Target.Executor(ctx, result.Parameters, writer)
	return err
}
