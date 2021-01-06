package emongo

import (
	"context"
	"fmt"
	"testing"

	"go.mongodb.org/mongo-driver/bson"
)

type obj struct {
	RowID string `json:"row_id"`
}

func TestWrappedCollection_FindOne(t *testing.T) {
	cmp := DefaultContainer().Build(WithDSN("mongodb://172.16.20.8:27017"))

	coll := cmp.Client.Database("test").Collection("cells")
	res := coll.FindOne(context.TODO(), bson.M{"row_id": "10000000001"})
	var result bson.M
	err := res.Decode(&result)
	fmt.Println(`err--------------->`, result, err)
}
