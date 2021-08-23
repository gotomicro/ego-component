package e2e

import (
	"context"
	"github.com/gotomicro/ego-component/ekafka"
	"github.com/segmentio/kafka-go"
	"testing"
	"time"
)

// 写入一条随机内容的消息然后验证是否能消费到
func Test_ConsumeWithConsumer(t *testing.T) {
	cmp := ekafka.Load("kafka").Build(
		ekafka.WithRegisterBalancer("my-balancer", &kafka.Hash{}),
	)

	randomMessage := RandomString(16)

	// 写一条随机字符串消息
	producer := cmp.Producer("p1")
	producerErr := make(chan error, 1)
	go writeMessage(producer, randomMessage, producerErr)

	// 尝试消费 producer 推送的随机字符串消息
	consumed := make(chan struct{}, 1)
	consumerGroupErr := make(chan error, 1)
	go func() {
		ctx, _ := context.WithTimeout(
			context.Background(),
			1*time.Minute,
		)
		consumer := cmp.Consumer("c1")
		for {
			msg, _, err := consumer.ReadMessage(ctx)
			if err != nil {
				consumerGroupErr <- err
				return
			}
			received := string(msg.Value)
			if received == randomMessage {
				if err := consumer.Close(); err != nil {
					consumerGroupErr <- err
					return
				}
				consumed <- struct{}{}
				return
			}
		}
	}()

	select {
	case <-consumed:
		// 成功
	case err := <-consumerGroupErr:
		t.Errorf("消费者发生错误: %s", err)
	case err := <-producerErr:
		t.Errorf("生产者发生错误: %s", err)
	}
}
