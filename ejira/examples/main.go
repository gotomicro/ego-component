package main

import (
	"fmt"

	"github.com/gotomicro/ego"
	"github.com/gotomicro/ego-component/ejira"
	"github.com/gotomicro/ego/core/elog"
)

// export EGO_DEBUG=true && go run main.go --config=config.toml
func main() {
	err := ego.New().Invoker(
		invokerJira,
	).Run()
	if err != nil {
		elog.Error("startup", elog.Any("err", err))
	}
}

func invokerJira() error {
	comp := ejira.Load("jira").Build()
	userInfo, err := comp.GetUserInfoByUsername("admin")
	fmt.Println(userInfo, err)

	userList, err := comp.FindUsers(&ejira.UserSearchOption{})
	fmt.Println(userList, err)
	return nil
}
