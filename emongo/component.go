package emongo

import (
	"github.com/gotomicro/ego/core/elog"
)

const PackageName = "component.emongo"

// Component client (cmdable and config)
type Component struct {
	Config *Config
	Client *Client
	logger *elog.Component
}
