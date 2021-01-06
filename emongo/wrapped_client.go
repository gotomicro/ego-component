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

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type WrappedClient struct {
	cc        *mongo.Client
	processor processor
}

func NewClient(opts ...*options.ClientOptions) (*WrappedClient, error) {
	client, err := mongo.NewClient(opts...)
	if err != nil {
		return nil, err
	}
	return &WrappedClient{cc: client, processor: defaultProcessor}, nil
}

func defaultProcessor(processFn processFn) error {
	return processFn()
}

func Connect(ctx context.Context, opts ...*options.ClientOptions) (wc *WrappedClient, err error) {
	cc, err := mongo.NewClient(opts...)
	if err != nil {
		return nil, err
	}

	wc = &WrappedClient{cc: cc, processor: defaultProcessor}
	err = wc.processor(func() error {
		return wc.Connect(ctx)
	})
	return
}

func (wc *WrappedClient) wrapProcess(fn func(oldProcessFn processFn) (newProcessFn processFn)) {
	wc.processor = func(fn processFn) error {
		return fn()
	}
}

func (wc *WrappedClient) Connect(ctx context.Context) error {
	return wc.processor(func() error {
		return wc.cc.Connect(ctx)
	})
}

func (wc *WrappedClient) Database(name string, opts ...*options.DatabaseOptions) *WrappedDatabase {
	var db *mongo.Database
	_ = wc.processor(func() error {
		db = wc.cc.Database(name, opts...)
		return nil
	})
	if db == nil {
		return nil
	}
	// return &WrappedDatabase{db: db, processor: func(processFn) error { return wc.processFn() }}
	return &WrappedDatabase{db: db, processor: wc.processor}
}

func (wc *WrappedClient) Disconnect(ctx context.Context) error {
	return wc.processor(func() error {
		return wc.cc.Disconnect(ctx)
	})
}

func (wc *WrappedClient) ListDatabaseNames(ctx context.Context, filter interface{}, opts ...*options.ListDatabasesOptions) (
	dbs []string, err error) {

	err = wc.processor(func() error {
		dbs, err = wc.cc.ListDatabaseNames(ctx, filter, opts...)
		return err
	})
	return
}

func (wc *WrappedClient) ListDatabases(ctx context.Context, filter interface{}, opts ...*options.ListDatabasesOptions) (
	dbr mongo.ListDatabasesResult, err error) {

	err = wc.processor(func() error {
		dbr, err = wc.cc.ListDatabases(ctx, filter, opts...)
		return err
	})
	return
}

func (wc *WrappedClient) Ping(ctx context.Context, rp *readpref.ReadPref) error {
	return wc.processor(func() error {
		return wc.cc.Ping(ctx, rp)
	})
}

func (wc *WrappedClient) StartSession(opts ...*options.SessionOptions) (ss mongo.Session, err error) {
	err = wc.processor(func() error {
		ss, err = wc.cc.StartSession(opts...)
		return err
	})
	return &WrappedSession{Session: ss}, nil
}

func (wc *WrappedClient) UseSession(ctx context.Context, fn func(mongo.SessionContext) error) error {
	return wc.processor(func() error {
		return wc.cc.UseSession(ctx, fn)
	})
}

func (wc *WrappedClient) UseSessionWithOptions(ctx context.Context, opts *options.SessionOptions, fn func(mongo.SessionContext) error) error {
	return wc.processor(func() error {
		return wc.cc.UseSessionWithOptions(ctx, opts, fn)
	})
}

func (wc *WrappedClient) Client() *mongo.Client { return wc.cc }
