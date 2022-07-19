package esession

import (
	"github.com/ego-component/eredis"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/memstore"
	"github.com/gin-contrib/sessions/redis"
	"github.com/gin-gonic/gin"
	"github.com/gotomicro/ego/core/econf"
	"github.com/gotomicro/ego/core/elog"
)

type Option func(c *Container)

type Container struct {
	config *config
	name   string
	logger *elog.Component
}

func DefaultContainer() *Container {
	return &Container{
		config: DefaultConfig(),
		logger: elog.EgoLogger.With(elog.FieldComponent("component.esession")),
	}
}

func Load(key string) *Container {
	c := DefaultContainer()
	if err := econf.UnmarshalKey(key, &c.config); err != nil {
		c.logger.Panic("parse config error", elog.FieldErr(err), elog.FieldKey(key))
		return c
	}
	c.name = key
	c.logger = c.logger.With(elog.FieldComponentName(key))
	return c
}

// Build 构建mpms.Config实例
func (c *Container) Build(options ...Option) gin.HandlerFunc {
	for _, option := range options {
		option(c)
	}

	var store sessions.Store
	var err error
	switch c.config.Mode {
	case "redis":
		store, err = redis.NewStore(c.config.Size, c.config.Network, c.config.Addr, c.config.Password, []byte(c.config.Keypairs))
		if err != nil {
			c.logger.Panic("config new store panic", elog.FieldErr(err))
		}
	case "eredis":
		var options = []eredis.Option{
			eredis.WithAddr(c.config.Addr),
			eredis.WithMasterName(c.config.MasterName),
			eredis.WithAddrs(c.config.Addrs),
			eredis.WithPoolSize(c.config.Size),
			eredis.WithPassword(c.config.Password),
		}
		switch c.config.RedisMode {
		case "sentinel":
			options = append(options, eredis.WithSentinel())
		case "cluster":
			options = append(options, eredis.WithCluster())
		default:
			options = append(options, eredis.WithStub())
		}
		rc := eredis.DefaultContainer().Build(options...)
		store = NewERedisStore(rc.Client(), []byte(c.config.Keypairs))
	case "memstore":
		store = memstore.NewStore([]byte(c.config.Keypairs))
	}
	return sessions.Sessions(c.config.Name, store)
}
