package ecronlock

import (
	"context"
	"sync"
	"time"

	"github.com/gotomicro/ego/core/elog"

	"github.com/gotomicro/ego-component/eredis"
)

type redisLock struct {
	mutex  sync.Mutex
	client eredis.ERedis
	key    string
	locker *eredis.Lock
	logger *elog.Component
}

func newRedisLock(client eredis.ERedis, key string, logger *elog.Component) *redisLock {
	return &redisLock{
		mutex:  sync.Mutex{},
		client: client,
		key:    key,
		locker: nil,
		logger: logger,
	}
}

func (c *redisLock) Lock(ctx context.Context, ttl time.Duration) error {
	locker := c.client.LockClient()
	lock, err := locker.Obtain(ctx, c.key, ttl, eredis.WithLockOptionRetryStrategy(eredis.LinearBackoffRetry(ttl)))
	if err != nil {
		return err
	}
	c.mutex.Lock()
	c.locker = lock
	c.mutex.Unlock()
	return nil
}

func (c *redisLock) Unlock(ctx context.Context) error {
	c.mutex.Lock()
	locker := c.locker
	c.mutex.Unlock()
	if locker == nil {
		return nil
	}

	err := c.locker.Release(ctx)
	if err != nil {
		c.logger.Warn("cron unlock warning", elog.FieldErr(err))
	}
	return nil
}

func (c *redisLock) Refresh(ctx context.Context, ttl time.Duration) error {
	c.mutex.Lock()
	locker := c.locker
	c.mutex.Unlock()
	if locker == nil {
		return nil
	}

	return locker.Refresh(ctx, ttl)
}