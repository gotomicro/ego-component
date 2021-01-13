package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/gotomicro/ego/core/econf"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/gotomicro/ego-component/emongo"
)

func main() {
	var stopCh = make(chan bool)
	// 假设你配置的toml如下所示
	conf := `
[mongo]
	debug=true
	dsn="mongodb://user:password@localhost:27017,localhost:27018"
`
	// 加载配置文件
	err := econf.LoadFromReader(strings.NewReader(conf), toml.Unmarshal)
	if err != nil {
		panic("LoadFromReader fail," + err.Error())
	}

	// 初始化emongo组件
	cmp := emongo.Load("mongo").Build()
	coll := cmp.Client.Database("test").Collection("cells")
	findOne(coll)

	stopCh <- true
}

func findOne(coll *emongo.Collection) {
	res := coll.FindOne(context.TODO(), bson.M{"row_id": "10000000001"}, options.FindOne().SetBatchSize(1024))
	var result bson.M
	err := res.Decode(&result)
	if err != nil {
		fmt.Println("err occurs", err)
	}
	fmt.Println("result is", result)
}
