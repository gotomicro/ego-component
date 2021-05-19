package emns

import (
	"github.com/google/uuid"
	"github.com/gotomicro/ego/core/elog"
	"github.com/gotomicro/ego/task/ecron"
)

type ParentPlugin struct {
	PluginKey  string // reg key
	PluginName string // Name
	Logger     *elog.Component
}

func (p *ParentPlugin) Key() string {
	return p.PluginKey
}

func (p *ParentPlugin) Name() string {
	return p.PluginName
}

// GenMsgId
// MsgId
func (p *ParentPlugin) GenMsgId() string {
	return uuid.New().String()
}

func (p ParentPlugin) Cron() ecron.Ecron {
	return nil
}
