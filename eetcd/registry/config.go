package registry

import (
	"time"
)

// Config ...
type Config struct {
	Scheme       string
	ReadTimeout  time.Duration
	Prefix       string
	ServiceTTL   time.Duration
	OnFailHandle string
}

const (
	defaultScheme = "etcd"
)

// DefaultConfig ...
func DefaultConfig() *Config {
	return &Config{
		Scheme:      defaultScheme,
		ReadTimeout: time.Second * 3,
		Prefix:      "ego",
		ServiceTTL:  0,
	}
}
