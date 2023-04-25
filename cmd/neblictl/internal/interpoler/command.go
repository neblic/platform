package interpoler

import (
	"context"
	"strconv"

	"github.com/neblic/platform/cmd/neblictl/internal"
)

type ExecutorFunc func(ctx context.Context, funcOptions ParametersWithValue, writer *internal.Writer) error
type CompleterFunc func(ctx context.Context, funcOptions ParametersWithValue) []string

type ParametersWithValue map[string]*ParameterWithValue

type ParameterWithValue struct {
	Parameter
	Value        string
	DefaultValue bool
	InvalidValue bool
}

func (p *ParametersWithValue) Get(name string) (*ParameterWithValue, bool) {
	parameter, ok := (*p)[name]
	return parameter, ok
}

func (p *ParametersWithValue) IsSet(name string) bool {
	parameter, ok := p.Get(name)
	return ok && !parameter.DefaultValue
}

func (p *ParameterWithValue) AsInt64() (int64, error) {
	return strconv.ParseInt(p.Value, 10, 64)
}

type Parameter struct {
	Name        string
	Description string
	Completer   CompleterFunc
	Optional    bool
	Default     string
	DoNotFilter bool
}

type CommandNodes []*CommandNode

type CommandNode struct {
	Name                string
	Description         string
	ExtendedDescription string
	Parameters          []Parameter
	Executor            ExecutorFunc
}
