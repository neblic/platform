package digest

import "github.com/neblic/platform/sampler/internal/sample"

type digest interface {
	isDigest()
}

type Digest interface {
	AddSampleData(*sample.Data) error
	JSON() ([]byte, error)
	Reset()
	String() string
}
