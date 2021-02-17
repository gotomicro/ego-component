// Copyright 2020 
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package ewechat

import (
	"sync"

	"github.com/go-resty/resty/v2"
	"github.com/gotomicro/ego-component/eredis"
	"github.com/gotomicro/ego/core/econf"
	"github.com/gotomicro/ego/core/elog"

	"github.com/gotomicro/ego-component/ewechat/cache"
	"github.com/gotomicro/ego-component/ewechat/context"
)

type Container struct {
	config *Config
	name string
	Context *context.Context
	client  cache.Cache
	logger  *elog.Component
}

func DefaultContainer() *Container {
	return &Container{
		config: DefaultConfig(),
		logger: elog.EgoLogger.With(elog.FieldComponent(ModName)),
	}
}


// Invoker ...
func Load(key string) *Container {
	container := DefaultContainer()
	if err := econf.UnmarshalKey(key, container.config); err != nil {
		container.logger.Panic("parse wechat config panic",
			elog.FieldErr(err), elog.FieldKey(key), elog.FieldValueAny(container.config))
	}
	container.logger = container.logger.With(elog.FieldComponentName(key))
	return container
}

// Build
func (con *Container) Build(options ...Option) *Component {
	cfg := con.config
	ctx := new(context.Context)
	ctx.AppID = cfg.AppID
	ctx.AppSecret = cfg.AppSecret
	ctx.Token = cfg.Token
	ctx.EncodingAESKey = cfg.EncodingAESKey
	ctx.PayMchID = cfg.PayMchID
	ctx.PayKey = cfg.PayKey
	ctx.PayNotifyURL = cfg.PayNotifyURL
	con.Context = ctx
	ctx.SetAccessTokenLock(new(sync.RWMutex))
	ctx.SetJsAPITicketLock(new(sync.RWMutex))
	ctx.RestyClient = resty.New().SetDebug(cfg.Debug)

	for _, option := range options {
		option(con)
	}
	ctx.Cache = con.client
	return newComponent(cfg, ctx, con.client, con.logger)
}

func WithRedis(client *eredis.Component) Option {
	return func(c *Container) {
		c.client = client
	}
}

