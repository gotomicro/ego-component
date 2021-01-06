package emongo

import (
	"github.com/gotomicro/ego/core/elog"
)

const PackageName = "component.emongo"

// Component client (cmdable and config)
type Component struct {
	Config *Config
	Client *WrappedClient
	logger *elog.Component
}
