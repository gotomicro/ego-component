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

package etoken

import (
	"github.com/go-redis/redis"
	"github.com/gotomicro/ego-component/eredis"
	"github.com/gotomicro/ego/core/econf"
	"github.com/gotomicro/ego/core/elog"
)

type Container struct {
	config *Config
	client                    *redis.Client
	logger                    *elog.Component
}

func DefaultContainer() *Container {
	return &Container{
		config: DefaultConfig(),
		logger: elog.EgoLogger.With(elog.FieldComponent(PackageName)),
	}
}

// Invoker ...
func Load(key string) *Container {
	var container = DefaultContainer()
	if err := econf.UnmarshalKey(key, &container.config); err != nil {
		container.logger.Panic("parse wechat config panic",
			elog.FieldErr(err), elog.FieldKey(key), elog.FieldValueAny(container.config))
	}
	container.logger = container.logger.With(elog.FieldComponentName(key))
	return container
}

// Build
func (con *Container) Build(options ...Option) *Component {
	for _, option := range options {
		option(con)
	}
	return newComponent(con.config, con.client, con.logger)
}

func WithRedis(client *eredis.Component) Option {
	return func(c *Container) {
		c.client = client.Stub()
	}
}

