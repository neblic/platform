package registry

type StorageType int

const (
	NoopStorage = StorageType(iota)
	DiskStorage
)

type Options struct {
	// StorageType contains the backend used to store data
	StorageType StorageType
	// Path contains the root folder where the data will be stored
	Path string
}

func NewOptionsDefault() *Options {
	return &Options{
		StorageType: NoopStorage,
		Path:        "",
	}
}
