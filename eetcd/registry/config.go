package registry

import (
	"time"
)

// Config ...
type Config struct {
	ReadTimeout  time.Duration
	Prefix       string
	ServiceTTL   time.Duration
	OnFailHandle string
}

// DefaultConfig ...
func DefaultConfig() *Config {
	return &Config{
		ReadTimeout: time.Second * 3,
		Prefix:      "ego",
		ServiceTTL:  0,
	}
}
