package eredis

import (
	"context"
	"fmt"

	"github.com/go-redis/redis/v8"
	"github.com/gotomicro/ego/core/econf"
	"github.com/gotomicro/ego/core/elog"
)

type Option func(c *Container)

type Container struct {
	config *config
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

	options = append(options, withInterceptor(fixedInterceptor(c.name, c.config, c.logger)))
	if c.config.Debug {
		options = append(options, withInterceptor(debugInterceptor(c.name, c.config, c.logger)))
	}
	if c.config.EnableMetricInterceptor {
		options = append(options, withInterceptor(metricInterceptor(c.name, c.config, c.logger)))
	}
	if c.config.EnableAccessInterceptor {
		options = append(options, withInterceptor(accessInterceptor(c.name, c.config, c.logger)))
	}
	for _, option := range options {
		option(c)
	}

	var client redis.Cmdable
	switch c.config.Mode {
	case ClusterMode:
		if len(c.config.Addrs) == 0 {
			c.logger.Panic(`invalid "addrs" config, "addrs" has none addresses but with cluster mode"`)
		}
		client = c.buildCluster()
	case StubMode:
		if c.config.Addr == "" {
			c.logger.Panic(`invalid "addr" config, "addr" is empty but with cluster mode"`)
		}
		client = c.buildStub()
	case SentinelMode:
		if len(c.config.Addrs) == 0 {
			c.logger.Panic(`invalid "addrs" config, "addrs" has none addresses but with sentinel mode"`)
		}
		if c.config.MasterName == "" {
			c.logger.Panic(`invalid "masterName" config, "masterName" is empty but with sentinel mode"`)
		}
		client = c.buildSentinel()
	default:
		c.logger.Panic(`redis mode must be one of ("stub", "cluster", "sentinel")`)
	}

	c.logger = c.logger.With(elog.FieldAddr(fmt.Sprintf("%s", c.config.Addrs)))
	return &Component{
		config:     c.config,
		client:     client,
		lockClient: &lockClient{client: client},
		logger:     c.logger,
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

	for _, incpt := range c.config.interceptors {
		clusterClient.AddHook(incpt)
	}

	if err := clusterClient.Ping(context.Background()).Err(); err != nil {
		switch c.config.OnFail {
		case "panic":
			c.logger.Panic("start cluster redis", elog.FieldErr(err))
		default:
			c.logger.Error("start cluster redis", elog.FieldErr(err))
		}
	}
	return clusterClient
}

func (c *Container) buildSentinel() *redis.Client {
	sentinelClient := redis.NewFailoverClient(&redis.FailoverOptions{
		MasterName:    c.config.MasterName,
		SentinelAddrs: c.config.Addrs,
		Password:      c.config.Password,
		DB:            c.config.DB,
		MaxRetries:    c.config.MaxRetries,
		DialTimeout:   c.config.DialTimeout,
		ReadTimeout:   c.config.ReadTimeout,
		WriteTimeout:  c.config.WriteTimeout,
		PoolSize:      c.config.PoolSize,
		MinIdleConns:  c.config.MinIdleConns,
		IdleTimeout:   c.config.IdleTimeout,
	})

	for _, incpt := range c.config.interceptors {
		sentinelClient.AddHook(incpt)
	}

	if err := sentinelClient.Ping(context.Background()).Err(); err != nil {
		switch c.config.OnFail {
		case "panic":
			c.logger.Panic("start sentinel redis", elog.FieldErr(err))
		default:
			c.logger.Error("start sentinel redis", elog.FieldErr(err))
		}
	}
	return sentinelClient
}

func (c *Container) buildStub() *redis.Client {
	stubClient := redis.NewClient(&redis.Options{
		Addr:         c.config.Addr,
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

	for _, incpt := range c.config.interceptors {
		stubClient.AddHook(incpt)
	}

	if err := stubClient.Ping(context.Background()).Err(); err != nil {
		switch c.config.OnFail {
		case "panic":
			c.logger.Panic("start stub redis", elog.FieldErr(err))
		default:
			c.logger.Error("start stub redis", elog.FieldErr(err))
		}
	}
	return stubClient
}
