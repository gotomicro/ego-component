package registry

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
