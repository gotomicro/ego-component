package ejenkins

import "time"

type Config struct {
	Addr     	string
	Username 	string
	Credential 	string

	Debug                      bool			 // default: false
	ReadTimeout                time.Duration // default: 2s
	RawDebug                   bool          // 是否开启原生调试，默认不开启
	EnableAccessInterceptor    bool          // 是否开启记录请求数据，默认不开启
	EnableAccessInterceptorRes bool          // 是否开启记录响应参数，默认不开启
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		Addr: "127.0.0.1:8080",
		Username: "admin",
		Credential: "admin",
		Debug: false,
		ReadTimeout: 2*time.Second,
		RawDebug: false,
		EnableAccessInterceptor: false,
		EnableAccessInterceptorRes: false,
	}
}

