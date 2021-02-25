package emongo

import (
	"github.com/gotomicro/ego/core/elog"
)

const PackageName = "component.emongo"

// Component client (cmdable and config)
type Component struct {
	config *config
	client *Client
	logger *elog.Component
}

// Client returns emongo Client
func (c *Component) Client() *Client {
	return c.client
}
