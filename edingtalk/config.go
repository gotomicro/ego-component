package edingtalk

import (
	"time"

	"github.com/gotomicro/ego-component/eredis"
)

// config options
type config struct {
	CorpID                string
	AgentID               int
	AppKey                string // 用于企业内部接口
	AppSecret             string // 用于企业内部接口
	Oauth2AppKey          string // 用于第三方的app key
	Oauth2AppSecret       string // 用于第三方的app secret
	Oauth2RedirectUri     string
	Oauth2StateCookieName string

	Debug                      bool          // 是否开启调试，默认不开启，开启后并加上export EGO_DEBUG=true，可以看到每次请求，配置名、地址、耗时、请求数据、响应数据
	RawDebug                   bool          // 是否开启原生调试，默认不开启
	ReadTimeout                time.Duration // 读超时，默认2s
	SlowLogThreshold           time.Duration // 慢日志记录的阈值，默认500ms
	EnableAccessInterceptor    bool          // 是否开启记录请求数据，默认不开启
	EnableAccessInterceptorRes bool          // 是否开启记录响应参数，默认不开启
	RedisPrefix                string        // redis前缀
	RedisBaseToken             string        // 存放gettoken地址的路径
	eredis                     *eredis.Component
}

// DefaultConfig 返回默认配置
func DefaultConfig() *config {
	return &config{
		CorpID:                "",
		AgentID:               0,
		AppKey:                "",
		AppSecret:             "",
		RedisPrefix:           "/ego/edingtalk",
		Oauth2StateCookieName: "dingtalk_oauth2_state",
		RedisBaseToken:        "/base/token",
	}
}
