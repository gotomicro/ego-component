package eetcd

type Option func(c *Container)

func WithAddrs(addrs []string) Option {
	return func(c *Container) {
		c.config.Addrs = addrs
	}
}

func WithCertFile(certFile string) Option {
	return func(c *Container) {
		c.config.CertFile = certFile
	}
}

func WithKeyFile(keyFile string) Option {
	return func(c *Container) {
		c.config.KeyFile = keyFile
	}
}

func WithCaCert(caCert string) Option {
	return func(c *Container) {
		c.config.CaCert = caCert
	}
}

func WithEnableBasicAuth(enable bool) Option {
	return func(c *Container) {
		c.config.EnableBasicAuth = enable
	}
}

func WithUserName(userName string) Option {
	return func(c *Container) {
		c.config.UserName = userName
	}
}

func WithPassword(password string) Option {
	return func(c *Container) {
		c.config.Password = password
	}
}

func WithEnableSecure(secure bool) Option {
	return func(c *Container) {
		c.config.EnableSecure = secure
	}
}
