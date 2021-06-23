package eetcd

import (
	"time"

	"github.com/gotomicro/ego/core/util/xtime"
)

// config ...
type config struct {
	Addrs                        []string      // 地址
	CertFile                     string        // cert file
	KeyFile                      string        // key file
	CaCert                       string        // ca cert
	UserName                     string        // 用户名
	Password                     string        // 密码
	ConnectTimeout               time.Duration // 连接超时时间
	AutoSyncInterval             time.Duration // 自动同步member list的间隔
	EnableBasicAuth              bool          // 是否开启认证
	EnableSecure                 bool          // 是否开启安全
	EnableBlock                  bool          // 是否开启阻塞，默认开启
	EnableFailOnNonTempDialError bool          // 是否开启gRPC连接的错误信息
}

// DefaultConfig 返回默认配置
func DefaultConfig() *config {
	return &config{
		EnableBasicAuth:              false,
		ConnectTimeout:               xtime.Duration("5s"),
		EnableSecure:                 false,
		EnableBlock:                  true,
		EnableFailOnNonTempDialError: true,
	}
}
