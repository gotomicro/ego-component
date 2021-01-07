package emongo

import (
	"time"

	"github.com/gotomicro/ego/core/util/xtime"
)

type Config struct {
	// DSN地址
	DSN string `json:"dsn" toml:"dsn"`
	// 是否开启debug模式
	Debug bool `json:"debug" toml:"debug"`
	// 创建连接的超时时间
	SocketTimeout time.Duration `json:"socketTimeout" toml:"socketTimeout"`
	// 连接池大小(最大连接数)
	PoolLimit    int `json:"poolLimit" toml:"poolLimit"`
	interceptors []Interceptor
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		DSN:           "",
		Debug:         true,
		SocketTimeout: xtime.Duration("300s"),
		PoolLimit:     100,
	}
}
