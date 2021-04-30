package e2e

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/gotomicro/ego-component/ekafka"
	"github.com/gotomicro/ego-component/ekafka/consumerserver"
	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
)

// 测试 ConsumeServer 的 OnConsumerEachMessage 方法
// 写入一条随机内容的消息然后验证是否能消费到
func Test_ConsumeServer_OnConsumerEachMessage(t *testing.T) {
	ekafkaComponent := ekafka.Load("kafka").Build()

	randomMessage := RandomString(16)

	// 写一条随机字符串消息
	producer := ekafkaComponent.Producer("p1")
	producerErr := make(chan error, 1)
	go writeMessage(producer, randomMessage, producerErr)

	// 尝试消费 producer 推送的随机字符串消息
	consumed := make(chan struct{}, 1)
	consumptionErr := make(chan error, 10)
	consumerServerComponent := consumerserver.Load("kafkaConsumerServers.s1").Build(
		consumerserver.WithEkafka(ekafkaComponent),
	)
	go func() {
		timeoutCtx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		consumerServerComponent.OnEachMessage(
			consumptionErr,
			func(ctx context.Context, message kafka.Message) error {
				received := string(message.Value)
				if received == randomMessage {
					consumed <- struct{}{}
				}
				return nil
			},
		)
		if err := consumerServerComponent.Start(); err != nil {
			consumptionErr <- err
		}

		go func() {
			<-timeoutCtx.Done()
			consumptionErr <- fmt.Errorf("time out err")
			if err := consumerServerComponent.Stop(); err != nil {
				consumptionErr <- err
			}
		}()
	}()

	select {
	case <-consumed:
		// 成功
		consumerServerComponent.Stop()
	case err := <-consumptionErr:
		t.Errorf("消费者发生错误: %s", err)
	case err := <-producerErr:
		t.Errorf("生产者发生错误: %s", err)
	}
}

// 测试 ConsumeServer 的 OnConsumerStart 方法
// 写入一条随机内容的消息然后验证是否能消费到
func Test_ConsumeServer_OnConsumerStart(t *testing.T) {
	ekafkaComponent := ekafka.Load("kafka").Build()

	randomMessage := RandomString(16)

	// 写一条随机字符串消息
	producer := ekafkaComponent.Producer("p1")
	producerErr := make(chan error, 1)
	go writeMessage(producer, randomMessage, producerErr)

	// 尝试消费 producer 推送的随机字符串消息
	consumed := make(chan struct{}, 1)
	consumptionErr := make(chan error, 10)
	consumerServerComponent := consumerserver.Load("kafkaConsumerServers.s1").Build(
		consumerserver.WithEkafka(ekafkaComponent),
	)
	go func() {
		timeoutCtx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		consumerServerComponent.OnStart(
			func(ctx context.Context, consumer *ekafka.Consumer) error {
				for {
					msg, err := consumer.ReadMessage(ctx)
					if err != nil {
						return err
					}
					received := string(msg.Value)
					if received == randomMessage {
						consumed <- struct{}{}
						return nil
					}
				}
			},
		)
		if err := consumerServerComponent.Start(); err != nil {
			consumptionErr <- err
		}

		go func() {
			<-timeoutCtx.Done()
			consumptionErr <- fmt.Errorf("time out err")
			if err := consumerServerComponent.Stop(); err != nil {
				consumptionErr <- err
			}
		}()
	}()

	select {
	case <-consumed:
		// 成功
		consumerServerComponent.Stop()
	case err := <-consumptionErr:
		t.Errorf("消费者发生错误: %s", err)
	case err := <-producerErr:
		t.Errorf("生产者发生错误: %s", err)
	}
}

// 测试 ConsumeServer 的 OnConsumerGroupStart 方法
// 写入一条随机内容的消息然后验证是否能消费到
func Test_ConsumeServer_OnConsumerGroupStart(t *testing.T) {
	ekafkaComponent := ekafka.Load("kafka").Build()

	randomMessage := RandomString(16)

	var assignedPartitionsEventCount int32 = 0
	var revokedPartitionsEventCount int32 = 0

	// 写一条随机字符串消息
	producer := ekafkaComponent.Producer("p1")
	producerErr := make(chan error, 1)
	go writeMessage(producer, randomMessage, producerErr)

	// 尝试消费 producer 推送的随机字符串消息
	consumed := make(chan struct{}, 1)
	consumptionErr := make(chan error, 10)
	go func() {
		consumerServerComponent := consumerserver.Load("kafkaConsumerServers.s1").Build(
			consumerserver.WithEkafka(ekafkaComponent),
		)

		timeoutCtx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
		defer cancel()

		consumerServerComponent.OnConsumerGroupStart(
			func(ctx context.Context, consumerGroup *ekafka.ConsumerGroup) error {
				consumerGroupConsume(
					consumerGroup,
					ctx,
					consumed,
					consumptionErr,
					randomMessage,
					&assignedPartitionsEventCount,
					&revokedPartitionsEventCount,
					false,
				)
				return nil
			},
		)
		if err := consumerServerComponent.Start(); err != nil {
			consumptionErr <- err
		}

		go func() {
			<-timeoutCtx.Done()
			consumptionErr <- fmt.Errorf("time out err")
			if err := consumerServerComponent.Stop(); err != nil {
				consumptionErr <- err
			}
		}()
	}()

	select {
	case <-consumed:
		assert.Equal(t, int32(1), assignedPartitionsEventCount)
		assert.Equal(t, int32(0), revokedPartitionsEventCount)
	case err := <-consumptionErr:
		t.Errorf("消费者发生错误: %s", err)
	case err := <-producerErr:
		t.Errorf("生产者发生错误: %s", err)
	}
}
