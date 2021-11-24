package main

import (
	"fmt"

	"github.com/gotomicro/ego"
	"github.com/gotomicro/ego-component/egorm"
)

func main() {
	err := ego.New().Invoker(
		openDB,
	).Run()
	fmt.Printf("err--------------->"+"%+v\n", err)
}

func openDB() error {
	db := egorm.Load("dm.test").Build()
	fmt.Printf("db--------------->"+"%+v\n", db)
	//1.查询多少张表
	tables := make([]UserTabComments, 0)
	db.Find(&tables)
	// select  TABLE_NAME,comments TABLE_COMMENT from user_tab_comments
	fmt.Printf("tables--------------->"+"%+v\n", tables)
	for _, value := range tables {
		fmt.Printf("value--------------->"+"%+v\n", value)
	}
	return nil
}

type UserTabComments struct {
	TableName    string `gorm:"column:TABLE_NAME"`    //表名
	TableComment string `gorm:"column:TABLE_COMMENT"` //表注释
}
