package ekafka

import (
	"context"

	"github.com/segmentio/kafka-go"
)

type Client struct {
	cc        *kafka.Client
	processor ClientInterceptor
	logMode   bool
}

type cmdOptsFunc func(logMode bool, c *cmd)

func cmdWithMsg(res Message) cmdOptsFunc {
	return func(logMode bool, c *cmd) {
		c.msg = res
	}
}

func cmdWithRes(res interface{}) cmdOptsFunc {
	return func(logMode bool, c *cmd) {
		// 只有开启log模式才会记录 res
		if logMode {
			c.res = res
		}
	}
}

func logCmd(logMode bool, c *cmd, name string, opts ...cmdOptsFunc) {
	c.name = name

	for _, opt := range opts {
		opt(logMode, c)
	}
}

func (wc *Client) wrapProcessor(p ClientInterceptor) {
	wc.processor = p
}

func (wc *Client) DeleteTopics(ctx context.Context, req *kafka.DeleteTopicsRequest) (res *kafka.DeleteTopicsResponse, err error) {
	err = wc.processor(func(ctx context.Context, msgs Messages, c *cmd) error {
		logCmd(wc.logMode, c, "DeleteTopics")
		res, err = wc.cc.DeleteTopics(ctx, req)
		return err
	})(ctx, nil, &cmd{})
	return
}

func (wc *Client) ListOffsets(ctx context.Context, req *kafka.ListOffsetsRequest) (res *kafka.ListOffsetsResponse, err error) {
	err = wc.processor(func(ctx context.Context, msgs Messages, c *cmd) error {
		logCmd(wc.logMode, c, "ListOffsets")
		res, err = wc.cc.ListOffsets(ctx, req)
		return err
	})(ctx, nil, &cmd{})
	return
}

func (wc *Client) OffsetFetch(ctx context.Context, req *kafka.OffsetFetchRequest) (res *kafka.OffsetFetchResponse, err error) {
	err = wc.processor(func(ctx context.Context, msgs Messages, c *cmd) error {
		logCmd(wc.logMode, c, "OffsetFetch")
		res, err = wc.cc.OffsetFetch(ctx, req)
		return err
	})(ctx, nil, &cmd{})
	return
}

func (wc *Client) Metadata(ctx context.Context, req *kafka.MetadataRequest) (res *kafka.MetadataResponse, err error) {
	err = wc.processor(func(ctx context.Context, msgs Messages, c *cmd) error {
		logCmd(wc.logMode, c, "Metadata")
		res, err = wc.cc.Metadata(ctx, req)
		return err
	})(ctx, nil, &cmd{})
	return
}

func (wc *Client) CreateTopics(ctx context.Context, req *kafka.CreateTopicsRequest) (res *kafka.CreateTopicsResponse, err error) {
	err = wc.processor(func(ctx context.Context, msgs Messages, c *cmd) error {
		logCmd(wc.logMode, c, "CreateTopics")
		res, err = wc.cc.CreateTopics(ctx, req)
		return err
	})(ctx, nil, &cmd{})
	return
}
