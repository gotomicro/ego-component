package ealiyun

// config options
type config struct {
	// 阿里云 AccessKeyId
	AccessKeyId string
	// 阿里云 AccessKeySecret
	AccessKeySecret string
	Endpoint        string
}

// DefaultConfig 返回默认配置
func DefaultConfig() *config {
	return &config{}
}
