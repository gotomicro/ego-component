package etoken

import (
	"github.com/go-redis/redis/v8"
	"github.com/gotomicro/ego-component/eredis"
	"github.com/gotomicro/ego/core/econf"
	"github.com/gotomicro/ego/core/elog"
)

type Container struct {
	config *config
	client *redis.Client
	logger *elog.Component
}

func DefaultContainer() *Container {
	return &Container{
		config: DefaultConfig(),
		logger: elog.EgoLogger.With(elog.FieldComponent(PackageName)),
	}
}

// Load ...
func Load(key string) *Container {
	var container = DefaultContainer()
	if err := econf.UnmarshalKey(key, &container.config); err != nil {
		container.logger.Panic("parse etoken config panic",
			elog.FieldErr(err), elog.FieldKey(key), elog.FieldValueAny(container.config))
	}
	container.logger = container.logger.With(elog.FieldComponentName(key))
	return container
}

// Build
func (con *Container) Build(options ...Option) *Component {
	for _, option := range options {
		option(con)
	}
	return newComponent(con.config, con.client, con.logger)
}

func WithRedis(client *eredis.Component) Option {
	return func(c *Container) {
		c.client = client.Stub()
	}
}
