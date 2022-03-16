package eemqtt

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/eclipse/paho.golang/autopaho"
	"github.com/eclipse/paho.golang/paho"
	"github.com/gotomicro/ego/core/elog"
	"net/url"
	"sync"
	"time"
)

const PackageName = "component.eemqtt"

// Component ...
type Component struct {
	ServerCtx        context.Context
	stopServer       context.CancelFunc
	name             string
	mod              int //0-初始化  1 运行中
	config           *config
	rmu              *sync.RWMutex
	logger           *elog.Component
	ec               *autopaho.ConnectionManager
	onPublishHandler OnPublishHandler
}

func newComponent(name string, config *config, logger *elog.Component) *Component {
	serverCtx, stopServer := context.WithCancel(context.Background())
	cc := &Component{
		ServerCtx:  serverCtx,
		stopServer: stopServer,
		mod:        0,
		name:       name,
		rmu:        &sync.RWMutex{},
		logger:     logger,
		config:     config,
	}
	logger.Info("dial emqtt server")
	return cc
}

/**
  建立连接，自动订阅以及消息回调
*/
func (c *Component) StartAndHandler(handler OnPublishHandler) {
	c.rmu.RLock()
	if c.mod == 0 {
		c.onPublishHandler = handler
		c.rmu.RUnlock()
		c.connServer()
	} else {
		c.rmu.RUnlock()
		c.logger.Error("client has started")
	}
}

func (c *Component) connServer() {
	if c.config.ServerURL == "" {
		c.logger.Panic("client emqtt serverUrl empty", elog.FieldValueAny(c.config))
	}
	urlParseStr, err := url.Parse(c.config.ServerURL)
	if err != nil {
		c.logger.Panic("client emqtt serverUrl Parse error", elog.FieldErr(err), elog.FieldValueAny(c.config))
	}
	cliCfg := autopaho.ClientConfig{
		BrokerUrls:        []*url.URL{urlParseStr},
		KeepAlive:         c.config.KeepAlive,
		ConnectRetryDelay: c.config.ConnectRetryDelay,
		ConnectTimeout:    c.config.ConnectTimeout,
		OnConnectionUp: func(cm *autopaho.ConnectionManager, connAck *paho.Connack) {
			c.logger.Info("mqtt connection up")
			sOs := make(map[string]paho.SubscribeOptions)
			for st := range c.config.SubscribeTopics {
				sOs[c.config.SubscribeTopics[st].Topic] = paho.SubscribeOptions{QoS: c.config.SubscribeTopics[st].Qos}
			}
			if len(sOs) > 0 {
				if _, err := cm.Subscribe(context.Background(), &paho.Subscribe{
					Subscriptions: sOs,
				}); err != nil {
					c.logger.Panic(fmt.Sprintf("failed to subscribe (%v). This is likely to mean no messages will be received.", sOs), elog.FieldErr(err))
				}
			}
		},
		OnConnectError: func(err error) { c.logger.Error("error whilst attempting connection", elog.FieldErr(err)) },
		ClientConfig: paho.ClientConfig{
			ClientID: c.config.ClientID,
			Router: paho.NewSingleHandlerRouter(func(pp *paho.Publish) {
				if c.onPublishHandler != nil {
					c.onPublishHandler(c.ServerCtx, pp)
				} else {
					c.logger.Info("Received message, but no handler is defined")
				}
			}),
			OnClientError: func(err error) { c.logger.Error("server requested disconnect", elog.FieldErr(err)) },
			OnServerDisconnect: func(d *paho.Disconnect) {
				if d.Properties != nil {
					c.logger.Info(fmt.Sprintf("server requested disconnect: %s\n", d.Properties.ReasonString))
				} else {
					c.logger.Info(fmt.Sprintf("server requested disconnect; reason code: %d\n", d.ReasonCode))
				}
			},
		},
	}

	if c.config.Debug {
		cliCfg.Debug = debugLogger{prefix: "emqtt-autoPaho"}
		cliCfg.PahoDebug = debugLogger{prefix: "emqtt-paho"}
	}

	if c.config.Username != "" && c.config.Password != "" {
		cliCfg.SetUsernamePassword(c.config.Username, ([]byte)(c.config.Password))
	}
	cm, err := autopaho.NewConnection(c.ServerCtx, cliCfg)
	if err != nil {
		c.logger.Panic("emqtt connect fialed", elog.FieldValueAny(c.config))
	} else {
		c.rmu.Lock()
		c.ec = cm
		c.mod = 1
		c.rmu.Unlock()
	}
}
func (c *Component) Client() *autopaho.ConnectionManager {
	return c.ec
}

func (c *Component) PublishMsg(topic string, qos byte, msg interface{}) {
	c.rmu.RLock()
	if c.mod == 0 {
		c.rmu.RUnlock()
		c.logger.Error("client not start")
		return
	}

	err := c.ec.AwaitConnection(c.ServerCtx)
	if err != nil { // Should only happen when context is cancelled
		c.logger.Error(fmt.Sprintf("publisher done (AwaitConnection: %s)", err))
		return
	}

	msgByte, err := json.Marshal(msg)
	if err != nil {
		c.logger.Panic("msg Parse error", elog.FieldErr(err), elog.FieldValueAny(msg))
		return
	}
	pr, err := c.ec.Publish(c.ServerCtx, &paho.Publish{
		QoS:     qos,
		Topic:   topic,
		Payload: msgByte,
	})
	if err != nil {
		c.logger.Error(fmt.Sprintf("error publishing: %s\n", err))
	} else if pr.ReasonCode != 0 && pr.ReasonCode != 16 { // 16 = Server received message but there are no subscribers
		c.logger.Info(fmt.Sprintf("reason code %d received\n", pr.ReasonCode))
	}
}

func (c *Component) Stop() {
	c.rmu.Lock()
	if c.mod == 1 {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		_ = c.ec.Disconnect(ctx)
		c.mod = 0
	}
	c.stopServer()
	c.rmu.Unlock()
}
