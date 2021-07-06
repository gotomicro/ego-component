package eetcd

import (
	"context"

	"go.etcd.io/etcd/client/v3/concurrency"
)

// Mutex ...
type Mutex struct {
	s *concurrency.Session
	m *concurrency.Mutex
}

// NewMutex ...
func (c *Component) NewMutex(key string, opts ...concurrency.SessionOption) (mutex *Mutex, err error) {
	mutex = &Mutex{}
	// 默认session ttl = 60s
	mutex.s, err = concurrency.NewSession(c.Client, opts...)
	if err != nil {
		return
	}
	mutex.m = concurrency.NewMutex(mutex.s, key)
	return
}

// Lock ...
func (mutex *Mutex) Lock(ctx context.Context) (err error) {
	return mutex.m.Lock(ctx)
}

// Unlock ...
func (mutex *Mutex) Unlock(ctx context.Context) (err error) {
	err = mutex.m.Unlock(ctx)
	if err != nil {
		return
	}
	return mutex.s.Close()
}
