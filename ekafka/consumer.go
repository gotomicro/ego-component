package ekafka

import (
	"context"
	"time"

	"github.com/gotomicro/ego/core/etrace"
	"github.com/gotomicro/ego/core/transport"
	"github.com/opentracing/opentracing-go"
	"github.com/segmentio/kafka-go"
)

// Consumer 消费者/消费者组，
type Consumer struct {
	r         *kafka.Reader
	processor ServerInterceptor
	logMode   bool
}

type Message = kafka.Message

type logMessage struct {
	Topic         string
	Partition     int
	Offset        int64
	HighWaterMark int64
	Key           string
	Value         string
	Headers       []logHeader

	// If not set at the creation, Time will be automatically set when
	// writing the message.
	Time time.Time
}
type logHeader struct {
	Key   string
	Value string
}

func messageToLog(value Message) logMessage {
	headers := make([]logHeader, 0, len(value.Headers))
	for _, val := range value.Headers {
		headers = append(headers, logHeader{
			Key:   val.Key,
			Value: string(val.Value),
		})
	}
	return logMessage{
		Topic:         value.Topic,
		Partition:     value.Partition,
		Offset:        value.Offset,
		HighWaterMark: value.HighWaterMark,
		Key:           string(value.Key),
		Value:         string(value.Value),
		Headers:       headers,
		Time:          value.Time,
	}
}

type Messages []*Message

func (m Messages) ToNoPointer() []Message {
	output := make([]Message, 0)
	for _, value := range m {
		output = append(output, *value)
	}
	return output
}

func (m Messages) ToLog() []logMessage {
	output := make([]logMessage, 0)
	for _, value := range m {
		output = append(output, messageToLog(*value))
	}
	return output
}

func (r *Consumer) setProcessor(p ServerInterceptor) {
	r.processor = p
}

func (r *Consumer) Close() error {
	return r.processor(func(ctx context.Context, msgs Messages, c *cmd) error {
		logCmd(r.logMode, c, "ConsumerClose")
		return r.r.Close()
	})(context.Background(), nil, &cmd{})
}

func (r *Consumer) CommitMessages(ctx context.Context, msgs ...*Message) (err error) {
	return r.processor(func(ctx context.Context, msgs Messages, c *cmd) error {
		logCmd(r.logMode, c, "CommitMessages")
		return r.r.CommitMessages(ctx, msgs.ToNoPointer()...)
	})(ctx, msgs, &cmd{})
}

func (r *Consumer) FetchMessage(ctx context.Context) (msg Message, ctxOutput context.Context, err error) {
	err = r.processor(func(ctx context.Context, msgs Messages, c *cmd) error {
		msg, err = r.r.FetchMessage(ctx)
		ctxOutput = getCtx(ctx, msg)
		logCmd(r.logMode, c, "FetchMessage", cmdWithMsg(msg))
		return err
	})(ctx, nil, &cmd{})
	return
}

func (r *Consumer) Lag() int64 {
	return r.r.Lag()
}

func (r *Consumer) Offset() int64 {
	return r.r.Offset()
}

func (r *Consumer) ReadLag(ctx context.Context) (lag int64, err error) {
	err = r.processor(func(ctx context.Context, msgs Messages, c *cmd) error {
		lag, err = r.r.ReadLag(ctx)
		logCmd(r.logMode, c, "ReadLag")
		return err
	})(ctx, nil, &cmd{})
	return
}

func (r *Consumer) ReadMessage(ctx context.Context) (msg Message, ctxOutput context.Context, err error) {
	err = r.processor(func(ctx context.Context, msgs Messages, c *cmd) error {
		msg, err = r.r.ReadMessage(ctx)
		ctxOutput = getCtx(ctx, msg)
		logCmd(r.logMode, c, "ReadMessage", cmdWithRes(msg), cmdWithMsg(msg))
		return err
	})(ctx, nil, &cmd{})
	return
}

func (r *Consumer) SetOffset(offset int64) (err error) {
	return r.processor(func(ctx context.Context, msgs Messages, c *cmd) error {
		logCmd(r.logMode, c, "SetOffset")
		return r.r.SetOffset(offset)
	})(context.Background(), nil, &cmd{})
}

func (r *Consumer) SetOffsetAt(ctx context.Context, t time.Time) (err error) {
	return r.processor(func(ctx context.Context, msgs Messages, c *cmd) error {
		logCmd(r.logMode, c, "SetOffsetAt")
		return r.r.SetOffsetAt(ctx, t)
	})(ctx, nil, &cmd{})
}

func getCtx(ctx context.Context, msg Message) context.Context {
	ctxOutput := ctx
	// 我也不想这么处理trace。奈何协议头在用户数据里，无能为力。。。
	if opentracing.IsGlobalTracerRegistered() {
		mds := make(map[string][]string)
		for _, value := range msg.Headers {
			mds[value.Key] = []string{string(value.Value)}
		}
		md := etrace.MetadataReaderWriter{MD: mds}
		sc, _ := opentracing.GlobalTracer().Extract(opentracing.HTTPHeaders, md)

		// 重新赋值ctx
		_, ctxOutput = etrace.StartSpanFromContext(
			ctx,
			"kafka",
			opentracing.ChildOf(sc),
		)

		for _, key := range transport.CustomContextKeys() {
			for _, value := range msg.Headers {
				if value.Key == key {
					ctxOutput = transport.WithValue(ctxOutput, value.Key, string(value.Value))
				}
			}
		}
	}
	return ctxOutput
}
