package emns

import (
	"github.com/gotomicro/ego/core/elog"
)

type Container struct {
	name    string
	plugins map[string]SenderPlugin
	logger  *elog.Component
}

// DefaultContainer 构造默认容器
func DefaultContainer() *Container {
	return &Container{plugins: make(map[string]SenderPlugin)}
}

// Build 构建组件
func (c *Container) Build(options ...Option) *Component {
	for _, option := range options {
		option(c)
	}
	return newComponent(c)
}
