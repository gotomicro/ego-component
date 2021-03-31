package ecronlock

import (
	"github.com/gotomicro/ego/core/eapp"
)

// Config ...
type Config struct {
	Prefix string // 默认 "ecronlock:{appName}:"
}

// DefaultConfig ...
func DefaultConfig() *Config {
	return &Config{
		Prefix: "ecronlock:" + eapp.Name() + ":",
	}
}
