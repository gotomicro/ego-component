package ecronlock

// Config ...
type Config struct {
	OnFailHandle string
}

// DefaultConfig ...
func DefaultConfig() *Config {
	return &Config{}
}
