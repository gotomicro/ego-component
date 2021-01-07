package emongo

import (
	"encoding/json"
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

func debugInterceptor(compName string, c *Config) func(processFn) processFn {
	return func(oldProcess processFn) processFn {
		return func(cmd *cmd) error {
			beg := time.Now()
			err := oldProcess(cmd)
			cost := time.Since(beg)
			if eapp.IsDevelopmentMode() {
				if err != nil {
					log.Println("[emongo.response]",
						xdebug.MakeReqResError(compName,
							fmt.Sprintf("%v", c.DSN), cost, fmt.Sprintf("%s %v", cmd.name, mustJsonMarshal(cmd.req)), err.Error()),
					)
				} else {
					log.Println("[emongo.response]",
						xdebug.MakeReqResInfo(compName,
							fmt.Sprintf("%v", c.DSN), cost, fmt.Sprintf("%s %v", cmd.name, mustJsonMarshal(cmd.req)), fmt.Sprintf("%v", cmd.res)),
					)
				}
			} else {
				// todo log debug info
			}
			return nil
		}
	}
}

func mustJsonMarshal(val interface{}) string {
	res, _ := json.Marshal(val)
	return string(res)
}
