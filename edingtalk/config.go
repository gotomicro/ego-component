package edingtalk

import (
	"github.com/gotomicro/ego-component/eredis"
	"time"
)

// config options
type Config struct {
	CorpID                       string
	AgentID                      int
	AppKey                       string
	AppSecret                    string
	Debug                        bool          // 是否开启调试，默认不开启，开启后并加上export EGO_DEBUG=true，可以看到每次请求，配置名、地址、耗时、请求数据、响应数据
	RawDebug                     bool          // 是否开启原生调试，默认不开启
	ReadTimeout                  time.Duration // 读超时，默认2s
	SlowLogThreshold             time.Duration // 慢日志记录的阈值，默认500ms
	EnableAccessInterceptor      bool          // 是否开启记录请求数据，默认不开启
	EnableAccessInterceptorReply bool          // 是否开启记录响应参数，默认不开启
	RedisPrefix                  string
	eredis                       *eredis.Component
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		CorpID:      "",
		AgentID:     0,
		AppKey:      "",
		AppSecret:   "",
		RedisPrefix: "/ego/edingtalk",
	}
}
