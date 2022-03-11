// Copyright 2018, OpenCensus Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package emongo

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Session = mongo.Session
type SessionContext = mongo.SessionContext

type session struct {
	mongo.Session
	processor processor
	logMode   bool
}

var _ mongo.Session = (*session)(nil)

//	WithTransaction(ctx context.Context, fn func(sessCtx SessionContext) (interface{}, error),
//		opts ...*options.TransactionOptions) (interface{}, error)
//	Client() *Client
//	ID() bson.Raw

func (ws *session) EndSession(ctx context.Context) {
	_ = ws.processor(func(c *cmd) error {
		ws.Session.EndSession(ctx)
		logCmd(ws.logMode, c, "EndSession", nil)
		return nil
	})
}

func (ws *session) StartTransaction(topts ...*options.TransactionOptions) error {
	return ws.processor(func(c *cmd) error {
		logCmd(ws.logMode, c, "StartTransaction", nil)
		return ws.Session.StartTransaction(topts...)
	})
}

func (ws *session) AbortTransaction(ctx context.Context) error {
	return ws.processor(func(c *cmd) error {
		logCmd(ws.logMode, c, "AbortTransaction", nil)
		return ws.Session.AbortTransaction(ctx)
	})
}

func (ws *session) CommitTransaction(ctx context.Context) error {
	return ws.processor(func(c *cmd) error {
		logCmd(ws.logMode, c, "CommitTransaction", nil)
		return ws.Session.CommitTransaction(ctx)
	})
}

func (ws *session) ClusterTime() (raw bson.Raw) {
	_ = ws.processor(func(c *cmd) error {
		raw = ws.Session.ClusterTime()
		logCmd(ws.logMode, c, "ClusterTime", raw)
		return nil
	})
	return
}

func (ws *session) AdvanceClusterTime(br bson.Raw) error {
	return ws.processor(func(c *cmd) error {
		logCmd(ws.logMode, c, "AdvanceClusterTime", nil)
		return ws.Session.AdvanceClusterTime(br)
	})
}

func (ws *session) OperationTime() (ts *primitive.Timestamp) {
	_ = ws.processor(func(c *cmd) error {
		ts = ws.Session.OperationTime()
		logCmd(ws.logMode, c, "OperationTime", ts)
		return nil
	})
	return
}

func (ws *session) AdvanceOperationTime(pt *primitive.Timestamp) error {
	return ws.processor(func(c *cmd) error {
		logCmd(ws.logMode, c, "AdvanceOperationTime", nil)
		return ws.Session.AdvanceOperationTime(pt)
	})
}

func (ws *session) WithTransaction(ctx context.Context, fn func(sessCtx SessionContext) (interface{}, error),
	opts ...*options.TransactionOptions) (out interface{}, err error) {
	err = ws.processor(func(c *cmd) error {
		logCmd(ws.logMode, c, "WithTransaction", nil)
		out, err = ws.Session.WithTransaction(ctx, fn, opts...)
		return err
	})
	return
}
