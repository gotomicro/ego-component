package eminio

// config minio 配置
// 官网地址： https://docs.min.io/
type config struct {
	Endpoint        string // S3兼容对象存储服务endpoint
	AccessKeyID     string // 对象存储的Access key
	SecretAccessKey string // 对象存储的Secret key
	Ssl             bool   // true代表使用HTTPS
	Region          string // 对象存储的region(避免了bucket-location操作，所以会快那么一丢丢。如果你的应用只使用一个region的话可以配置Region值)
}

// DefaultConfig 返回默认配置
func DefaultConfig() *config {
	return &config{
		Ssl:      false,
		Endpoint: "localhost:9000",
	}
}
