package ekafka

import (
	"context"
	"fmt"
	"log"
	"strings"
	"testing"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/gotomicro/ego/core/econf"
	"github.com/stretchr/testify/assert"
)

func produce(w *Producer) {
	err := w.WriteMessages(context.Background(),
		&Message{Key: []byte("Key-A"), Value: []byte("Hello World!")},
		&Message{Key: []byte("Key-B"), Value: []byte("One!")},
		&Message{Key: []byte("Key-C"), Value: []byte("Two!")},
	)
	if err != nil {
		log.Fatal("failed to write messages:", err)
	}
	if err := w.Close(); err != nil {
		log.Fatal("failed to close writer:", err)
	}
	fmt.Println(`produce message succ--------------->`)
}

func consume(r *Consumer) {
	ctx := context.Background()
	for {
		// the `ReadMessage` method blocks until we receive the next event
		msg, _, err := r.ReadMessage(ctx)
		if err != nil {
			panic("could not read message " + err.Error())
		}
		// after receiving the message, log its value
		fmt.Println("received: ", string(msg.Value))
		err = r.CommitMessages(ctx, &msg)
		if err != nil {
			log.Printf("fail to commit msg:%v", err)
		}
	}
}

func TestProduceConsume(t *testing.T) {
	conf := `
[kafka]
	debug=true
	brokers=["localhost:9091","localhost:9092","localhost:9093"]
	[kafka.client]
        timeout="3s"
	[kafka.producers.p1]
		topic="sre-infra-test"
	[kafka.consumers.c1]
		topic="sre-infra-test"
		groupID="group-1"
	[kafka.consumers.c2]
		topic="sre-infra-test"
		groupID="group-2"
`
	err := econf.LoadFromReader(strings.NewReader(conf), toml.Unmarshal)
	assert.NoError(t, err)
	cmp := Load("kafka").Build()
	go produce(cmp.Producer("p1"))
	consume(cmp.Consumer("c1"))

	time.Sleep(60 * time.Second)
}
