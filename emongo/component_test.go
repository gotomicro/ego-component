package emongo

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func newColl() *Collection {
	cmp := DefaultContainer().Build(WithDSN(os.Getenv("EMONGO_DSN")))
	coll := cmp.Client().Database("test").Collection("cells")
	return coll
}

func TestWrappedCollection_FindOne(t *testing.T) {
	coll := newColl()
	res := coll.FindOne(context.TODO(), bson.M{"row_id": "10000000001"}, options.FindOne().SetBatchSize(1024))
	var result bson.M
	err := res.Decode(&result)
	t.Log(result)
	assert.NoError(t, err)
}

func TestWrappedCollection_Find(t *testing.T) {
	var ctx = context.TODO()
	coll := newColl()
	cur, err := coll.Find(ctx, bson.M{"row_id": "10000000001", "table_id": "ZUwjubnYEg"}, options.Find().SetBatchSize(1024))
	assert.NoError(t, err)

	for cur.Next(ctx) {
		var result bson.M
		err := cur.Decode(&result)
		if err != nil {
			t.Fatal(err)
		}
		// do something with result....
		t.Log(result)
	}
}

func TestSession(t *testing.T) {
	var ctx = context.TODO()
	client := DefaultContainer().Build(WithDSN(os.Getenv("EMONGO_DSN"))).Client()
	sess, err := client.StartSession()
	assert.NoError(t, err)
	defer sess.EndSession(context.Background())

	coll := client.Database("foo").Collection("bar")
	defer func() {
		_ = coll.Drop(ctx)
	}()
	_, err = sess.WithTransaction(context.Background(), func(sessCtx mongo.SessionContext) (interface{}, error) {
		res := coll.FindOne(sessCtx, bson.D{{"x", 1}})
		return res, err
	})
	assert.NotNil(t, err, "expected WithTransaction error, got nil")
}
