package ewechat

import "time"

type Option func(c *Container)

// ModName ..
const ModName = "contrib.wechat"

// config
type config struct {
	AppID          string
	AppSecret      string
	Token          string
	EncodingAESKey string
	PayMchID       string
	PayKey         string
	PayNotifyURL   string

	Debug                      bool          // 是否开启调试，默认不开启，开启后并加上export EGO_DEBUG=true，可以看到每次请求，配置名、地址、耗时、请求数据、响应数据
	RawDebug                   bool          // 是否开启原生调试，默认不开启
	ReadTimeout                time.Duration // 读超时，默认2s
	SlowLogThreshold           time.Duration // 慢日志记录的阈值，默认500ms
	EnableAccessInterceptor    bool          // 是否开启记录请求数据，默认不开启
	EnableAccessInterceptorRes bool          // 是否开启记录响应参数，默认不开启
}

// DefaultConfig ...
func DefaultConfig() *config {
	return &config{}
}
