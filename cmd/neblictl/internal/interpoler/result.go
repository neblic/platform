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
	Parents    CommandNodes
	Target     *CommandNode
	Parameters ParametersWithValue
	Remaining  *TokanizedCommand
	Error      *InterpolateError
}

func NewInterpolateResult() *InterpolateResult {
	return &InterpolateResult{
		Parents:    CommandNodes{},
		Target:     nil,
		Parameters: ParametersWithValue{},
		Remaining:  nil,
		Error:      nil,
	}
}

func (r *InterpolateResult) PreappendParent(parent *CommandNode) *InterpolateResult {
	r.Parents = append([]*CommandNode{parent}, r.Parents...)
	return r
}

func (r *InterpolateResult) WithTarget(target *CommandNode) *InterpolateResult {
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

func (r *InterpolateResult) WithRemaining(command *TokanizedCommand) *InterpolateResult {
	r.Remaining = command
	return r
}

func (r *InterpolateResult) WithError(err *InterpolateError) *InterpolateResult {
	r.Error = err
	return r
}
