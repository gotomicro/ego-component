package emongo

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
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
