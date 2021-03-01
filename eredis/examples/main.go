package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-redis/redis/v8"
	"github.com/gotomicro/ego"
	"github.com/gotomicro/ego/core/elog"

	"github.com/gotomicro/ego-component/eredis"
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
	eredisClient = eredis.Load("redis.test").Build()
	return nil
}

func testRedis() error {
	err := eredisClient.Set(context.Background(), "hello", "world", 0)
	fmt.Println("set hello", err)

	str, err := eredisClient.Get(context.Background(), "hello")
	fmt.Println("get hello", str, err)

	str, err = eredisClient.Get(context.Background(), "lee")
	fmt.Println("Get lee", err, errors.Is(err, redis.Nil))

	return nil
}
