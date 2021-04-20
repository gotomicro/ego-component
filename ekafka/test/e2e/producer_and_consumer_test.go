package e2e

import (
	"context"
	"testing"
	"time"

	"github.com/gotomicro/ego-component/ekafka"
	"github.com/segmentio/kafka-go"
)

// 写入一条随机内容的消息然后验证是否能消费到
func Test_ProduceAndConsume(t *testing.T) {
	cmp := ekafka.Load("kafka").Build(
		ekafka.WithRegisterBalancer("my-balancer", &kafka.Hash{}),
	)

	randomMessage := RandomString(16)
	consumed := make(chan struct{}, 1)

	// 写一条随机字符串消息
	go func() {
		producer := cmp.Producer("p1")
		err := producer.WriteMessages(
			context.Background(),
			ekafka.Message{Value: []byte(randomMessage)},
		)
		if err != nil {
			panic(err)
		}
		if err := producer.Close(); err != nil {
			panic(err)
		}
	}()

	// 尝试消费 producer 推送的随机字符串消息
	go func() {
		ctx, _ := context.WithTimeout(
			context.Background(),
			1*time.Minute,
		)
		consumer := cmp.Consumer("c1")
		for {
			msg, err := consumer.ReadMessage(ctx)
			if err != nil {
				panic(err)
			}
			received := string(msg.Value)
			if received == randomMessage {
				consumed <- struct{}{}
				if err := consumer.Close(); err != nil {
					panic(err)
				}
				return
			}
		}
	}()

	<-consumed
}
