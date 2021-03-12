package consumerserver

import "github.com/segmentio/kafka-go"

// SingleMessageHandler ...
type SingleMessageHandler = func(message kafka.Message) error
