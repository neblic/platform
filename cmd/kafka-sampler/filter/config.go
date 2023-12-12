package filter

type Config struct {
	Allowlist Predicates
	Denylist  Predicates
}

func NewConfig() *Config {
	return &Config{
		Allowlist: Predicates{},
		Denylist:  Predicates{},
	}
}
