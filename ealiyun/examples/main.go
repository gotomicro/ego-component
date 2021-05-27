package main

import (
	"fmt"
	"github.com/gotomicro/ego"
	"github.com/gotomicro/ego-component/ealiyun"
	"github.com/gotomicro/ego/core/elog"
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
	comp := ealiyun.Load("aliyun").Build()
	userName := "lisi@xxxx.onaliyun.com"
	res, err := comp.CreateRamUser(ealiyun.SaveRamUserRequest{
		UserPrincipalName: userName,
		DisplayName:       "李四",
		MobilePhone:       "xxxxxxx",
		Email:             "xxxxx",
	})
	if err != nil {
		fmt.Println("createUser err:" + err.Error())
		return err
	}
	fmt.Printf("createUser res:%#v\n", res)
	fmt.Println("=============================================")
	res, err = comp.GetRamUser(userName)
	if err != nil {
		fmt.Println("createUser err:" + err.Error())
		return err
	}
	fmt.Printf("getUser res:%#v\n", res)
	return nil
}
