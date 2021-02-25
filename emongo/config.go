package emongo

import (
	"time"

	"github.com/gotomicro/ego/core/util/xtime"
)

type config struct {
	// DSN DSN地址
	DSN string `json:"dsn" toml:"dsn"`
	// Debug 是否开启debug模式
	Debug bool `json:"debug" toml:"debug"`
	// SocketTimeout 创建连接的超时时间
	SocketTimeout time.Duration `json:"socketTimeout" toml:"socketTimeout"`
	// PoolLimit 连接池大小(最大连接数)
	PoolLimit int `json:"poolLimit" toml:"poolLimit"`
	// EnableMetricInterceptor 是否启用prometheus metric拦截器
	EnableMetricInterceptor bool `json:"enableMetricInterceptor" toml:"enableMetricInterceptor"`
	// EnableAccessInterceptorReq 是否启用access req拦截器，此配置只有在EnableAccessInterceptor=true时才会生效
	EnableAccessInterceptorReq bool `json:"enableAccessInterceptorReq" toml:"enableAccessInterceptorReq"`
	// EnableAccessInterceptorRes 是否启用access res拦截器，此配置只有在EnableAccessInterceptor=true时才会生效
	EnableAccessInterceptorRes bool `json:"enableAccessInterceptorRes" toml:"enableAccessInterceptorRes"`
	// EnableAccessInterceptor 是否启用access拦截器
	EnableAccessInterceptor bool `json:"enableAccessInterceptor" toml:"enableAccessInterceptor"`
	// SlowLogThreshold 慢日志门限值，超过该门限值的请求，将被记录到慢日志中
	SlowLogThreshold time.Duration
	interceptors     []Interceptor
}

// DefaultConfig 返回默认配置
func DefaultConfig() *config {
	return &config{
		DSN:           "",
		Debug:         true,
		SocketTimeout: xtime.Duration("300s"),
		PoolLimit:     100,
	}
}
