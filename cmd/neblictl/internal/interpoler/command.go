package interpoler

import (
	"context"
	"strconv"

	"github.com/neblic/platform/cmd/neblictl/internal"
)

type ExecutorFunc func(ctx context.Context, funcOptions ParametersWithValue, writer *internal.Writer) error
type CompleterFunc func(ctx context.Context, funcOptions ParametersWithValue) []string

type ParametersWithValue []*ParameterWithValue

func (p *ParametersWithValue) Get(name string) (*ParameterWithValue, bool) {
	for _, parameter := range *p {
		if parameter.Name == name {
			return parameter, true
		}
	}
	return nil, false
}

func (p *ParametersWithValue) GetLast() (*ParameterWithValue, bool) {
	if len(*p) == 0 {
		return nil, false
	}

	return (*p)[len(*p)-1], true
}

type ParameterWithValue struct {
	Parameter
	Value string
}

func (p *ParameterWithValue) AsInt64() (int64, error) {
	return strconv.ParseInt(p.Value, 10, 64)
}

type Parameter struct {
	Name        string
	Description string
	Completer   CompleterFunc
}

type CommandNodes []*CommandNode

type CommandNode struct {
	Name        string
	Description string
	Subcommands CommandNodes
	Parameters  []Parameter
	Executor    ExecutorFunc
}
