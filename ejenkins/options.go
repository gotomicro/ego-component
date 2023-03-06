package ejenkins

import (
	"github.com/gotomicro/ego/core/elog"
)

type Option func(c *Container)

// WithLogger assign specific logger to ejenkins component. Optional
func WithLogger(logger *elog.Component) Option {
	return func(c *Container) {
		c.logger = logger
	}
}
