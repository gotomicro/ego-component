package ekafka

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

// WithBrokers 注入brokers配置
func WithBrokers(brokers ...string) Option {
	return func(c *Container) {
		c.config.Brokers = brokers
	}
}
