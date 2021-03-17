package eredis

import (
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/gotomicro/ego/core/util/xtime"
)

const (
	// ClusterMode using clusterClient
	ClusterMode string = "cluster"
	// StubMode using stubClient
	StubMode string = "stub"
	// SentinelMode using Failover sentinel client
	SentinelMode string = "sentinel"
)

// config for redis, contains RedisStubConfig, RedisClusterConfig and RedisSentinelConfig
type config struct {
	Addrs                      []string      // Addrs 实例配置地址
	Addr                       string        // Addr stubConfig 实例配置地址
	Mode                       string        // Mode Redis模式 cluster|stub|sentinel
	MasterName                 string        // MasterName 哨兵主节点名称，sentinel模式下需要配置此项
	Password                   string        // Password 密码
	DB                         int           // DB，默认为0, 一般应用不推荐使用DB分片
	PoolSize                   int           // PoolSize 集群内每个节点的最大连接池限制 默认每个CPU10个连接
	MaxRetries                 int           // MaxRetries 网络相关的错误最大重试次数 默认8次
	MinIdleConns               int           // MinIdleConns 最小空闲连接数
	DialTimeout                time.Duration // DialTimeout 拨超时时间
	ReadTimeout                time.Duration // ReadTimeout 读超时 默认3s
	WriteTimeout               time.Duration // WriteTimeout 读超时 默认3s
	IdleTimeout                time.Duration // IdleTimeout 连接最大空闲时间，默认60s, 超过该时间，连接会被主动关闭
	Debug                      bool          // Debug开关， 是否开启调试，默认不开启，开启后并加上export EGO_DEBUG=true，可以看到每次请求，配置名、地址、耗时、请求数据、响应数据
	ReadOnly                   bool          // ReadOnly 集群模式 在从属节点上启用读模式
	SlowLogThreshold           time.Duration // 慢日志门限值，超过该门限值的请求，将被记录到慢日志中
	OnFail                     string        // OnFail panic|error
	EnableMetricInterceptor    bool          // 是否开启监控，默认开启
	EnableTraceInterceptor     bool          // 是否开启链路，默认开启
	EnableAccessInterceptor    bool          // 是否开启，记录请求数据
	EnableAccessInterceptorReq bool          // 是否开启记录请求参数
	EnableAccessInterceptorRes bool          // 是否开启记录响应参数
	interceptors               []redis.Hook
}

// DefaultConfig default config ...
func DefaultConfig() *config {
	return &config{
		Mode:                    StubMode,
		DB:                      0,
		PoolSize:                10,
		MaxRetries:              3,
		MinIdleConns:            100,
		DialTimeout:             xtime.Duration("1s"),
		ReadTimeout:             xtime.Duration("1s"),
		WriteTimeout:            xtime.Duration("1s"),
		IdleTimeout:             xtime.Duration("60s"),
		ReadOnly:                false,
		Debug:                   false,
		EnableMetricInterceptor: true,
		EnableTraceInterceptor:  true,
		SlowLogThreshold:        xtime.Duration("250ms"),
		OnFail:                  "panic",
	}
}
