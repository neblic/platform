package interpoler

type TokanizedCommand struct {
	Parts            []string
	HasTrailingSpace bool
}

func NewTokanizedCommand(parts []string, hasTrailingSpace bool) *TokanizedCommand {
	return &TokanizedCommand{
		Parts:            parts,
		HasTrailingSpace: hasTrailingSpace,
	}
}

func (c *TokanizedCommand) Len() int {
	return len(c.Parts)
}

func (c *TokanizedCommand) Cut() (string, *TokanizedCommand) {
	if c.Len() == 0 {
		panic("tokenized command has size 0")
	}

	// Cut the token in two parts, the head and the remaining parts
	part, remainingParts := c.Parts[0], c.Parts[1:]

	// Create remaining command using remaining parts
	remainingCommand := &TokanizedCommand{
		Parts:            remainingParts,
		HasTrailingSpace: c.HasTrailingSpace,
	}

	return part, remainingCommand
}
