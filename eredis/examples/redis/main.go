package main

import (
	"fmt"
	"github.com/gotomicro/ego"
	"github.com/gotomicro/ego-component/eredis"
	"github.com/gotomicro/ego/core/elog"
)

// export EGO_DEBUG=true && go run main.go --config=config.toml
func main() {
	err := ego.New().Invoker(
		invokerRedis,
		testRedis,
	).Run()
	if err != nil {
		elog.Panic("startup", elog.FieldErr(err))
	}
}

var eredisClient *eredis.Component

func invokerRedis() error {
	eredisClient = eredis.Load("redis.test").Build(eredis.WithStub())
	return nil
}

func testRedis() error {
	err := eredisClient.Set("hello", "world", 0)
	if err != nil {
		fmt.Println(err)
	}
	str, err := eredisClient.GetString("hello")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(str)
	return nil
}
