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

type ClientEncryption struct {
	cc        *mongo.ClientEncryption
	processor processor
	logMode   bool
}

func (wc *Client) NewClientEncryption(opts ...*options.ClientEncryptionOptions) (*ClientEncryption, error) {
	client, err := mongo.NewClientEncryption(wc.Client(), opts...)
	if err != nil {
		return nil, err
	}
	return &ClientEncryption{cc: client, processor: defaultProcessor, logMode: wc.logMode}, nil
}

func (wce *ClientEncryption) CreateDataKey(ctx context.Context, kmsProvider string, opts ...*options.DataKeyOptions) (
	id primitive.Binary, err error) {

	err = wce.processor(func(c *cmd) error {
		id, err = wce.cc.CreateDataKey(ctx, kmsProvider, opts...)
		logCmd(wce.logMode, c, "CreateDataKey", id)
		return err
	})
	return
}

func (wce *ClientEncryption) Encrypt(ctx context.Context, val bson.RawValue, opts ...*options.EncryptOptions) (
	value primitive.Binary, err error) {

	err = wce.processor(func(c *cmd) error {
		value, err = wce.cc.Encrypt(ctx, val, opts...)
		logCmd(wce.logMode, c, "Encrypt", value, val)
		return err
	})
	return
}

func (wce *ClientEncryption) Decrypt(ctx context.Context, val primitive.Binary) (value bson.RawValue, err error) {
	err = wce.processor(func(c *cmd) error {
		value, err = wce.cc.Decrypt(ctx, val)
		logCmd(wce.logMode, c, "Decrypt", value, val)
		return err
	})
	return
}

func (wce *ClientEncryption) Close(ctx context.Context) error {
	return wce.processor(func(c *cmd) error {
		logCmd(wce.logMode, c, "Close", nil)
		return wce.cc.Close(ctx)
	})
}
