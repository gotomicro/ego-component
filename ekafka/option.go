package ekafka

import (
	"github.com/segmentio/kafka-go"
)

type Balancer = kafka.Balancer

// WithClientInterceptor 注入拦截器
func WithClientInterceptor(interceptors ...ClientInterceptor) Option {
	return func(c *Container) {
		c.config.clientInterceptors = append(c.config.clientInterceptors, interceptors...)
	}
}

// WithServerInterceptor 注入拦截器
func WithServerInterceptor(interceptors ...ServerInterceptor) Option {
	return func(c *Container) {
		c.config.serverInterceptors = append(c.config.serverInterceptors, interceptors...)
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
