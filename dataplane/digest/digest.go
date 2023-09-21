package digest

import (
	"github.com/neblic/platform/controlplane/control"
	"github.com/neblic/platform/internal/pkg/data"
)

type digest interface {
	isDigest()
}

type Digest interface {
	AddSampleData(*data.Data) error
	JSON() ([]byte, error)
	Reset()
	String() string
	SampleType() control.SampleType
}
