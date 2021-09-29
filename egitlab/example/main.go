package main

import (
	"fmt"

	"github.com/gotomicro/ego"
	"github.com/gotomicro/ego/core/elog"
	"github.com/xanzy/go-gitlab"
	"go.uber.org/zap"

	"github.com/gotomicro/ego-component/egitlab"
)

func main() {
	err := ego.New().Invoker(invokerGitlab).Run()
	if err != nil {
		elog.Error("startup", elog.Any("err", err))
	}
}

func invokerGitlab() error {
	comp := egitlab.Load("gitlab").Build()
	client := comp.Client()
	user, _, err := client.Users.GetUser(11, gitlab.GetUsersOptions{})
	if err != nil {
		elog.Error("get user failed", zap.Error(err))
		return err
	}
	fmt.Printf("user:%v \n", user)
	return nil
}
