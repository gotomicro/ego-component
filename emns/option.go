package emns

import "github.com/gotomicro/ego/core/elog"

type Option func(c *Container)

// WithName 设置name
func WithName(name string) Option {
	return func(c *Container) {
		c.name = name
	}
}

// WithSenderPlugin 设置插件
func WithSenderPlugin(plugins ...SenderPlugin) Option {
	return func(c *Container) {
		if plugins != nil && len(plugins) > 0 {
			for _, plugin := range plugins {
				c.plugins[plugin.Key()] = plugin
			}
		}
	}
}

// WithLogger 设置 log
func WithLogger(logger *elog.Component) Option {
	return func(c *Container) {
		c.logger = logger
	}
}
