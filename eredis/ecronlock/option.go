package ecronlock

import "github.com/gotomicro/ego-component/eredis"

//WithClient Required. 设置 redis client
func WithClient(client *eredis.Component) Option {
	return func(c *Container) {
		c.client = client
	}
}
