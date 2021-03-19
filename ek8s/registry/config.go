package registry

import (
	"github.com/gotomicro/ego-component/ek8s"
)

// Config ...
type Config struct {
	Kind         string
	OnFailHandle string
}

// DefaultConfig ...
func DefaultConfig() *Config {
	return &Config{
		Kind: ek8s.KindEndpoints,
	}
}
