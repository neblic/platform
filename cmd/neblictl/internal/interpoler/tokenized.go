package interpoler

type TokanizedCommand struct {
	Tokens           []Token
	HasTrailingSpace bool
	CursorPos        int
}

type Token struct {
	Value string
	Pos   int
}

func NewTokanizedCommand(tokens []Token, hasTrailingSpace bool, cursorPos int) *TokanizedCommand {
	return &TokanizedCommand{
		Tokens:           tokens,
		HasTrailingSpace: hasTrailingSpace,
		CursorPos:        cursorPos,
	}
}

func (c *TokanizedCommand) Len() int {
	return len(c.Tokens)
}

func (c *TokanizedCommand) Cut() (string, *TokanizedCommand) {
	if c.Len() == 0 {
		panic("tokenized command has size 0")
	}

	// Cut the token in two parts, the head and the remaining parts
	token, remainingTokens := c.Tokens[0], c.Tokens[1:]

	// Create remaining command using remaining parts
	remainingCommand := &TokanizedCommand{
		Tokens:           remainingTokens,
		HasTrailingSpace: c.HasTrailingSpace,
		CursorPos:        c.CursorPos - token.Pos,
	}

	return token.Value, remainingCommand
}
