package ecronlock

import (
	"sync"

	"github.com/gotomicro/ego/core/elog"
	"github.com/gotomicro/ego/task/ecron"

	"github.com/gotomicro/ego-component/eredis"
)

type Component struct {
	name   string
	config *Config
	logger *elog.Component
	client eredis.ERedis
	mutuex sync.RWMutex
}

func newComponent(name string, config *Config, logger *elog.Component, client eredis.ERedis) *Component {
	reg := &Component{
		name:   name,
		logger: logger,
		config: config,
		client: client,
	}
	return reg
}

func (c *Component) NewLock(key string) ecron.Lock {
	return newRedisLock(c.client, key, c.logger)
}
