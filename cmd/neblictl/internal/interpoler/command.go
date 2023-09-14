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

func (p *ParameterWithValue) AsBool() (bool, error) {
	return strconv.ParseBool(p.Value)
}

func (p *ParameterWithValue) AsInt64() (int64, error) {
	return strconv.ParseInt(p.Value, 10, 64)
}

func (p *ParameterWithValue) AsInt32() (int32, error) {
	val, err := strconv.ParseInt(p.Value, 10, 32)
	if err != nil {
		return 0, err
	}

	return int32(val), nil
}

type Parameter struct {
	Name        string
	Description string
	Completer   CompleterFunc
	Optional    bool
	Default     string
	Filter      bool
}

type CommandNodes []*CommandNode

type CommandNode struct {
	Name                string
	Description         string
	ExtendedDescription string
	Parameters          []Parameter
	Executor            ExecutorFunc
}
