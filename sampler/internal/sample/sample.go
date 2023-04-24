package sample

import (
	"time"

	"github.com/neblic/platform/controlplane/data"
)

// Match identifies what determined that the sample should have been exported
type Match struct {
	StreamUID data.SamplerStreamRuleUID
}

// Sample contains a data sample.
type Sample struct {
	Ts      time.Time
	Data    map[string]any
	Matches []Match
}

// SamplerSamples contains a group of samples that originate from the same sampler
type SamplerSamples struct {
	Samples []Sample
}

// ResourceSamples contains a group of samples that originate from the same resource
// e.g. operator, service, PU
type ResourceSamples struct {
	ResourceName    string
	SamplerName     string
	SamplersSamples []SamplerSamples
}
