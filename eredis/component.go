package eredis

import (
	"github.com/go-redis/redis"
	"github.com/gotomicro/ego/core/elog"
)

const PackageName = "component.eredis"

// Component client (cmdable and config)
type Component struct {
	Config *Config
	Client redis.Cmdable
	logger *elog.Component
}

// Cluster try to get a redis.ClusterClient
func (r *Component) Cluster() *redis.ClusterClient {
	if c, ok := r.Client.(*redis.ClusterClient); ok {
		return c
	}
	return nil
}

// Stub try to get a redis.Client
func (r *Component) Stub() *redis.Client {
	if c, ok := r.Client.(*redis.Client); ok {
		return c
	}
	return nil
}

// Sentinel try to get a redis Failover Sentinel Client
func (r *Component) Sentinel() *redis.Client {
	if c, ok := r.Client.(*redis.Client); ok {
		return c
	}
	return nil
}
