package emongo

// WithInterceptor 注入拦截器
func WithInterceptor(interceptors ...Interceptor) Option {
	return func(c *Container) {
		if c.config.interceptors == nil {
			c.config.interceptors = make([]Interceptor, 0)
		}
		c.config.interceptors = append(c.config.interceptors, interceptors...)
	}
}

// WithDebug 注入Debug配置
func WithDebug(debug bool) Option {
	return func(c *Container) {
		c.config.Debug = debug
	}
}

// WithDSN 注入DSN配置
func WithDSN(dsn string) Option {
	return func(c *Container) {
		c.config.DSN = dsn
	}
}
