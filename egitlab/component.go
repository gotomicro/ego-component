package egitlab

import (
	"github.com/gotomicro/ego/core/elog"
	"github.com/xanzy/go-gitlab"
)

const packageName = "component.egitlab"

type Component struct {
	config *config
	logger *elog.Component
	client *gitlab.Client
}

func newComponent(config *config, logger *elog.Component) *Component {
	client, err := gitlab.NewClient(config.Token, gitlab.WithBaseURL(config.BaseUrl))
	if err != nil {
		panic("NewClient failed:" + err.Error())
	}
	return &Component{
		client: client,
		logger: logger,
		config: config,
	}
}

// Client 暴露gitlab api 原生 client
func (c *Component) Client() *gitlab.Client {
	return c.client
}
