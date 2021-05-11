package e2e

import (
	"context"
	"time"

	"github.com/gotomicro/ego-component/ekafka"
)

func writeMessage(producer *ekafka.Producer, message string, errCh chan<- error) {
	writeCtx, _ := context.WithTimeout(context.Background(), 30*time.Second)
	err := producer.WriteMessages(
		writeCtx,
		ekafka.Message{Value: []byte(message)},
	)
	if err != nil {
		errCh <- err
		return
	}

	if err := producer.Close(); err != nil {
		errCh <- err
		return
	}
}
