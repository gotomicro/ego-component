package ees

import (
	"time"
)

// config ...
type config struct {
	Addrs                      []string      // A list of Elasticsearch nodes to use.
	Username                   string        // Username for HTTP Basic Authentication.
	Password                   string        // Password for HTTP Basic Authentication.
	APIKey                     string        // Base64-encoded token for authorization; if set, overrides username/password and service token.
	ServiceToken               string        // Service token for authorization; if set, overrides username/password.
	RetryOnStatus              []int         // List of status codes for retry. Default: 502, 503, 504.
	EnableRetry                bool          // Default: false.
	EnableRetryOnTimeout       bool          // Default: false.
	MaxRetries                 int           // Default: 3.
	EnableDiscoverNodesOnStart bool          // Discover nodes when initializing the client. Default: false.
	DiscoverNodesInterval      time.Duration // Discover nodes periodically. Default: disabled.
	EnableMetrics              bool          // Enable the metrics collection.
	EnableDebugLogger          bool          // Enable the debug logging.
	EnableMetaHeader           bool          // Disable the additional "X-Elastic-Client-Meta" HTTP header.
}

// DefaultConfig 返回默认配置
func DefaultConfig() *config {
	return &config{
		Addrs:                      nil,
		Username:                   "",
		Password:                   "",
		APIKey:                     "",
		ServiceToken:               "",
		RetryOnStatus:              []int{502, 503, 504},
		EnableRetry:                false,
		EnableRetryOnTimeout:       false,
		MaxRetries:                 3,
		EnableDiscoverNodesOnStart: false,
		DiscoverNodesInterval:      0,
		EnableMetrics:              false,
		EnableDebugLogger:          false,
		EnableMetaHeader:           false,
	}
}
