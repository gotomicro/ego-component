package main

import (
	"context"
	"fmt"
	"log"

	"github.com/gotomicro/ego"
	"github.com/gotomicro/ego-component/ekafka"
	"github.com/gotomicro/ego/core/etrace"
	"github.com/opentracing/opentracing-go"
	"github.com/segmentio/kafka-go"
	"github.com/uber/jaeger-client-go"
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
	// 设置一个ctx
	_, ctx = etrace.StartSpanFromContext(
		context.Background(),
		"kafka",
	)
	md := etrace.MetadataReaderWriter{MD: map[string][]string{}}
	span := opentracing.SpanFromContext(ctx)
	err := opentracing.GlobalTracer().Inject(span.Context(), opentracing.HTTPHeaders, md)
	if err != nil {
		fmt.Println(err)
		return
	}
	headers := make([]kafka.Header, 0)
	md.ForeachKey(func(key, val string) error {
		headers = append(headers, kafka.Header{
			Key:   key,
			Value: []byte(val),
		})
		return nil
	})

	fmt.Println(md)
	span = opentracing.SpanFromContext(ctx)
	fmt.Printf("tid provider--------------->"+"%+v\n", span.Context().(jaeger.SpanContext).TraceID())

	// 生产3条消息
	err = w.WriteMessages(ctx,
		&ekafka.Message{Headers: headers, Key: []byte("Key-A"), Value: []byte("Hello World!")},
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

		mds := make(map[string][]string)
		for _, value := range msg.Headers {
			mds[value.Key] = []string{string(value.Value)}
		}

		md := etrace.MetadataReaderWriter{MD: mds}

		sc, err := opentracing.GlobalTracer().Extract(opentracing.HTTPHeaders, md)
		if err != nil {
			fmt.Println(err)
			return
		}

		span, ctx := etrace.StartSpanFromContext(
			ctx,
			"kafka",
			opentracing.ChildOf(sc),
		)

		fmt.Printf("tid consume--------------->"+"%+v\n", span.Context().(jaeger.SpanContext).TraceID())

		// 打印消息
		fmt.Println("received headers: ", msg.Headers)
		fmt.Println("received: ", string(msg.Value))
		err = r.CommitMessages(ctx, &msg)
		if err != nil {
			log.Printf("fail to commit msg:%v", err)
		}
	}
}
