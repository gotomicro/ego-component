package registry

import (
	"github.com/gotomicro/ego-component/ek8s"
)

type Option func(c *Container)

func WithScheme(scheme string) Option {
	return func(c *Container) {
		c.config.Scheme = scheme
	}
}

func WithKind(kind string) Option {
	return func(c *Container) {
		c.config.Kind = kind
	}
}

func WithOnFailHandle(onFileHandle string) Option {
	return func(c *Container) {
		c.config.OnFailHandle = onFileHandle
	}
}

var WithClient = WithClientK8s

// Deprecated: 使用WithClient
func WithClientK8s(k8s *ek8s.Component) Option {
	return func(c *Container) {
		c.client = k8s
	}
}
