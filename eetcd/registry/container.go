package registry

import (
	"github.com/gotomicro/ego-component/eetcd"
	"github.com/gotomicro/ego/core/econf"
	"github.com/gotomicro/ego/core/elog"
)

type Option func(c *Container)

type Container struct {
	config *Config
	name   string
	logger *elog.Component
	client *eetcd.Component
}

func DefaultContainer() *Container {
	return &Container{
		config: DefaultConfig(),
		logger: elog.EgoLogger.With(elog.FieldComponent(eetcd.PackageName)),
	}
}

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

func WithClientEtcd(etcdClient *eetcd.Component) Option {
	return func(c *Container) {
		c.client = etcdClient
	}
}

// Build ...
func (c *Container) Build(options ...Option) *Component {
	for _, option := range options {
		option(c)
	}
	if c.client == nil {
		if c.config.OnFailHandle == "panic" {
			c.logger.Panic("client etcd nil", elog.FieldKey("use WithClientEtcd method"))
		} else {
			c.logger.Error("client etcd nil", elog.FieldKey("use WithClientEtcd method"))
		}
	}
	return newComponent(c.name, c.config, c.logger, c.client)
}
