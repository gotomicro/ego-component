package emongo

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/gotomicro/ego/core/econf"
	"github.com/gotomicro/ego/core/elog"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Option func(c *Container)

type Container struct {
	config *Config
	name   string
	logger *elog.Component
}

func DefaultContainer() *Container {
	return &Container{
		config: DefaultConfig(),
		logger: elog.EgoLogger.With(elog.FieldComponent(PackageName)),
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

// WithInterceptor ...
func WithInterceptor(interceptors ...Interceptor) Option {
	return func(c *Container) {
		if c.config.interceptors == nil {
			c.config.interceptors = make([]Interceptor, 0)
		}
		c.config.interceptors = append(c.config.interceptors, interceptors...)
	}
}

func WithDSN(dsn string) Option {
	return func(c *Container) {
		c.config.DSN = dsn
	}
}

func (c *Container) newSession(config Config) *WrappedClient {
	// check config param
	c.isConfigErr(config)
	mps := uint64(config.PoolLimit)

	clientOpts := options.Client()
	clientOpts.MaxPoolSize = &mps
	clientOpts.SocketTimeout = &config.SocketTimeout
	client, err := Connect(context.Background(), clientOpts.ApplyURI(config.DSN))
	if err != nil {
		c.logger.Panic("dial mongo", elog.FieldAddr(config.DSN), elog.Any("error", err))
	}

	instances.Store(config.Name, client)
	client.wrapProcess(InterceptorChain(config.interceptors...))
	return client
}

var instances = sync.Map{}

// Range 遍历所有实例
func Range(fn func(name string, db *mongo.Client) bool) {
	instances.Range(func(key, val interface{}) bool {
		return fn(key.(string), val.(*mongo.Client))
	})
}

// Get 返回指定实例
func Get(name string) *mongo.Client {
	if ins, ok := instances.Load(name); ok {
		return ins.(*mongo.Client)
	}
	return nil
}

func (c *Container) isConfigErr(config Config) {
	if config.SocketTimeout == time.Duration(0) {
		c.logger.Panic("invalid config", elog.FieldExtMessage("socketTimeout"))
	}
	if config.PoolLimit == 0 {
		c.logger.Panic("invalid config", elog.FieldExtMessage("poolLimit"))
	}
}

// Build ...
func (c *Container) Build(options ...Option) *Component {
	if options == nil {
		options = make([]Option, 0)
	}
	if c.config.Debug {
		options = append(options, WithInterceptor(debugInterceptor(c.config)))
	}
	// if c.config.EnableMetricInterceptor {
	// 	options = append(options, WithInterceptor(metricInterceptor(c.config, c.logger)))
	// }
	for _, option := range options {
		option(c)
	}

	c.logger = c.logger.With(elog.FieldAddr(fmt.Sprintf("%s", c.config.DSN)))
	client := c.newSession(*c.config)
	return &Component{
		Config: c.config,
		Client: client,
		logger: c.logger,
	}
}
