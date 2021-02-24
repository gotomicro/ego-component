package eredis

// WithStub set mode to "stub"
func WithStub() Option {
	return func(c *Container) {
		c.config.Mode = StubMode
	}
}

// WithStub set mode to "cluster"
func WithCluster() Option {
	return func(c *Container) {
		c.config.Mode = ClusterMode
	}
}

// WithStub set mode to "sentinel"
func WithSentinel() Option {
	return func(c *Container) {
		c.config.Mode = SentinelMode
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
