package filter

import (
	"regexp"
	"testing"
)

func TestPredicate_Match(t *testing.T) {
	type args struct {
		element string
	}
	tests := []struct {
		name      string
		predicate Predicate
		args      args
		want      bool
	}{
		{
			name:      "Test matching string",
			predicate: NewString("s1"),
			args: args{
				element: "s1",
			},
			want: true,
		},
		{
			name:      "Test not matching string",
			predicate: NewString("s1"),
			args: args{
				element: "random_element",
			},
			want: false,
		},
		{
			name: "Test matching regex",
			predicate: &Regex{
				regex: regexp.MustCompile("^re.*$"),
			},
			args: args{
				element: "re1234",
			},
			want: true,
		},
		{
			name: "Test not matching regex",
			predicate: &Regex{
				regex: regexp.MustCompile("^re.*$"),
			},
			args: args{
				element: "random_element",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.predicate.Match(tt.args.element); got != tt.want {
				t.Errorf("Predicate.Match() = %v, want %v", got, tt.want)
			}
		})
	}
}
