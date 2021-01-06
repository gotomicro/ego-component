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
	"sync"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readconcern"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
)

type WrappedDatabase struct {
	mu        sync.Mutex
	db        *mongo.Database
	processor processor
}

func (wd *WrappedDatabase) Client() *WrappedClient {
	wd.mu.Lock()
	defer wd.mu.Unlock()

	cc := wd.db.Client()
	if cc == nil {
		return nil
	}
	return &WrappedClient{cc: cc}
}

func (wd *WrappedDatabase) Collection(name string, opts ...*options.CollectionOptions) *WrappedCollection {
	if wd.db == nil {
		return nil
	}
	coll := wd.db.Collection(name, opts...)
	if coll == nil {
		return nil
	}
	return &WrappedCollection{coll: coll, processor: wd.processor}
}

func (wd *WrappedDatabase) Drop(ctx context.Context) error {
	return wd.processor(func() error {
		return wd.db.Drop(ctx)
	})
}

func (wd *WrappedDatabase) ListCollections(ctx context.Context, filter interface{}, opts ...*options.ListCollectionsOptions) (
	cur *mongo.Cursor, err error) {
	err = wd.processor(func() error {
		cur, err = wd.db.ListCollections(ctx, filter, opts...)
		return err
	})
	return
}

func (wd *WrappedDatabase) Name() string                          { return wd.db.Name() }
func (wd *WrappedDatabase) ReadConcern() *readconcern.ReadConcern { return wd.db.ReadConcern() }
func (wd *WrappedDatabase) ReadPreference() *readpref.ReadPref    { return wd.db.ReadPreference() }

func (wd *WrappedDatabase) RunCommand(ctx context.Context, runCommand interface{}, opts ...*options.RunCmdOptions) (res *mongo.SingleResult) {
	_ = wd.processor(func() error {
		res = wd.db.RunCommand(ctx, runCommand, opts...)
		return res.Err()
	})
	return
}

func (wd *WrappedDatabase) WriteConcern() (res *writeconcern.WriteConcern) {
	_ = wd.processor(func() error {
		res = wd.db.WriteConcern()
		return nil
	})
	return
}

func (wd *WrappedDatabase) Database() *mongo.Database {
	return wd.db
}
