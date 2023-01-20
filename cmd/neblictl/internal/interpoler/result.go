package interpoler

import (
	"errors"
	"fmt"
)

var (
	ErrEmptyCommand         = errors.New("empty command")
	ErrInvalidCommand       = errors.New("command is invalid")
	ErrParameterNotProvided = errors.New("parameter not provided")
	ErrInvalidParameter     = errors.New("parameter is invalid")
	ErrPartialCommand       = errors.New("partial command")
)

type InterpolateError struct {
	Error   error
	Details string
}

func (e InterpolateError) DetailedError() string {
	if e.Details != "" {
		return fmt.Sprintf("%s %s", e.Details, e.Error)
	}
	return e.Error.Error()
}

func newInterpolateError(err error) *InterpolateError {
	return &InterpolateError{
		Error:   err,
		Details: "",
	}
}

func (e *InterpolateError) WithDetails(details string) *InterpolateError {
	e.Details = details
	return e
}

type InterpolateResult struct {
	Parents    []*Command
	Target     *Command
	Parameters ParametersWithValue
	Error      *InterpolateError
}

func NewInterpolateResult() *InterpolateResult {
	return &InterpolateResult{
		Parents:    []*Command{},
		Target:     nil,
		Parameters: ParametersWithValue{},
		Error:      nil,
	}
}

func (r *InterpolateResult) PreappendParent(parent *Command) *InterpolateResult {
	r.Parents = append([]*Command{parent}, r.Parents...)
	return r
}

func (r *InterpolateResult) WithTarget(target *Command) *InterpolateResult {
	r.Target = target
	return r
}

func (r *InterpolateResult) WithParameters(parameters ParametersWithValue) *InterpolateResult {
	r.Parameters = parameters
	return r
}

func (r *InterpolateResult) PreappendParameters(parameters ParametersWithValue) *InterpolateResult {
	r.Parameters = append(parameters, r.Parameters...)
	return r
}

func (r *InterpolateResult) WithError(err *InterpolateError) *InterpolateResult {
	r.Error = err
	return r
}
