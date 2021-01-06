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

type WrappedSession struct {
	mongo.Session
	processor processor
}

var _ mongo.Session = (*WrappedSession)(nil)

func (ws *WrappedSession) EndSession(ctx context.Context) {
	_ = ws.processor(func() error {
		ws.Session.EndSession(ctx)
		return nil
	})
}

func (ws *WrappedSession) StartTransaction(topts ...*options.TransactionOptions) error {
	return ws.processor(func() error {
		return ws.Session.StartTransaction(topts...)
	})
}

func (ws *WrappedSession) AbortTransaction(ctx context.Context) error {
	return ws.processor(func() error {
		return ws.Session.AbortTransaction(ctx)
	})
}

func (ws *WrappedSession) CommitTransaction(ctx context.Context) error {
	return ws.processor(func() error {
		return ws.Session.CommitTransaction(ctx)
	})
}

func (ws *WrappedSession) ClusterTime() (raw bson.Raw) {
	_ = ws.processor(func() error {
		raw = ws.Session.ClusterTime()
		return nil
	})
	return
}

func (ws *WrappedSession) AdvanceClusterTime(br bson.Raw) error {
	return ws.processor(func() error {
		return ws.Session.AdvanceClusterTime(br)
	})
}

func (ws *WrappedSession) OperationTime() (ts *primitive.Timestamp) {
	_ = ws.processor(func() error {
		ts = ws.Session.OperationTime()
		return nil
	})
	return
}

func (ws *WrappedSession) AdvanceOperationTime(pt *primitive.Timestamp) error {
	return ws.processor(func() error {
		return ws.Session.AdvanceOperationTime(pt)
	})
}
