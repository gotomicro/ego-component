package e2e

import (
	"context"
	"errors"
	"fmt"
	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gotomicro/ego-component/ekafka"
)

func consumerGroupConsume(
	consumerGroup *ekafka.ConsumerGroup,
	ctx context.Context,
	consumedCh chan<- struct{},
	consumerGroupErrCh chan<- error,
	expectedMessage string,
	assignedPartitionsEventCount *int32,
	revokedPartitionsEventCount *int32,
	closeConsumerGroup bool,
) {
	for {
		pollCtx, _ := context.WithTimeout(ctx, 1*time.Minute)
		event, err := consumerGroup.Poll(pollCtx)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return
			}
			consumerGroupErrCh <- err
			return
		}
		switch e := event.(type) {
		case ekafka.Message:
			received := string(e.Value)
			if received == expectedMessage {
				commitCtx, _ := context.WithTimeout(ctx, 10*time.Second)
				err := consumerGroup.CommitMessages(commitCtx, ekafka.Message{
					Partition: e.Partition,
					Offset:    e.Offset,
				})
				if err != nil {
					if errors.Is(err, context.Canceled) {
						return
					}
					consumerGroupErrCh <- err
					return
				}

				if closeConsumerGroup {
					if err := consumerGroup.Close(); err != nil {
						consumerGroupErrCh <- err
						return
					}
				}

				consumedCh <- struct{}{}
				return
			}
		case ekafka.AssignedPartitions:
			atomic.AddInt32(assignedPartitionsEventCount, 1)
		case ekafka.RevokedPartitions:
			atomic.AddInt32(revokedPartitionsEventCount, 1)
		case error:
			consumerGroupErrCh <- e
		default:
			consumerGroupErrCh <- fmt.Errorf("不应该收到任何其他消息")
		}
	}
}

// 启动两个 consumer（ConsumerGroup）同时消费，测试：
// - 两方都可以收到 AssignedPartitions Event
// - 一方可以收到 RevokedPartitions Event
// - 一方可以消费到消息
func Test_ConsumeWithConsumerGroup(t *testing.T) {
	cmp := ekafka.Load("kafka").Build(
		ekafka.WithRegisterBalancer("my-balancer", &kafka.Hash{}),
	)

	randomMessage := RandomString(16)

	var assignedPartitionsEventCount int32 = 0
	var revokedPartitionsEventCount int32 = 0

	// 写一条随机字符串消息
	producerErr := make(chan error, 1)
	producer := cmp.Producer("p1")
	go func() {
		time.Sleep(30 * time.Second)
		writeMessage(producer, randomMessage, producerErr)
	}()

	consumed := make(chan struct{}, 1)
	consumerGroupErr := make(chan error, 1)
	consumeCtx, cancelConsume := context.WithCancel(context.Background())
	defer cancelConsume()

	consumerGroup1 := cmp.ConsumerGroup("cg1")
	go consumerGroupConsume(
		consumerGroup1,
		consumeCtx,
		consumed,
		consumerGroupErr,
		randomMessage,
		&assignedPartitionsEventCount,
		&revokedPartitionsEventCount,
		true,
	)

	go func() {
		time.Sleep(15 * time.Second)
		consumerGroup2 := cmp.ConsumerGroup("cg2")
		consumerGroupConsume(
			consumerGroup2,
			consumeCtx,
			consumed,
			consumerGroupErr,
			randomMessage,
			&assignedPartitionsEventCount,
			&revokedPartitionsEventCount,
			true,
		)
	}()

	select {
	case <-consumed:
		// 只要消费成功则关闭所有消费者
		cancelConsume()

		// 因为有两个 ConsumerGroup 实例，一个先启动、一个后启动，所以应该收到：
		// - 3 次 AssignedPartitions Event
		// - 1 次 RevokedPartitions Event
		assert.Equal(t, int32(3), assignedPartitionsEventCount)
		assert.Equal(t, int32(1), revokedPartitionsEventCount)
	case err := <-consumerGroupErr:
		t.Errorf("消费者发生错误: %s", err)
	case err := <-producerErr:
		t.Errorf("生产者发生错误: %s", err)
	}
}
