package main

import (
	"github.com/gotomicro/ego"
	"github.com/gotomicro/ego-component/eredis/ecronlock"
	"github.com/gotomicro/ego/core/elog"
	"github.com/gotomicro/ego/task/ecron"
	"time"

	"github.com/gotomicro/ego-component/eredis"
)

// export EGO_DEBUG=true && go run main.go --config=config.toml
func main() {
	err := ego.New().Cron(cronJob()).Run()
	if err != nil {
		elog.Panic("startup", elog.FieldErr(err))
	}
}

var eredisClient *eredis.Component

func cronJob() ecron.Ecron {
	lock := ecronlock.Load("").Build(ecronlock.WithClientRedis(eredis.Load("redis.test").Build()))
	cron := ecron.Load("cron.default").Build(ecron.WithLocker(lock))
	cron.Schedule(ecron.Every(time.Second*10), ecron.FuncJob(helloWorld))
	return cron
}

func helloWorld() error {
	elog.Info("info job")
	elog.Warn("warn job")
	return nil
}
