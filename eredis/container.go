package eredis

import (
	"fmt"

	"github.com/go-redis/redis"
	"github.com/gotomicro/ego/core/econf"
	"github.com/gotomicro/ego/core/elog"
)

type Option func(c *Container)

type Container struct {
	config *Config
	name   string
	logger *elog.Component
}

// DefaultContainer 定义了默认Container配置
func DefaultContainer() *Container {
	return &Container{
		config: DefaultConfig(),
		logger: elog.EgoLogger.With(elog.FieldComponent(PackageName)),
	}
}

// Load 载入配置，初始化Container
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

// Build 构建Component
func (c *Container) Build(options ...Option) *Component {
	if options == nil {
		options = make([]Option, 0)
	}

	if c.config.Debug {
		options = append(options, WithInterceptor(debugInterceptor(c.name, c.config, c.logger)))
	}

	if c.config.EnableMetricInterceptor {
		options = append(options, WithInterceptor(metricInterceptor(c.name, c.config, c.logger)))
	}

	for _, option := range options {
		option(c)
	}

	count := len(c.config.Addrs)
	if count < 1 {
		c.logger.Panic("no address in redis config")
	}
	if len(c.config.Mode) == 0 {
		c.config.Mode = StubMode
		if count > 1 {
			c.config.Mode = ClusterMode
		}
	}
	var client redis.Cmdable
	switch c.config.Mode {
	case ClusterMode:
		if count == 1 {
			c.logger.Warn("redis config has only 1 address but with cluster mode")
		}
		client = c.buildCluster()
	case StubMode:
		if count > 1 {
			c.logger.Warn("redis config has more than 1 address but with stub mode")
		}
		client = c.buildStub()
	default:
		c.logger.Panic("redis mode must be one of (stub, cluster)")
	}

	c.logger = c.logger.With(elog.FieldAddr(fmt.Sprintf("%s", c.config.Addrs)))
	return &Component{
		Config: c.config,
		Client: client,
		logger: c.logger,
	}
}

func (c *Container) buildCluster() *redis.ClusterClient {
	clusterClient := redis.NewClusterClient(&redis.ClusterOptions{
		Addrs:        c.config.Addrs,
		MaxRedirects: c.config.MaxRetries,
		ReadOnly:     c.config.ReadOnly,
		Password:     c.config.Password,
		MaxRetries:   c.config.MaxRetries,
		DialTimeout:  c.config.DialTimeout,
		ReadTimeout:  c.config.ReadTimeout,
		WriteTimeout: c.config.WriteTimeout,
		PoolSize:     c.config.PoolSize,
		MinIdleConns: c.config.MinIdleConns,
		IdleTimeout:  c.config.IdleTimeout,
	})

	clusterClient.WrapProcess(InterceptorChain(c.config.interceptors...))

	if err := clusterClient.Ping().Err(); err != nil {
		switch c.config.OnFail {
		case "panic":
			c.logger.Panic("start cluster redis", elog.FieldErr(err))
		default:
			c.logger.Error("start cluster redis", elog.FieldErr(err))
		}
	}
	return clusterClient
}

func (c *Container) buildStub() *redis.Client {
	stubClient := redis.NewClient(&redis.Options{
		Addr:         c.config.Addrs[0],
		Password:     c.config.Password,
		DB:           c.config.DB,
		MaxRetries:   c.config.MaxRetries,
		DialTimeout:  c.config.DialTimeout,
		ReadTimeout:  c.config.ReadTimeout,
		WriteTimeout: c.config.WriteTimeout,
		PoolSize:     c.config.PoolSize,
		MinIdleConns: c.config.MinIdleConns,
		IdleTimeout:  c.config.IdleTimeout,
	})

	stubClient.WrapProcess(InterceptorChain(c.config.interceptors...))

	if err := stubClient.Ping().Err(); err != nil {
		switch c.config.OnFail {
		case "panic":
			c.logger.Panic("start stub redis", elog.FieldErr(err))
		default:
			c.logger.Error("start stub redis", elog.FieldErr(err))
		}
	}
	return stubClient
}
