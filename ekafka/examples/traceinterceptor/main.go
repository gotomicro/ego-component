package main

import (
	"context"
	"fmt"
	"log"

	"github.com/gotomicro/ego"
	"github.com/gotomicro/ego-component/ekafka"
)

func main() {
	ego.New().Invoker(func() error {
		ctx := context.Background()
		// 初始化ekafka组件
		cmp := ekafka.Load("kafka").Build()
		// 使用p1生产者生产消息
		produce(ctx, cmp.Producer("p1"))

		//md.ForeachKey(func(key, val string) error {
		//	fmt.Println(key)
		//	fmt.Println(val)
		//	return nil
		//})

		// 使用c1消费者消费消息
		consume(cmp.Consumer("c1"))
		return nil
	}).Run()

}

// produce 生产消息
func produce(ctx context.Context, w *ekafka.Producer) {
	// 生产3条消息
	ctx = context.WithValue(ctx, "hello", "world")
	err := w.WriteMessages(ctx,
		&ekafka.Message{Key: []byte("Key-A"), Value: []byte("Hellohahah World!22222")},
	)
	if err != nil {
		log.Fatal("failed to write messages:", err)
	}
	if err := w.Close(); err != nil {
		log.Fatal("failed to close writer:", err)
	}
}

// consume 使用consumer/consumerGroup消费消息
func consume(r *ekafka.Consumer) {
	ctx := context.Background()
	for {
		// ReadMessage 再收到下一个Message时，会阻塞
		msg, ctxOutput, err := r.ReadMessage(ctx)
		if err != nil {
			panic("could not read message " + err.Error())
		}

		// 打印消息
		fmt.Println("received headers: ", msg.Headers)
		fmt.Println("received: ", string(msg.Value))
		err = r.CommitMessages(ctxOutput, &msg)
		if err != nil {
			log.Printf("fail to commit msg:%v", err)
		}
	}
}
