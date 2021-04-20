# ekafka 组件使用指南

## 基本组件

对 [kafka-go](https://github.com/segmentio/kafka-go) 进行了轻量封装，并提供了以下功能：

- 规范了标准配置格式，提供了统一的 Load().Build() 方法。
- 支持自定义拦截器
- 提供了默认的 Debug 拦截器，开启 Debug 后可输出 Request、Response 至终端。

### 快速上手

生产者消费者使用样例可参考 [example](examples/main.go)

## Consumer Server 组件

> 必须配置 ConsumerGroup 使用。

基本组件通常是搭配其他服务模块（如 [HTTP 服务](https://ego.gocn.vip/frame/server/http.html)）一起使用的，如果只想使用 Ego 做单纯的 Kafka 消费应用，可以使用 [Consumer Server 组件](consumerserver/)。

Consumer Server 组件依赖于基本组件提供 Kafka 消费者实例，同时实现了 `ego.Server` 接口以达到常驻运行的目的，并且和 Ego 框架共享生命周期。

Consumer Server 支持两种消费模式：

- 逐条消费：会在处理消息的回调执行完成之后立即 commit。但实际 commit 时机依赖于 [kafka-go 的配置](https://github.com/segmentio/kafka-go#managing-commits)，kafka-go 支持设置定时提交来提高性能
- 手动消费：Consumer Server 提供一个生命周期的 context 和 consumer 对象，开发者自行决定如何消费

### 配置说明

```toml
[kafka]
debug=true
brokers=["localhost:9094"]

[kafka.client]
timeout="3s"

[kafka.consumers.c1]
topic="sre-infra-test"
groupID="group-1"
# 指定 Consumer Group 未曾提交过 offset 时从何处开始消费
startOffset = -1
# 默认为同步提交，可以配置自动批量提交间隔来提高性能
commitInterval = "1s"

[kafkaConsumerServers.s1]
debug=true
# 使用 ekafka 中注册的哪一个 consumer，对应 `kafka.consumers.[name]` 配置项
consumerName="c1"
```

#### StartOffset

[Ref](https://github.com/segmentio/kafka-go/blob/882ccd8dc16155638a653defe226d6492b0a9da8/reader.go#L17-L18)

| 配置值 | 说明                             |
| ------ | -------------------------------- |
| -1     | LastOffset 从最新位置            |
| -2     | FirstOffset 从最旧位置 (default) |

### 示例

逐条消费：

```go
package main

import (
	"github.com/gotomicro/ego"
	"github.com/gotomicro/ego-component/ekafka"
	"github.com/gotomicro/ego-component/ekafka/consumerserver"
	"github.com/gotomicro/ego/core/elog"
	"github.com/gotomicro/ego/server/egovernor"
	"github.com/segmentio/kafka-go"
)

func main() {
	app := ego.New().Serve(
		// 可以搭配其他服务模块一起使用
		egovernor.Load("server.governor").Build(),

		// 初始化 Consumer Server
		func() *consumerserver.Component {
			// 依赖 `ekafka` 管理 Kafka consumer
			ec := ekafka.Load("kafka").Build()
			cs := consumerserver.Load("kafkaConsumerServers.s1").Build(
				consumerserver.WithEkafka(ec),
			)

			// 用来接收、处理 `kafka-go` 和处理消息的回调产生的错误
			consumptionErrors := make(chan error)

			// 注册处理消息的回调函数
			cs.OnEachMessage(consumptionErrors, func(ctx context.Context, message kafka.Message) error {
				elog.Infof("got a message: %s\n", string(message.Value))
				// 如果返回错误则会被转发给 `consumptionErrors`
				return nil
			})

			return cs
		}(),
		// 还可以启动多个 Consumer Server
	)
	if err := app.Run(); err != nil {
		elog.Panic("startup", elog.Any("err", err))
	}
}
```

手动消费：

```go
package main

import (
	"github.com/gotomicro/ego"
	"github.com/gotomicro/ego-component/ekafka"
	"github.com/gotomicro/ego-component/ekafka/consumerserver"
	"github.com/gotomicro/ego/core/elog"
)

func main() {
	app := ego.New().Serve(
		// 可以搭配其他服务模块一起使用
		egovernor.Load("server.governor").Build(),

		// 初始化 Consumer Server
		func() *consumerserver.Component {
			// 依赖 `ekafka` 管理 Kafka consumer
			ec := ekafka.Load("kafka").Build()
			cs := consumerserver.Load("kafkaConsumerServers.s1").Build(
				consumerserver.WithEkafka(ec),
			)

			// 注册处理消息的回调函数
			cs.OnStart(func(ctx context.Context, consumer *ekafka.Consumer) error {
				// 编写自己的消费逻辑...

				return nil
			})

			return cs
		}(),
		// 还可以启动多个 Consumer Server
	)
	if err := app.Run(); err != nil {
		elog.Panic("startup", elog.Any("err", err))
	}
}
```

## 测试

### E2E 测试

> 运行 E2E 测试需要 Kafka 环境。

首先将 `test/e2e/config/example.toml` 复制为 `test/e2e/config/e2e.toml` 并按实际情况修改，该文件即是运行 E2E 测试的配置文件。

在模块目录下执行命令运行测试：

```
$ make test-e2e
```
