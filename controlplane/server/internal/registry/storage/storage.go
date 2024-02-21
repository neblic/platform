package storage

import (
	"fmt"
)

var (
	ErrUnknownSampler = fmt.Errorf("unknown sampler")
)

type Storage interface {
	GetSampler(resource string, sampler string) (SamplerEntry, error)
	RangeSamplers(func(entry SamplerEntry)) error
	SetSampler(entry SamplerEntry) error
	DeleteSampler(resource string, sampler string) error
}
