package eemqtt

import (
	"context"
	"github.com/eclipse/paho.golang/paho"
)

type Message struct {
	Count uint64
}

//订推消息处理
type OnPublishHandler = func(ctx context.Context, pp *paho.Publish)
