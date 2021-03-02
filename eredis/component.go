package eredis

import (
	"github.com/go-redis/redis/v8"
	"github.com/gotomicro/ego/core/elog"
)

const PackageName = "component.eredis"

// Component client (cmdable and config)
type Component struct {
	config     *config
	client     redis.Cmdable
	lockClient *lockClient
	logger     *elog.Component
}

// Client returns a universal redis client(ClusterClient, StubClient or SentinelClient), it depends on you config.
func (r *Component) Client() redis.Cmdable {
	return r.client
}

// Cluster try to get a redis.ClusterClient
func (r *Component) Cluster() *redis.ClusterClient {
	if c, ok := r.client.(*redis.ClusterClient); ok {
		return c
	}
	return nil
}

// Stub try to get a redis.client
func (r *Component) Stub() *redis.Client {
	if c, ok := r.client.(*redis.Client); ok {
		return c
	}
	return nil
}

// Sentinel try to get a redis Failover Sentinel client
func (r *Component) Sentinel() *redis.Client {
	if c, ok := r.client.(*redis.Client); ok {
		return c
	}
	return nil
}

// LockClient gets default distributed Lock client
func (r *Component) LockClient() *lockClient {
	return r.lockClient
}
