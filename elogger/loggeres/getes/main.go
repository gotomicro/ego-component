package main

import (
	"github.com/gotomicro/ego"
	"github.com/gotomicro/ego-component/elogger/loggeres"
	"github.com/gotomicro/ego/core/elog"
)

func init() {
	elog.Register(&loggeres.EsWriterBuilder{})
}

//  export EGO_DEBUG=true && go run main.go --config=config.toml
func main() {
	err := ego.New().Invoker(func() error {
		elog.EgoLogger.Info("hello world2222")
		return nil
	}).Run()
	if err != nil {
		panic(err)
	}

}
