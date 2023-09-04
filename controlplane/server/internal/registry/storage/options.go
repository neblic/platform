package storage

type Type int

const (
	NopType = Type(iota)
	DiskType
)

type Options struct {
	// Type contains the backend used to store data
	Type Type
	// Path contains the root folder where the data will be stored
	Path string
}

func NewOptionsDefault() *Options {
	return &Options{
		Type: NopType,
		Path: "",
	}
}
