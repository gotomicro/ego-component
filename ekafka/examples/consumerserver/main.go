package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/gotomicro/ego"
	"github.com/gotomicro/ego-component/ekafka"
	"github.com/gotomicro/ego-component/ekafka/consumerserver"
	"github.com/gotomicro/ego/core/econf"
	"github.com/gotomicro/ego/core/elog"
	"github.com/gotomicro/ego/server/egovernor"
	"github.com/segmentio/kafka-go"
)

func main() {
	conf := `
	[kafka]
	debug=true
	brokers=["localhost:9094"]
	[kafka.client]
        timeout="3s"
	[kafka.producers.p1]        # 定义了名字为p1的producer
		topic="sre-infra-test"  # 指定生产消息的topic

	[kafka.consumers.c1]        # 定义了名字为c1的consumer
		topic="sre-infra-test"  # 指定消费的topic
		groupID="group-1"       # 如果配置了groupID，将初始化为consumerGroup	

	[kafkaConsumerServers.s1]
	debug=true
	consumerName="c1"
`
	// 加载配置文件
	err := econf.LoadFromReader(strings.NewReader(conf), toml.Unmarshal)
	if err != nil {
		panic("LoadFromReader fail," + err.Error())
	}

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
				fmt.Printf("got a message: %s\n", string(message.Value))
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
