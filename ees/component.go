package ees

import (
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/gotomicro/ego/core/elog"
)

const PackageName = "component.ees"

// Component ...
type Component struct {
	name   string
	config *config
	logger *elog.Component
	Client *elasticsearch.Client
}

// New ...
func newComponent(name string, config *config, logger *elog.Component) *Component {
	client, err := elasticsearch.NewClient(elasticsearch.Config{
		Addresses:             config.Addrs,
		Username:              config.Username,
		Password:              config.Password,
		APIKey:                config.APIKey,
		ServiceToken:          config.ServiceToken,
		RetryOnStatus:         config.RetryOnStatus,
		DisableRetry:          !config.EnableRetry,
		EnableRetryOnTimeout:  config.EnableRetryOnTimeout,
		MaxRetries:            config.MaxRetries,
		DiscoverNodesOnStart:  config.EnableDiscoverNodesOnStart,
		DiscoverNodesInterval: config.DiscoverNodesInterval,
		EnableMetrics:         config.EnableMetrics,
		EnableDebugLogger:     config.EnableDebugLogger,
		DisableMetaHeader:     !config.EnableMetaHeader,
	})
	if err != nil {
		logger.Panic("component new panic", elog.FieldErr(err))
	}

	cc := &Component{
		name:   name,
		logger: logger,
		config: config,
		Client: client,
	}

	return cc
}
