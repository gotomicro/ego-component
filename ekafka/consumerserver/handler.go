package consumerserver

import (
	"context"

	"github.com/gotomicro/ego-component/ekafka"
	"github.com/segmentio/kafka-go"
)

// OnEachMessageHandler ...
type OnEachMessageHandler = func(ctx context.Context, message kafka.Message) error

// OnStartHandler ...
type OnStartHandler = func(ctx context.Context, consumer *ekafka.Consumer) error
