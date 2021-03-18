package ejira

import "time"

// config option
type config struct {
	Addr     string
	Username string
	Password string

	Debug                      bool
	ReadTimeout                time.Duration // 读超时，默认2s
	RawDebug                   bool          // 是否开启原生调试，默认不开启
	SlowLogThreshold           time.Duration // 慢日志记录的阈值，默认500ms
	EnableAccessInterceptor    bool          // 是否开启记录请求数据，默认不开启
	EnableAccessInterceptorRes bool          // 是否开启记录响应参数，默认不开启
}

// DefaultConfig 返回默认配置
func DefaultConfig() *config {
	return &config{
		Addr: "",
	}
}
