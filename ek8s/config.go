package ek8s

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"k8s.io/client-go/rest"
)

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
		Addr:                    inClusterAddr(),
		Token:                   inClusterToken(),
		Namespaces:              []string{inClusterNamespace()},
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

func inClusterAddr() string {
	host := os.Getenv("KUBERNETES_SERVICE_HOST")
	port := os.Getenv("KUBERNETES_SERVICE_PORT")
	if host == "" || port == "" {
		return ""
	}
	return fmt.Sprintf("https://%s:%s", host, port)
}

func inClusterToken() string {
	t, err := ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/token")
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(t))
}

func inClusterNamespace() string {
	t, err := ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace")
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(t))
}
