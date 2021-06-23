package registry

import (
	"time"
)

// Config Registry配置
type Config struct {
	Scheme       string        // 协议
	Prefix       string        // 注册前缀
	ReadTimeout  time.Duration // 读超时
	ServiceTTL   time.Duration // 服务续期
	OnFailHandle string        // 错误后处理手段，panic，error
}

const (
	defaultScheme = "etcd"
)

// DefaultConfig ...
func DefaultConfig() *Config {
	return &Config{
		Scheme:      defaultScheme,
		ReadTimeout: time.Second * 3,
		Prefix:      "ego",
		ServiceTTL:  0,
	}
}
