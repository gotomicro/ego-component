package ekafka

import (
	"context"

	"github.com/segmentio/kafka-go"
)

type Client struct {
	cc        *kafka.Client
	processor processor
	logMode   bool
}

func defaultProcessor(processFn processFn) error {
	return processFn(&cmd{req: make([]interface{}, 0, 1)})
}

type cmdOptsFunc func(logMode bool, c *cmd)

func cmdWithRes(res interface{}) cmdOptsFunc {
	return func(logMode bool, c *cmd) {
		// 只有开启log模式才会记录 res
		if logMode {
			c.res = res
		}
	}
}

func cmdWithReq(req interface{}) cmdOptsFunc {
	return func(logMode bool, c *cmd) {
		// 只有开启log模式才会记录 req
		if logMode {
			c.req = req
		}
	}
}

func cmdWithContext(ctx context.Context) cmdOptsFunc {
	return func(logMode bool, c *cmd) {
		c.ctx = ctx
	}
}

func logCmd(logMode bool, c *cmd, name string, opts ...cmdOptsFunc) {
	c.name = name
	c.ctx = context.Background()

	for _, opt := range opts {
		opt(logMode, c)
	}
}

func (wc *Client) wrapProcessor(wrapFn func(processFn) processFn) {
	wc.processor = func(fn processFn) error {
		return wrapFn(fn)(&cmd{req: make([]interface{}, 0, 1)})
	}
}

func (wc *Client) DeleteTopics(ctx context.Context, req *kafka.DeleteTopicsRequest) (res *kafka.DeleteTopicsResponse, err error) {
	err = wc.processor(func(c *cmd) error {
		logCmd(wc.logMode, c, "DeleteTopics")
		res, err = wc.cc.DeleteTopics(ctx, req)
		return err
	})
	return
}

func (wc *Client) ListOffsets(ctx context.Context, req *kafka.ListOffsetsRequest) (res *kafka.ListOffsetsResponse, err error) {
	err = wc.processor(func(c *cmd) error {
		logCmd(wc.logMode, c, "ListOffsets")
		res, err = wc.cc.ListOffsets(ctx, req)
		return err
	})
	return
}

func (wc *Client) OffsetFetch(ctx context.Context, req *kafka.OffsetFetchRequest) (res *kafka.OffsetFetchResponse, err error) {
	err = wc.processor(func(c *cmd) error {
		logCmd(wc.logMode, c, "OffsetFetch")
		res, err = wc.cc.OffsetFetch(ctx, req)
		return err
	})
	return
}

func (wc *Client) Metadata(ctx context.Context, req *kafka.MetadataRequest) (res *kafka.MetadataResponse, err error) {
	err = wc.processor(func(c *cmd) error {
		logCmd(wc.logMode, c, "Metadata")
		res, err = wc.cc.Metadata(ctx, req)
		return err
	})
	return
}

func (wc *Client) CreateTopics(ctx context.Context, req *kafka.CreateTopicsRequest) (res *kafka.CreateTopicsResponse, err error) {
	err = wc.processor(func(c *cmd) error {
		logCmd(wc.logMode, c, "CreateTopics")
		res, err = wc.cc.CreateTopics(ctx, req)
		return err
	})
	return
}
