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

type Client struct {
	cc        *mongo.Client
	processor processor
	logMode   bool
}

func NewClient(opts ...*options.ClientOptions) (*Client, error) {
	client, err := mongo.NewClient(opts...)
	if err != nil {
		return nil, err
	}
	return &Client{cc: client, processor: defaultProcessor}, nil
}

func (wc *Client) setLogMode(logMode bool) {
	wc.logMode = logMode
}

func defaultProcessor(processFn processFn) error {
	return processFn(&cmd{req: make([]interface{}, 0, 1)})
}

func Connect(ctx context.Context, opts ...*options.ClientOptions) (wc *Client, err error) {
	cc, err := mongo.NewClient(opts...)
	if err != nil {
		return nil, err
	}

	wc = &Client{cc: cc, processor: defaultProcessor}
	err = wc.processor(func(c *cmd) error {
		return wc.Connect(ctx)
	})
	return
}

func (wc *Client) wrapProcessor(wrapFn func(processFn) processFn) {
	wc.processor = func(fn processFn) error {
		return wrapFn(fn)(&cmd{req: make([]interface{}, 0, 1)})
	}
}

func (wc *Client) Connect(ctx context.Context) error {
	return wc.processor(func(c *cmd) error {
		logCmd(wc.logMode, c, "Connect", nil)
		return wc.cc.Connect(ctx)
	})
}

func (wc *Client) Database(name string, opts ...*options.DatabaseOptions) *Database {
	var db *mongo.Database
	_ = wc.processor(func(c *cmd) error {
		db = wc.cc.Database(name, opts...)
		logCmd(wc.logMode, c, "Connect", db, name)
		return nil
	})
	if db == nil {
		return nil
	}
	return &Database{db: db, processor: wc.processor, logMode: wc.logMode}
}

func (wc *Client) Disconnect(ctx context.Context) error {
	return wc.processor(func(c *cmd) error {
		logCmd(wc.logMode, c, "Disconnect", nil)
		return wc.cc.Disconnect(ctx)
	})
}

func (wc *Client) ListDatabaseNames(ctx context.Context, filter interface{}, opts ...*options.ListDatabasesOptions) (
	dbs []string, err error) {

	err = wc.processor(func(c *cmd) error {
		dbs, err = wc.cc.ListDatabaseNames(ctx, filter, opts...)
		logCmd(wc.logMode, c, "ListDatabaseNames", dbs, filter)
		return err
	})
	return
}

func (wc *Client) ListDatabases(ctx context.Context, filter interface{}, opts ...*options.ListDatabasesOptions) (
	dbr mongo.ListDatabasesResult, err error) {

	err = wc.processor(func(c *cmd) error {
		dbr, err = wc.cc.ListDatabases(ctx, filter, opts...)
		logCmd(wc.logMode, c, "ListDatabases", dbr, filter)
		return err
	})
	return
}

func (wc *Client) Ping(ctx context.Context, rp *readpref.ReadPref) error {
	return wc.processor(func(c *cmd) error {
		logCmd(wc.logMode, c, "Ping", nil, rp)
		return wc.cc.Ping(ctx, rp)
	})
}

func (wc *Client) StartSession(opts ...*options.SessionOptions) (ss mongo.Session, err error) {
	err = wc.processor(func(c *cmd) error {
		ss, err = wc.cc.StartSession(opts...)
		logCmd(wc.logMode, c, "StartSession", ss)
		return err
	})
	return &Session{Session: ss, logMode: wc.logMode, processor: wc.processor}, nil
}

func (wc *Client) UseSession(ctx context.Context, fn func(mongo.SessionContext) error) error {
	return wc.processor(func(c *cmd) error {
		logCmd(wc.logMode, c, "UseSession", nil)
		return wc.cc.UseSession(ctx, fn)
	})
}

func (wc *Client) UseSessionWithOptions(ctx context.Context, opts *options.SessionOptions, fn func(mongo.SessionContext) error) error {
	return wc.processor(func(c *cmd) error {
		logCmd(wc.logMode, c, "UseSessionWithOptions", nil)
		return wc.cc.UseSessionWithOptions(ctx, opts, fn)
	})
}

func (wc *Client) Client() *mongo.Client { return wc.cc }
