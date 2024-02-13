package sample

type Options struct {
	Key  string
	Size int
}

type Option interface {
	apply(*Options)
}

type funcOption struct {
	f func(*Options)
}

func (fco *funcOption) apply(co *Options) {
	fco.f(co)
}

func newFuncOption(f func(*Options)) *funcOption {
	return &funcOption{
		f: f,
	}
}

// WithKey sets the key for the sample.
// This key is used for exxample as the determinant in the deterministic sampler or to partition keyed streams.
func WithKey(key string) Option {
	return newFuncOption(func(o *Options) {
		o.Key = key
	})
}

// WithSize sets the sample size, ideally in bytes but any other unit can be used as long as it is consistent with the set limit.
// It is used to discard samples that are considered too large to be processed.
// The maximum allowed size can be set when creating a stream.
func WithSize(size int) Option {
	return newFuncOption(func(o *Options) {
		o.Size = size
	})
}
