package ekafka

import (
	"context"

	"github.com/segmentio/kafka-go"
)

type Producer struct {
	w         *kafka.Writer
	processor ClientInterceptor
	logMode   bool
}

func (p *Producer) setProcessor(c ClientInterceptor) {
	p.processor = c
}

func (p *Producer) Close() error {
	return p.processor(func(ctx context.Context, msgs Messages, c *cmd) error {
		logCmd(p.logMode, c, "ProducerClose")
		return p.w.Close()
	})(context.Background(), nil, &cmd{})
}

func (p *Producer) WriteMessages(ctx context.Context, msgs ...*Message) error {
	return p.processor(func(ctx context.Context, req Messages, c *cmd) error {
		logCmd(p.logMode, c, "WriteMessages")
		return p.w.WriteMessages(ctx, req.ToNoPointer()...)
	})(ctx, msgs, &cmd{})
}
