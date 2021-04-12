package registry

import (
	"github.com/gotomicro/ego-component/ek8s"
)

// Config ...
type Config struct {
	Scheme       string
	Kind         string
	OnFailHandle string
}

const (
	defaultScheme = "k8s"
)

// DefaultConfig ...
func DefaultConfig() *Config {
	return &Config{
		Kind:   ek8s.KindEndpoints,
		Scheme: defaultScheme,
	}
}
