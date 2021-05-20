package emns

import (
	"github.com/gotomicro/ego/core/econf"
	"github.com/gotomicro/ego/core/elog"
	"github.com/gotomicro/ego/task/ecron"
)

const PackageName = "component.emns"

type (
	// Component emns 组件
	Component struct {
		plugins map[string]SenderPlugin
		crons   []ecron.Ecron
		name    string
		logger  *elog.Component
	}

	// SenderPlugin emns plugins
	SenderPlugin interface {
		Name() string                               // 名称
		Key() string                                // 唯一 key register
		Destroy() error                             // 销毁
		Init() error                                // 初始化
		Send(req *SendRequest) (resp *SendResponse) // send
		Cron() ecron.Ecron                          // 定时任务 没有可返回nil
	}

	// SendRequest 对业务 pb 不产生依赖
	SendRequest struct {
		Receiver       string            // 接收者
		Vars           map[string]string // 参数变量
		ExtraContent   string            // 业务方备注
		ExtraId        string            // 业务方扩展id
		Tpl            interface{}       // 模板对象
		ReceiverDetail interface{}       // 接收者详情
	}

	// SendResponse 对业务 pb 不产生依赖
	SendResponse struct {
		Code         int32  // code
		ExtraId      string // 业务方扩展id
		MsgId        string // msgId
		Reason       string // 详情
		FinalContent string // 最终发送内容
		PluginType   string // 使用的插件
		Retry        int    // 重试次数
	}
)

func newComponent(c *Container) *Component {
	return &Component{
		plugins: c.plugins,
		name:    c.name,
		logger:  c.logger,
		crons:   make([]ecron.Ecron, 0),
	}
}

func (c *Component) Name() string {
	return c.name
}

func (c *Component) PackageName() string {
	return PackageName
}

func (c *Component) Start() error {
	var err error
	if c.plugins != nil && len(c.plugins) > 0 {
		for k, v := range c.plugins {
			err = v.Init()
			if err != nil {
				c.logger.Error("init emns plugin error", elog.FieldName(k), elog.FieldErr(err))
			} else {
				c.logger.Info("init emns plugin", elog.FieldName(k))
			}
			// config 变动回调
			econf.OnChange(onConfChange(c, v, k))
			cron := v.Cron()
			if cron != nil {
				c.crons = append(c.crons, cron)
			}
		}
	}
	return err
}

func onConfChange(c *Component, v SenderPlugin, k string) func(conf *econf.Configuration) {
	return func(conf *econf.Configuration) {
		onchangeErr := v.Destroy()
		if onchangeErr != nil {
			c.logger.Error("onchange destroy emns plugin error", elog.FieldName(k), elog.FieldErr(onchangeErr))
		} else {
			c.logger.Info("emns plugin destroy success", elog.FieldName(v.Name()))
		}

		onchangeErr = v.Init()
		if onchangeErr != nil {
			c.logger.Error("onchange init emns plugin error", elog.FieldName(k), elog.FieldErr(onchangeErr))
		} else {
			c.logger.Info("emns plugin init success", elog.FieldName(v.Name()))
		}
	}
}

func (c *Component) Stop() error {
	var err error
	if c.plugins != nil && len(c.plugins) > 0 {
		for k, v := range c.plugins {
			err = v.Destroy()
			if err != nil {
				c.logger.Error("destroy emns plugin error", elog.FieldName(k), elog.FieldErr(err))
			} else {
				c.logger.Info("destroy emns plugin", elog.FieldName(k))
			}
		}
	}
	return nil
}

func (c *Component) Send(key string, req *SendRequest) (resp *SendResponse) {
	if plugin, ok := c.plugins[key]; !ok {
		c.logger.Error("plugin not exist", elog.FieldName(key))
		return &SendResponse{
			Code:    1,
			ExtraId: req.ExtraId,
			MsgId:   "",
			Reason:  "plugin not exist",
		}
	} else {
		resp = plugin.Send(req)
		resp.PluginType = plugin.Key()
		if resp.Code != 0 {
			for i := 0; i < 2; i++ {
				resp = plugin.Send(req)
				resp.Retry = i + 1
				if resp.Code == 0 {
					break
				}
			}
		}
		return
	}
}

// Crons 定时任务列表
func (c *Component) Crons() []ecron.Ecron {
	if len(c.crons) > 0 {
		return c.crons
	} else {
		return make([]ecron.Ecron, 0)
	}
}

// PluginStopFuncs 获取stop hook
func (c *Component) PluginStopFuncs() []func() error {
	arr := make([]func() error, 0, len(c.plugins))
	for _, p := range c.plugins {
		arr = append(arr, p.Destroy)
	}
	return arr
}
