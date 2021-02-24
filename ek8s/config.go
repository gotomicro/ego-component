package ek8s

import "k8s.io/client-go/rest"

// Config ...
type Config struct {
	Addr                    string
	Debug                   bool
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

func (c *Config) toRestConfig() *rest.Config {
	return &rest.Config{
		Host:        c.Addr,
		BearerToken: c.Token,
		TLSClientConfig: rest.TLSClientConfig{
			Insecure: c.TLSClientConfigInsecure,
		},
	}
}
