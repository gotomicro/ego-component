package ekafka

import (
	"context"

	"github.com/segmentio/kafka-go"
)

type Producer struct {
	w         *kafka.Writer
	processor processor
	logMode   bool
}

func (w *Producer) wrapProcessor(wrapFn func(processFn) processFn) {
	w.processor = func(fn processFn) error {
		return wrapFn(fn)(&cmd{req: make([]interface{}, 0, 1), ctx: context.Background()})
	}
}

func (w *Producer) Close() error {
	return w.processor(func(c *cmd) error {
		logCmd(w.logMode, c, "ProducerClose", nil)
		return w.w.Close()
	})
}

func (w *Producer) WriteMessages(ctx context.Context, msgs ...Message) error {
	return w.processor(func(c *cmd) error {
		logCmd(w.logMode, c, "WriteMessages", nil, msgs)
		return w.w.WriteMessages(ctx, msgs...)
	})
}
