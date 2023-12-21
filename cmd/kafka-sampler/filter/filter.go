package filter

import "errors"

type Filter struct {
	predicate Predicate
	evalFunc  func(predicate Predicate, element string) bool
}

func trueEvalFunc(predicate Predicate, element string) bool {
	return true
}

func allowEvalFunc(predicate Predicate, element string) bool {
	return predicate.Match(element)
}

func denyEvalFunc(predicate Predicate, element string) bool {
	return !predicate.Match(element)
}

func New(config *Config) (*Filter, error) {
	if config.Allow != nil && config.Deny != nil {
		return nil, errors.New("allow and deny at the same time is not supported. Specify one of the two")
	}

	var predicate Predicate
	var evalFunc func(predicate Predicate, element string) bool
	if config.Allow != nil {
		predicate = config.Allow
		evalFunc = allowEvalFunc
	} else if config.Deny != nil {
		predicate = config.Deny
		evalFunc = denyEvalFunc
	} else {
		evalFunc = trueEvalFunc
	}

	return &Filter{
		predicate: predicate,
		evalFunc:  evalFunc,
	}, nil
}

func (f *Filter) Evaluate(element string) bool {
	return f.evalFunc(f.predicate, element)
}

func (f *Filter) EvaluateList(elements []string) ([]string, []string) {
	allowedList := make([]string, 0, len(elements))
	deniedList := make([]string, 0, len(elements))
	for _, element := range elements {
		if f.Evaluate(element) {
			allowedList = append(allowedList, element)
		} else {
			deniedList = append(deniedList, element)
		}
	}
	return allowedList, deniedList
}
