package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/gotomicro/ego/core/econf"
	"github.com/segmentio/kafka-go"

	"github.com/gotomicro/ego-component/ekafka"
)

// produce 生产消息
func produce(w *ekafka.Producer) {
	// 生产3条消息
	err := w.WriteMessages(context.Background(),
		&ekafka.Message{Key: []byte("Key-A"), Value: []byte("Hello World!")},
		&ekafka.Message{Key: []byte("Key-B"), Value: []byte("One!")},
		&ekafka.Message{Key: []byte("Key-C"), Value: []byte("Two!")},
	)
	if err != nil {
		log.Fatal("failed to write messages:", err)
	}
	if err := w.Close(); err != nil {
		log.Fatal("failed to close writer:", err)
	}
	fmt.Println(`produce message succ--------------->`)
}

// consume 使用consumer/consumerGroup消费消息
func consume(r *ekafka.Consumer) {
	ctx := context.Background()
	for {
		// ReadMessage 再收到下一个Message时，会阻塞
		msg, _, err := r.ReadMessage(ctx)
		if err != nil {
			panic("could not read message " + err.Error())
		}
		// 打印消息
		fmt.Println("received: ", string(msg.Value))
		err = r.CommitMessages(ctx, &msg)
		if err != nil {
			log.Printf("fail to commit msg:%v", err)
		}
	}
}

func main() {
	var stopCh = make(chan bool)
	// 假设你配置的toml如下所示
	conf := `
[kafka]
	debug=true
	brokers=["localhost:9091","localhost:9092","localhost:9093"]
	[kafka.client]
        timeout="3s"
	[kafka.producers.p1]        # 定义了名字为p1的producer
		topic="sre-infra-test"  # 指定生产消息的topic
		balancer="my-balancer"  # 指定balancer，此balancer非默认balancer，需要使用ekafka.WithRegisterBalancer()注册
	[kafka.consumers.c1]        # 定义了名字为c1的consumer
		topic="sre-infra-test"  # 指定消费的topic
		groupID="group-1"       # 如果配置了groupID，将初始化为consumerGroup	
	[kafka.consumers.c2]        # 定义了名字为c2的consumer
		topic="sre-infra-test"  # 指定消费的topic
		groupID="group-2"       # 如果配置了groupID，将初始化为consumerGroup	
`
	// 加载配置文件
	err := econf.LoadFromReader(strings.NewReader(conf), toml.Unmarshal)
	if err != nil {
		panic("LoadFromReader fail," + err.Error())
	}

	// 初始化ekafka组件
	cmp := ekafka.Load("kafka").Build(
		// 注册名为my-balancer的自定义balancer
		ekafka.WithRegisterBalancer("my-balancer", &kafka.Hash{}),
	)

	// 使用p1生产者生产消息
	go produce(cmp.Producer("p1"))

	// 使用c1消费者消费消息
	consume(cmp.Consumer("c1"))

	stopCh <- true
}
