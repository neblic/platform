package otlp

type AuthBearerOptions struct {
	Token string
}

type AuthOptions struct {
	Type   string
	Bearer AuthBearerOptions
}

type Options struct {
	TLSEnable bool
	Auth      AuthOptions
}

func newDefaultOptions() *Options {
	return &Options{
		TLSEnable: false,
	}
}
