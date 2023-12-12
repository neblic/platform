package filter

type Config struct {
	Allow Predicate
	Deny  Predicate
}

func NewConfig() *Config {
	return &Config{
		Allow: nil,
		Deny:  nil,
	}
}
