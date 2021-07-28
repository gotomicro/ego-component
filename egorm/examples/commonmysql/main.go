package main

import (
	"github.com/gotomicro/ego"
	"github.com/gotomicro/ego/core/elog"

	"github.com/gotomicro/ego-component/egorm"
)

// 1.新建一个数据库叫test
// 2.执行以下example，export EGO_DEBUG=true && go run main.go --config=config.toml
type User struct {
	Id       int    `gorm:"not null" json:"id"`
	Nickname string `gorm:"not null" json:"name"`
}

func (User) TableName() string {
	return "user2"
}

func main() {
	err := ego.New().Invoker(
		openDB,
		testDB,
	).Run()
	if err != nil {
		elog.Error("startup", elog.Any("err", err))
	}
}

var DBs []*egorm.Component

func openDB() error {
	DBs = []*egorm.Component{
		egorm.Load("mysql.test").Build(),
		//egorm.Load("pg.test").Build(),
		//egorm.Load("other.test").Build(egorm.WithDSNParser(dsn.DefaultPostgresDSNParser)),
	}
	models := []interface{}{
		&User{},
	}
	for _, db := range DBs {
		db.Config.NamingStrategy = &egorm.NamingStrategy{
			SingularTable: true,
		}
		db.AutoMigrate(models...)
		db.Create(&User{
			Nickname: "ego",
		})
	}

	return nil
}

func testDB() error {
	var user User
	for _, db := range DBs {
		err := db.Where("id = ?", 100).First(&user).Error
		elog.Info("user info", elog.String("name", user.Nickname), elog.FieldErr(err))
	}
	return nil
}
