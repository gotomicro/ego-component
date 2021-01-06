package emongo

import (
	"fmt"
	"log"
	"time"

	"github.com/gotomicro/ego/core/eapp"
	"github.com/gotomicro/ego/core/util/xdebug"
)

type Interceptor func(oldProcessFn processFn) (newProcessFn processFn)

func InterceptorChain(interceptors ...Interceptor) func(oldProcess processFn) processFn {
	build := func(interceptor Interceptor, oldProcess processFn) processFn {
		return interceptor(oldProcess)
	}

	return func(oldProcess processFn) processFn {
		chain := oldProcess
		for i := len(interceptors) - 1; i >= 0; i-- {
			chain = build(interceptors[i], chain)
		}
		return chain
	}
}

func debugInterceptor(c *Config) func(processFn) processFn {
	return func(oldProcess processFn) processFn {
		return func() error {
			beg := time.Now()
			err := oldProcess()
			cost := time.Since(beg)
			if eapp.IsDevelopmentMode() {
				if err != nil {
					log.Println("[emongo.response]",
						xdebug.MakeReqResError(PackageName,
							fmt.Sprintf("%v", c.DSN), cost, "", err.Error()),
					)
				} else {
					log.Println("[emongo.response]",
						xdebug.MakeReqResInfo(PackageName,
							fmt.Sprintf("%v", c.DSN), cost, "", ""),
					)
				}
			} else {
				// todo log debug info
			}
			return nil
		}
	}
}
