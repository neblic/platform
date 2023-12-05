package filter

import "time"

type Config struct {
	RefreshPeriod time.Duration
	Allowlist     Predicates
	Denylist      Predicates
}

func NewConfig() *Config {
	return &Config{
		RefreshPeriod: time.Minute,
		Allowlist:     Predicates{},
		Denylist:      Predicates{},
	}
}
