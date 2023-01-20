package filter

import (
	"fmt"
	"regexp"
)

type Predicates []Predicate

type Predicate interface {
	Match(element string) bool
}

type Regex struct {
	regex *regexp.Regexp
}

func NewRegex(expr string) (*Regex, error) {
	regex, err := regexp.Compile(expr)
	if err != nil {
		return nil, fmt.Errorf("error compiling regex: %w", err)
	}

	return &Regex{
		regex: regex,
	}, nil
}

func (r *Regex) Match(element string) bool {
	return r.regex.MatchString(element)
}

type String struct {
	str string
}

func NewString(str string) *String {
	return &String{
		str: str,
	}
}

func (s *String) Match(element string) bool {
	return s.str == element
}
