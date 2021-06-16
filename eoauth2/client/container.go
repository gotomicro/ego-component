package client

import (
	"github.com/gotomicro/ego/core/econf"
	"github.com/gotomicro/ego/core/elog"
)

type Option func(c *Container)

type Container struct {
	Config *Config
	name   string
	logger *elog.Component
}

func DefaultContainer() *Container {
	return &Container{
		Config: DefaultConfig(),
		logger: elog.EgoLogger.With(elog.FieldComponent("component.eoauth2.server")),
	}
}

func Load(key string) *Container {
	c := DefaultContainer()
	if err := econf.UnmarshalKey(key, &c.Config); err != nil {
		c.logger.Panic("parse Config error", elog.FieldErr(err), elog.FieldKey(key))
		return c
	}
	c.logger = c.logger.With(elog.FieldComponentName(key))
	c.name = key
	return c
}

// Build 构建实例
func (c *Container) Build(options ...Option) *Component {
	for _, option := range options {
		option(c)
	}
	return newComponent(c.name, c.Config, c.logger)
}
