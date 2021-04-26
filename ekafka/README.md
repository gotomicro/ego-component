# ekafka 组件使用指南

## 基本组件

对 [kafka-go](https://github.com/segmentio/kafka-go) 进行了轻量封装，并提供了以下功能：

- 规范了标准配置格式，提供了统一的 Load().Build() 方法。
- 支持自定义拦截器
- 提供了默认的 Debug 拦截器，开启 Debug 后可输出 Request、Response 至终端。

### 快速上手

生产者消费者使用样例可参考 [example](examples/main.go)

## ConsumerGroup

相对于 Consumer（对 [kafka-go](https://github.com/segmentio/kafka-go) 的封装）来说，ConsumerGroup 则提供了更加易用的 API。这是一个简单的例子：

首先添加所需配置：

```yaml
[kafka]
brokers = ["127.0.0.1:9092", "127.0.0.1:9093", "127.0.0.1:9094"]

[kafka.client]
timeout = "3s"

[kafka.consumerGroups.cg1]
JoinGroupBackoff = "1s"
groupID = "group-1"
topic = "my-topic"
```

```go
package main

import "github.com/gotomicro/ego-component/ekafka"

func main() {
	cmp := ekafka.Load("kafka").Build()
	// 获取实例（第一次调用时初始化，再次获取时会复用）
	cg := cmp.ConsumerGroup("cg1")

	for {
		pollCtx, _ := context.WithTimeout(ctx, 1*time.Minute)
		// 拉取事件，可能是消息、Rebalancing 事件或者错误等
		event, err := consumerGroup.Poll(pollCtx)
		if err != nil {
			elog.Panic("poll error")
			return
		}
		switch e := event.(type) {
		case ekafka.Message:
			// 按需处理消息
		case ekafka.AssignedPartitions:
			// 在 Kafka 完成分区分配时触发一次
		case ekafka.RevokedPartitions:
			// 在当前 Generation 结束时触发一次
		case error:
			// 错误处理

			// 结束
			if err := cg.Close(); err != nil {
				elog.Panic("关闭ConsumerGroup失败")
			}
		default:
			// ...
		}
	}
}
```

### 配置说明

```yaml
[kafka.consumerGroups.cg1]
# GroupID is the name of the consumer group.
groupID = "group-1"
# The topic to read messages from.
topic = "my-topic"
# HeartbeatInterval sets the optional frequency at which the reader sends the consumer
# group heartbeat update.
#
# Default: 3s
heartbeatInterval = "3s"
# PartitionWatchInterval indicates how often a reader checks for partition changes.
# If a reader sees a partition change (such as a partition add) it will rebalance the group
# picking up new partitions.
#
# Default: 5s
partitionWatchInterval = "5s"
# WatchForPartitionChanges is used to inform kafka-go that a consumer group should be
# polling the brokers and rebalancing if any partition changes happen to the topic.
watchPartitionChanges = false
# SessionTimeout optionally sets the length of time that may pass without a heartbeat
# before the coordinator considers the consumer dead and initiates a rebalance.
#
# Default: 30s
sessionTimeout = "30s"
# RebalanceTimeout optionally sets the length of time the coordinator will wait
# for members to join as part of a rebalance.  For kafka servers under higher
# load, it may be useful to set this value higher.
#
# Default: 30s
rebalanceTimeout = "30s"
# JoinGroupBackoff optionally sets the length of time to wait before re-joining
# the consumer group after an error.
#
# Default: 5s
joinGroupBackoff = "5s"
# StartOffset determines from whence the consumer group should begin
# consuming when it finds a partition without a committed offset.  If
# non-zero, it must be set to one of FirstOffset or LastOffset.
#
# Default: `-2` (FirstOffset)
startOffset = "-2"
# RetentionTime optionally sets the length of time the consumer group will
# be saved by the broker.  -1 will disable the setting and leave the
# retention up to the broker's offsets.retention.minutes property.  By
# default, that setting is 1 day for kafka < 2.0 and 7 days for kafka >=
# 2.0.
#
# Default: -1
retentionTime = "-1"
# MinBytes indicates to the broker the minimum batch size that the consumer
# will accept. Setting a high minimum when consuming from a low-volume topic
# may result in delayed delivery when the broker does not have enough data to
# satisfy the defined minimum.
#
# Default: 1
minBytes = 1
# MaxBytes indicates to the broker the maximum batch size that the consumer
# will accept. The broker will truncate a message to satisfy this maximum, so
# choose a value that is high enough for your largest message size.
#
# Default: 1MB
maxBytes = 1048576
# Maximum amount of time to wait for new data to come when fetching batches
# of messages from kafka.
#
# Default: 10s
maxWait = "10s"
# ReadLagInterval sets the frequency at which the reader lag is updated.
# Setting this field to a negative value disables lag reporting.
#
# Default: 60s
readLagInterval = "60s"
# CommitInterval indicates the interval at which offsets are committed to
# the broker.  If 0, commits will be handled synchronously.
#
# Default: 0
commitInterval = "0"
# BackoffDelayMin optionally sets the smallest amount of time the reader will wait before
# polling for new messages
#
# Default: 100ms
readBackoffMin = "100ms"
# BackoffDelayMax optionally sets the maximum amount of time the reader will wait before
# polling for new messages
#
# Default: 1s
readBackoffMax = "1s"
```

## Consumer Server 组件

> 必须配合 ConsumerGroup 使用。

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

[kafka.consumerGroups.cg1]
topic="sre-infra-test"
groupID="group-1"
# 指定 Consumer Group 未曾提交过 offset 时从何处开始消费
startOffset = -1
# 默认为同步提交，可以配置自动批量提交间隔来提高性能
commitInterval = "1s"

[kafkaConsumerServers.s1]
debug=true
# 使用 ekafka 中注册的哪一个 Consumer，对应 `kafka.consumers.[name]` 配置项
consumerName="c1"
# 也可以配合 ConsumerGroup 使用
consumerGroupName="cg1"

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
				// 如果返回错误则会被转发给 `consumptionErrors`，默认出现任何错误都会导致消费终止、
				// ConsumerGroup 退出；但可以将错误标记为 Retryable 以实现重试，ConsumerGroup 最多重试 3 次
				// 如：
				// return fmt.Errorf("%w 写入数据库时发生错误", consumerserver.ErrRecoverableError)

				return nil
			})

			return cs
		}(),
		// 还可以启动多个 ConsumerServer
	)
	if err := app.Run(); err != nil {
		elog.Panic("startup", elog.Any("err", err))
	}
}
```

获取 Consumer 实例后手动消费：

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

		// 初始化 ConsumerServer
		func() *consumerserver.Component {
			// 依赖 `ekafka` 管理 Kafka Consumer
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
		// 还可以启动多个 ConsumerServer
	)
	if err := app.Run(); err != nil {
		elog.Panic("startup", elog.Any("err", err))
	}
}
```

获取 ConsumerGroup 实例后手动消费：

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

		// 初始化 ConsumerServer
		func() *consumerserver.Component {
			// 依赖 `ekafka` 管理 Kafka ConsumerGroup
			ec := ekafka.Load("kafka").Build()
			cs := consumerserver.Load("kafkaConsumerServers.s1").Build(
				consumerserver.WithEkafka(ec),
			)

			// 注册处理消息的回调函数
			cs.OnConsumerGroupStart(func(ctx context.Context, consumerGroup *ekafka.ConsumerGroup) error {
				// 编写自己的消费逻辑...

				return nil
			})

			return cs
		}(),
		// 还可以启动多个 ConsumerServer
	)
	if err := app.Run(); err != nil {
		elog.Panic("startup", elog.Any("err", err))
	}
}
```

## 测试

### E2E 测试

> 运行 E2E 测试需要准备 Kafka 环境，推荐 3 个 broker、每 topic 3 个 partition，否则有些测试会报错。

首先将 `test/e2e/config/example.toml` 复制为 `test/e2e/config/e2e.toml` 并按实际情况修改，该文件即是运行 E2E 测试的配置文件。

在 `ekafka/` 目录下执行命令运行测试：

```
$ make test-e2e
```
