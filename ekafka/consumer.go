package ekafka

import (
	"context"
	"time"

	"github.com/segmentio/kafka-go"
)

// Consumer 消费者/消费者组，
type Consumer struct {
	r         *kafka.Reader
	processor processor
	logMode   bool
}

type Message = kafka.Message

func (r *Consumer) wrapProcessor(wrapFn Interceptor) {
	r.processor = func(fn processFn) error {
		return wrapFn(fn)(&cmd{req: make([]interface{}, 0, 1)})
	}
}

func (r *Consumer) Close() error {
	return r.processor(func(c *cmd) error {
		logCmd(r.logMode, c, "ConsumerClose")
		return r.r.Close()
	})
}

func (r *Consumer) CommitMessages(ctx context.Context, msgs ...Message) (err error) {
	return r.processor(func(c *cmd) error {
		logCmd(r.logMode, c, "CommitMessages", cmdWithContext(ctx), cmdWithReq(msgs))
		return r.r.CommitMessages(ctx, msgs...)
	})
}

func (r *Consumer) FetchMessage(ctx context.Context) (msg Message, err error) {
	err = r.processor(func(c *cmd) error {
		msg, err = r.r.FetchMessage(ctx)
		logCmd(r.logMode, c, "FetchMessage", cmdWithContext(ctx), cmdWithRes(msg))
		return err
	})
	return
}

func (r *Consumer) Lag() int64 {
	return r.r.Lag()
}

func (r *Consumer) Offset() int64 {
	return r.r.Offset()
}

func (r *Consumer) ReadLag(ctx context.Context) (lag int64, err error) {
	err = r.processor(func(c *cmd) error {
		lag, err = r.r.ReadLag(ctx)
		logCmd(r.logMode, c, "ReadLag", cmdWithContext(ctx))
		return err
	})
	return
}

func (r *Consumer) ReadMessage(ctx context.Context) (msg Message, err error) {
	err = r.processor(func(c *cmd) error {
		msg, err = r.r.ReadMessage(ctx)
		logCmd(r.logMode, c, "ReadMessage", cmdWithContext(ctx), cmdWithRes(msg))
		return err
	})
	return
}

func (r *Consumer) SetOffset(offset int64) (err error) {
	return r.processor(func(c *cmd) error {
		logCmd(r.logMode, c, "SetOffset")
		return r.r.SetOffset(offset)
	})
}

func (r *Consumer) SetOffsetAt(ctx context.Context, t time.Time) (err error) {
	return r.processor(func(c *cmd) error {
		logCmd(r.logMode, c, "SetOffsetAt", cmdWithContext(ctx))
		return r.r.SetOffsetAt(ctx, t)
	})
}
