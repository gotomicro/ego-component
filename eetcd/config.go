package eetcd

import (
	"time"

	"github.com/gotomicro/ego/core/util/xtime"
)

// config ...
type config struct {
	Addrs                        []string      `json:"endpoints"`
	CertFile                     string        `json:"certFile"`
	KeyFile                      string        `json:"keyFile"`
	CaCert                       string        `json:"caCert"`
	BasicAuth                    bool          `json:"basicAuth"`
	UserName                     string        `json:"userName"`
	Password                     string        `json:"-"`
	ConnectTimeout               time.Duration `json:"connectTimeout"` // 连接超时时间
	Secure                       bool          `json:"secure"`
	AutoSyncInterval             time.Duration `json:"autoAsyncInterval"` // 自动同步member list的间隔
	EnableBlock                  bool          // 是否开启阻塞，默认开启
	EnableFailOnNonTempDialError bool
	TTL                          int // 单位：s
}

// DefaultConfig 返回默认配置
func DefaultConfig() *config {
	return &config{
		BasicAuth:                    false,
		ConnectTimeout:               xtime.Duration("5s"),
		Secure:                       false,
		EnableBlock:                  true,
		EnableFailOnNonTempDialError: true,
	}
}
