package main

import (
	"fmt"

	"github.com/gotomicro/ego"
	"github.com/gotomicro/ego/core/elog"

	"github.com/gotomicro/ego-component/ehuawei"
)

func main() {
	err := ego.New().Invoker(
		invoker,
	).Run()
	if err != nil {
		elog.Error("startup", elog.Any("err", err))
	}
}
func invoker() error {
	comp := ehuawei.Load("huawei").Build()
	res, err := comp.KeystoneListGroups("xxxxx")
	if err != nil {
		panic("KeystoneListGroups:" + err.Error())
	}
	fmt.Println(res)
	return nil
}
