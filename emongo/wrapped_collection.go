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
)

type processor func(fn processFn) error
type processFn func() error

type WrappedCollection struct {
	coll      *mongo.Collection
	processor processor
}

func (wc *WrappedCollection) Aggregate(ctx context.Context, pipeline interface{}, opts ...*options.AggregateOptions) (cur *mongo.Cursor, err error) {
	err = wc.processor(func() error {
		cur, err = wc.coll.Aggregate(ctx, pipeline, opts...)
		return err
	})
	return
}

func (wc *WrappedCollection) BulkWrite(ctx context.Context, models []mongo.WriteModel, opts ...*options.BulkWriteOptions) (
	bwres *mongo.BulkWriteResult, err error) {

	err = wc.processor(func() error {
		bwres, err = wc.coll.BulkWrite(ctx, models, opts...)
		return err
	})
	return
}

func (wc *WrappedCollection) Clone(opts ...*options.CollectionOptions) (coll *mongo.Collection, err error) {
	err = wc.processor(func() error {
		coll, err = wc.coll.Clone(opts...)
		return err
	})
	return
}

func (wc *WrappedCollection) CountDocuments(ctx context.Context, filter interface{}, opts ...*options.CountOptions) (count int64, err error) {
	err = wc.processor(func() error {
		count, err = wc.coll.CountDocuments(ctx, filter, opts...)
		return err
	})
	return count, err
}

func (wc *WrappedCollection) Database() *mongo.Database { return wc.coll.Database() }

func (wc *WrappedCollection) DeleteMany(ctx context.Context, filter interface{}, opts ...*options.DeleteOptions) (
	res *mongo.DeleteResult, err error) {

	err = wc.processor(func() error {
		res, err = wc.coll.DeleteMany(ctx, filter, opts...)
		return err
	})
	return
}

func (wc *WrappedCollection) DeleteOne(ctx context.Context, filter interface{}, opts ...*options.DeleteOptions) (res *mongo.DeleteResult, err error) {
	err = wc.processor(func() error {
		res, err = wc.coll.DeleteOne(ctx, filter, opts...)
		return err
	})
	return
}

func (wc *WrappedCollection) Distinct(ctx context.Context, fieldName string, filter interface{}, opts ...*options.DistinctOptions) (res []interface{}, err error) {
	err = wc.processor(func() error {
		res, err = wc.coll.Distinct(ctx, fieldName, filter, opts...)
		return err
	})
	return
}

func (wc *WrappedCollection) Drop(ctx context.Context) error {
	return wc.processor(func() error {
		return wc.coll.Drop(ctx)
	})
}

func (wc *WrappedCollection) EstimatedDocumentCount(ctx context.Context, opts ...*options.EstimatedDocumentCountOptions) (res int64, err error) {
	err = wc.processor(func() error {
		res, err = wc.coll.EstimatedDocumentCount(ctx, opts...)
		return err
	})
	return
}

func (wc *WrappedCollection) Find(ctx context.Context, filter interface{}, opts ...*options.FindOptions) (cur *mongo.Cursor, err error) {
	err = wc.processor(func() error {
		cur, err = wc.coll.Find(ctx, filter, opts...)
		return err
	})
	return
}

func (wc *WrappedCollection) FindOne(ctx context.Context, filter interface{}, opts ...*options.FindOneOptions) (res *mongo.SingleResult) {
	_ = wc.processor(func() error {
		res = wc.coll.FindOne(ctx, filter, opts...)
		return res.Err()
	})
	return
}

func (wc *WrappedCollection) FindOneAndDelete(ctx context.Context, filter interface{}, opts ...*options.FindOneAndDeleteOptions) (res *mongo.SingleResult) {
	_ = wc.processor(func() error {
		res = wc.coll.FindOneAndDelete(ctx, filter, opts...)
		return res.Err()
	})
	return
}

func (wc *WrappedCollection) FindOneAndReplace(ctx context.Context, filter, replacement interface{}, opts ...*options.FindOneAndReplaceOptions) (res *mongo.SingleResult) {
	_ = wc.processor(func() error {
		res = wc.coll.FindOneAndReplace(ctx, filter, replacement, opts...)
		return res.Err()
	})
	return
}

func (wc *WrappedCollection) FindOneAndUpdate(ctx context.Context, filter, update interface{}, opts ...*options.FindOneAndUpdateOptions) (res *mongo.SingleResult) {
	_ = wc.processor(func() error {
		res = wc.coll.FindOneAndUpdate(ctx, filter, update, opts...)
		return res.Err()
	})
	return
}

func (wc *WrappedCollection) Indexes() mongo.IndexView { return wc.coll.Indexes() }

func (wc *WrappedCollection) InsertMany(ctx context.Context, documents []interface{}, opts ...*options.InsertManyOptions) (res *mongo.InsertManyResult, err error) {
	_ = wc.processor(func() error {
		res, err = wc.coll.InsertMany(ctx, documents, opts...)
		return err
	})
	return
}

func (wc *WrappedCollection) InsertOne(ctx context.Context, document interface{}, opts ...*options.InsertOneOptions) (res *mongo.InsertOneResult, err error) {
	_ = wc.processor(func() error {
		res, err = wc.coll.InsertOne(ctx, document, opts...)
		return err
	})
	return
}

func (wc *WrappedCollection) Name() string { return wc.coll.Name() }

func (wc *WrappedCollection) ReplaceOne(ctx context.Context, filter, replacement interface{}, opts ...*options.ReplaceOptions) (res *mongo.UpdateResult, err error) {
	_ = wc.processor(func() error {
		res, err = wc.coll.ReplaceOne(ctx, filter, replacement, opts...)
		return err
	})
	return
}

func (wc *WrappedCollection) UpdateMany(ctx context.Context, filter, replacement interface{}, opts ...*options.UpdateOptions) (res *mongo.UpdateResult, err error) {
	_ = wc.processor(func() error {
		res, err = wc.coll.UpdateMany(ctx, filter, replacement, opts...)
		return err
	})
	return
}

func (wc *WrappedCollection) UpdateOne(ctx context.Context, filter, replacement interface{}, opts ...*options.UpdateOptions) (res *mongo.UpdateResult, err error) {
	_ = wc.processor(func() error {
		res, err = wc.coll.UpdateOne(ctx, filter, replacement, opts...)
		return err
	})
	return
}

func (wc *WrappedCollection) Watch(ctx context.Context, pipeline interface{}, opts ...*options.ChangeStreamOptions) (cs *mongo.ChangeStream, err error) {
	_ = wc.processor(func() error {
		cs, err = wc.coll.Watch(ctx, pipeline, opts...)
		return err
	})
	return
}

func (wc *WrappedCollection) Collection() *mongo.Collection {
	return wc.coll
}
