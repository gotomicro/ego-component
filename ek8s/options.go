package ek8s

// Option 可选项
type Option func(c *Container)

func WithAddr(addr string) Option {
	return func(c *Container) {
		c.config.Addr = addr
	}
}

func WithDebug(debug bool) Option {
	return func(c *Container) {
		c.config.Debug = debug
	}
}

func WithToken(token string) Option {
	return func(c *Container) {
		c.config.Token = token
	}
}

func WithNamespaces(namespaces []string) Option {
	return func(c *Container) {
		c.config.Namespaces = namespaces
	}
}

func WithDeploymentPrefix(deploymentPrefix string) Option {
	return func(c *Container) {
		c.config.DeploymentPrefix = deploymentPrefix
	}
}

func WithTLSClientConfigInsecure(insecure bool) Option {
	return func(c *Container) {
		c.config.TLSClientConfigInsecure = insecure
	}
}
