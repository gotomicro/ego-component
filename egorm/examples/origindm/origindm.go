package main

import (
	"database/sql"
	"fmt"

	_ "github.com/gotomicro/dmgo"
)

func main() {
	obj, err := sql.Open("dm", "dm://sysdba:shimo2021@192.168.242.136:5236")
	if err != nil {
		panic(err)
		return
	}
	rows, err := obj.Query("select  TABLE_NAME,comments TABLE_COMMENT from user_tab_comments")
	for rows.Next() {
		a := TableStruct{}
		rows.Scan(&a.TableName, &a.TableComment)
		fmt.Printf("a--------------->"+"%+v\n", a)
	}
}

type TableStruct struct {
	TableName    string //表名
	TableComment string //表注释
}
