package main

import (
	"fmt"
	"github.com/gotomicro/ego"
	"github.com/gotomicro/ego-component/edingtalk"
	"github.com/gotomicro/ego-component/eredis"
	"github.com/gotomicro/ego/core/elog"
)

// export EGO_DEBUG=true && go run main.go --config=config.toml
func main() {
	err := ego.New().Invoker(
		invokerDingTalk,
	).Run()
	if err != nil {
		elog.Error("startup", elog.Any("err", err))
	}
}

func invokerDingTalk() error {
	redis := eredis.Load("redis").Build(eredis.WithStub())
	comp := edingtalk.Load("dingtalk").Build(edingtalk.WithERedis(redis))
	user, err := comp.GetUserInfo("5eca6dcdc3c93cf8b9eb91a922e90d97")
	fmt.Println(user)
	fmt.Println(err)
	return nil
}
