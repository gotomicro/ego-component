package ekafka

import (
	"github.com/segmentio/kafka-go"
)

type Balancer = kafka.Balancer

// WithInterceptor 注入拦截器
func WithInterceptor(interceptors ...Interceptor) Option {
	return func(c *Container) {
		if c.config.interceptors == nil {
			c.config.interceptors = make([]Interceptor, 0)
		}
		c.config.interceptors = append(c.config.interceptors, interceptors...)
	}
}

// WithDebug 注入Debug配置
func WithDebug(debug bool) Option {
	return func(c *Container) {
		c.config.Debug = debug
	}
}

// WithBrokers 注入brokers配置
func WithBrokers(brokers ...string) Option {
	return func(c *Container) {
		c.config.Brokers = brokers
	}
}

// WithRegisterBalancer 注册名字为<balancerName>的balancer
// 注册之后可通过在producer配置文件中可通过<balancerName>来指定使用此balancer
func WithRegisterBalancer(balancerName string, balancer Balancer) Option {
	return func(c *Container) {
		c.config.balancers[balancerName] = balancer
	}
}
