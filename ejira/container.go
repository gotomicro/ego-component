package ejira

import (
	"github.com/gotomicro/ego/core/econf"
	"github.com/gotomicro/ego/core/elog"
)

// Option ...
type Option func(c *Container)

// Container ...
type Container struct {
	config *config
	name   string
	logger *elog.Component
}

// DefaultContainer ...
func DefaultContainer() *Container {
	return &Container{
		config: DefaultConfig(),
		logger: elog.EgoLogger.With(elog.FieldComponent(PackageName)),
	}
}

// Load ...
func Load(key string) *Container {
	c := DefaultContainer()
	if err := econf.UnmarshalKey(key, &c.config); err != nil {
		c.logger.Panic("parse config error", elog.FieldErr(err), elog.FieldKey(key))
		return c
	}
	c.logger = c.logger.With(elog.FieldComponentName(key))
	c.name = key
	return c
}

// Build ...
func (c *Container) Build(options ...Option) *Component {
	for _, option := range options {
		option(c)
	}
	return newComponent(c.name, c.config, c.logger)
}
