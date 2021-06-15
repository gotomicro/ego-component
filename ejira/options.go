package ejira

func WithAddr(addr string) Option {
	return func(c *Container) {
		c.config.Addr = addr
	}
}

func WithUsername(username string) Option {
	return func(c *Container) {
		c.config.Username = username
	}
}

func WithPassword(password string) Option {
	return func(c *Container) {
		c.config.Password = password
	}
}
