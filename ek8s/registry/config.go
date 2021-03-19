package registry

// Config ...
type Config struct {
	Kind         string
	OnFailHandle string
}

// DefaultConfig ...
func DefaultConfig() *Config {
	return &Config{
		Kind: "Pod",
	}
}
