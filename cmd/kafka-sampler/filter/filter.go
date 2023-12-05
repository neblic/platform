package filter

import "errors"

type Filter struct {
	predicates []Predicate
	evalFunc   func(predicates []Predicate, element string) bool
}

func trueEvalFunc(predicates []Predicate, element string) bool {
	return true
}

func allowlistEvalFunc(predicates []Predicate, element string) bool {
	for _, predicate := range predicates {
		if predicate.Match(element) {
			// A matching allow predicate was found, accept element
			return true
		}
	}

	// No matching allow predicate, reject element.
	return false
}

func denylistEvalFunc(predicates []Predicate, element string) bool {
	for _, predicate := range predicates {
		if predicate.Match(element) {
			// A matching deny predicate was found, reject element
			return false
		}
	}

	// No matching deny predicate, accept element.
	return true
}

func New(config *Config) (*Filter, error) {
	if len(config.Allowlist) > 0 && len(config.Denylist) > 0 {
		return nil, errors.New("allowlist and denylist at the same time is not supported. Specify one of the two")
	}

	var predicates []Predicate
	var evalFunc func(predicates []Predicate, element string) bool
	if len(config.Allowlist) > 0 {
		predicates = config.Allowlist
		evalFunc = allowlistEvalFunc
	} else if len(config.Denylist) > 0 {
		predicates = config.Denylist
		evalFunc = denylistEvalFunc
	} else {
		evalFunc = trueEvalFunc
	}

	return &Filter{
		predicates: predicates,
		evalFunc:   evalFunc,
	}, nil
}

func (f *Filter) Evaluate(element string) bool {
	return f.evalFunc(f.predicates, element)
}

func (f *Filter) EvaluateList(elements []string) []string {
	evaluatedList := make([]string, 0, len(elements))
	for _, element := range elements {
		if f.Evaluate(element) {
			evaluatedList = append(evaluatedList, element)
		}
	}
	return evaluatedList
}
