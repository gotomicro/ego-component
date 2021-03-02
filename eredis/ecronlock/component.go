package ecronlock

import (
	"context"
	"github.com/gotomicro/ego-component/eredis"
	"github.com/gotomicro/ego/core/elog"
	"sync"
	"time"
)

type Component struct {
	name   string
	config *Config
	logger *elog.Component
	client *eredis.Component
	mutuex sync.RWMutex
	locker *eredis.Lock
}

func newComponent(name string, config *Config, logger *elog.Component, client *eredis.Component) *Component {
	reg := &Component{
		name:   name,
		logger: logger,
		client: client,
	}
	return reg
}

func (c *Component) Lock(ctx context.Context, key string, ttl time.Duration) error {
	locker := c.client.LockClient()
	lock, err := locker.Obtain(ctx, key, ttl, eredis.WithLockOptionRetryStrategy(eredis.LinearBackoffRetry(ttl)))
	if err != nil {
		return err
	}
	c.mutuex.Lock()
	c.locker = lock
	c.mutuex.Unlock()
	return nil
}

func (c *Component) Unlock(ctx context.Context, key string) error {
	c.mutuex.Lock()
	locker := c.locker
	c.mutuex.Unlock()
	if locker == nil {
		return nil
	}

	err := c.locker.Release(ctx)
	if err != nil {
		c.logger.Warn("cron unlock warning", elog.FieldErr(err))
	}
	return nil
}

func (c *Component) Refresh(ctx context.Context, key string, ttl time.Duration) error {
	c.mutuex.Lock()
	locker := c.locker
	c.mutuex.Unlock()
	if locker == nil {
		return nil
	}
	return locker.Refresh(ctx, ttl)
}
