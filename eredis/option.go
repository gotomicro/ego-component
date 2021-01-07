package eredis

import (
	"github.com/gotomicro/ego/core/elog"
)

// WithStub 注入stub配置
func WithStub() Option {
	return func(c *Container) {
		if c.config.Addr == "" && len(c.config.Addrs) == 0 {
			c.logger.Panic("no address in redis config", elog.FieldName(c.name))
		}
		if c.config.Addr != "" {
			c.config.Addrs = []string{c.config.Addr}
		}
		c.config.Mode = StubMode
	}
}

// WithCluster 注入Cluster配置
func WithCluster() Option {
	return func(c *Container) {
		c.config.Mode = ClusterMode
	}
}

// WithInterceptor 注入拦截器
func WithInterceptor(interceptors ...Interceptor) Option {
	return func(c *Container) {
		if c.config.interceptors == nil {
			c.config.interceptors = make([]Interceptor, 0)
		}
		c.config.interceptors = append(c.config.interceptors, interceptors...)
	}
}
