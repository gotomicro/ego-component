package main

import (
	"github.com/gotomicro/ego"
	"github.com/gotomicro/ego-component/emns"
	"github.com/gotomicro/ego/core/econf"
	"github.com/gotomicro/ego/core/elog"
)

var emnsClient *emns.Component

// export EGO_DEBUG=true && go run main.go --config=config.toml
func main() {
	err := ego.New().
		Invoker(
			invokerMns,
			testMns,
		).
		Cron(emnsClient.Crons()...).
		Run()
	if err != nil {
		elog.Panic("startup", elog.FieldErr(err))
	}
}

func invokerMns() error {
	logger := elog.DefaultLogger
	emnsClient = emns.DefaultContainer().Build(
		emns.WithName("emns"), emns.WithLogger(logger),
		emns.WithSenderPlugin(NewEchoPlugin("echo", logger)))
	return nil
}

func testMns() error {
	request := &emns.SendRequest{
		Receiver:       "user1",
		Vars:           nil,
		ExtraContent:   "test",
		ExtraId:        "100",
		Tpl:            nil,
		ReceiverDetail: nil,
	}
	resp := emnsClient.Send("echo", request)
	elog.Infof("resp %+v", resp)
	return nil
}

type echoConfig struct {
	AppId     string `toml:"appId"`
	AppSecret string `toml:"appSecret"`
}

type echoPlugin struct {
	config echoConfig
	emns.ParentPlugin
}

func NewEchoPlugin(key string, logger *elog.Component) *echoPlugin {
	return &echoPlugin{
		ParentPlugin: emns.ParentPlugin{
			PluginKey:  key,
			PluginName: "echo",
			Logger:     logger,
		},
	}
}

func (p *echoPlugin) Init() (err error) {
	if err = econf.UnmarshalKey("emns.echo", &p.config); err != nil {
		p.Logger.Error("get email config error", elog.FieldErr(err))
		return
	}
	p.Logger.Info("config", elog.FieldValueAny(p.config))
	// do something
	return
}

func (p *echoPlugin) Destroy() error {
	// do something...
	return nil
}

func (p *echoPlugin) Send(req *emns.SendRequest) *emns.SendResponse {
	// req
	p.Logger.Info("req--------->", elog.FieldValueAny(req))
	// do something
	return &emns.SendResponse{
		Code:         0,
		ExtraId:      req.ExtraId,
		MsgId:        p.GenMsgId(),
		Reason:       "",
		FinalContent: "hello foo",
	}
}
