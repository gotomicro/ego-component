package ekubernetes

// Config ...
type Config struct {
	Addr                    string
	Token                   string
	Namespaces              []string
	DeploymentPrefix        string // 命名前缀
	TLSClientConfigInsecure bool
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		Addr:                    "127.0.0.1",
		TLSClientConfigInsecure: true,
	}
}
