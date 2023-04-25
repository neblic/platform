package interpoler

import (
	"errors"
	"fmt"
)

var (
	ErrEmptyCommand          = errors.New("empty command")
	ErrInvalidCommand        = errors.New("command is invalid")
	ErrInvalidParameterName  = errors.New("parameter name is invalid")
	ErrInvalidParameterValue = errors.New("parameter value is invalid")
	ErrEmptyParamaterValue   = errors.New("parameter value is not set")
)

type InterpolateError struct {
	Error   error
	Details string
}

func (e InterpolateError) DetailedError() error {
	if e.Details != "" {
		return fmt.Errorf("%s %s", e.Details, e.Error)
	}
	return e.Error
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
	Parents                   CommandNodes
	Target                    *CommandNode
	TargetParameter           *Parameter
	Parameters                ParametersWithValue
	Remaining                 *TokanizedCommand
	Error                     *InterpolateError
	MisplacedCursor           bool
	MissingRequiredParameters []string
}

func NewInterpolateResult() *InterpolateResult {
	return &InterpolateResult{
		Parents:                   CommandNodes{},
		Target:                    nil,
		TargetParameter:           nil,
		Parameters:                ParametersWithValue{},
		Remaining:                 nil,
		Error:                     nil,
		MisplacedCursor:           false,
		MissingRequiredParameters: []string{},
	}
}

func (r *InterpolateResult) WithTarget(target *CommandNode) *InterpolateResult {
	r.Target = target
	return r
}

func (r *InterpolateResult) WithParameters(parameters ParametersWithValue) *InterpolateResult {
	r.Parameters = parameters
	return r
}

func (r *InterpolateResult) WithTargetParameter(parameter *Parameter) *InterpolateResult {
	r.TargetParameter = parameter
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

func (r *InterpolateResult) WithMisplacedCursor() *InterpolateResult {
	r.MisplacedCursor = true
	return r
}

func (r *InterpolateResult) WithMissingRequiredParameters(missingReqParams []string) *InterpolateResult {
	r.MissingRequiredParameters = missingReqParams
	return r
}
