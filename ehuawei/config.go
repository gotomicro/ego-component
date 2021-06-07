package ehuawei

type config struct {
	// 华为云 ak
	AK string
	// 华为云 sk
	SK     string
	Region string
}

// DefaultConfig 返回默认配置
func DefaultConfig() *config {
	return &config{}
}
