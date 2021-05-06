package ekafka

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gotomicro/ego/core/emetric"
	"log"
	"strings"
	"time"

	"github.com/gotomicro/ego/core/eapp"
	"github.com/gotomicro/ego/core/util/xdebug"
)

const (
	ctxStartTimeKey = "_cmdStart_"
)

type processor func(fn processFn) error
type processFn func(*cmd) error

type cmd struct {
	ctx  context.Context
	name string
	req  interface{}
	res  interface{}
}

type Interceptor func(oldProcessFn processFn) (newProcessFn processFn)

func InterceptorChain(interceptors ...Interceptor) Interceptor {
	build := func(interceptor Interceptor, oldProcess processFn) processFn {
		return interceptor(oldProcess)
	}

	return func(p processFn) processFn {
		chain := p
		for i := len(interceptors) - 1; i >= 0; i-- {
			chain = build(interceptors[i], chain)
		}
		return chain
	}
}

func fixedInterceptor(_ string, _ *config) Interceptor {
	return func(next processFn) processFn {
		return func(cmd *cmd) error {
			start := time.Now()
			err := next(cmd)
			cmd.ctx = context.WithValue(cmd.ctx, ctxStartTimeKey, start)
			return err
		}
	}
}

func debugInterceptor(compName string, c *config) Interceptor {
	return func(next processFn) processFn {
		return func(cmd *cmd) error {
			err := next(cmd)
			cost := time.Since(cmd.ctx.Value(ctxStartTimeKey).(time.Time))
			if eapp.IsDevelopmentMode() {
				if err != nil {
					log.Println("[ekafka.response]", xdebug.MakeReqResError(compName,
						fmt.Sprintf("%v", c.Brokers), cost, fmt.Sprintf("%s %v", cmd.name, mustJsonMarshal(cmd.req)), err.Error()),
					)
				} else {
					log.Println("[ekafka.response]", xdebug.MakeReqResInfo(compName,
						fmt.Sprintf("%v", c.Brokers), cost, fmt.Sprintf("%s %v", cmd.name, mustJsonMarshal(cmd.req)), fmt.Sprintf("%v", cmd.res)),
					)
				}
			} else {
				// todo log debug info
			}
			return err
		}
	}
}

func metricInterceptor(compName string, config *config) func(processFn) processFn {
	return func(next processFn) processFn {
		return func(cmd *cmd) error {
			err := next(cmd)
			cost := time.Since(cmd.ctx.Value(ctxStartTimeKey).(time.Time))
			emetric.ClientHandleHistogram.WithLabelValues("kafka", compName, cmd.name, strings.Join(config.Brokers, ",")).Observe(cost.Seconds())
			if err != nil {
				emetric.ClientHandleCounter.Inc("kafka", compName, cmd.name, strings.Join(config.Brokers, ","), "Error")
				return err
			}
			emetric.ClientHandleCounter.Inc("kafka", compName, cmd.name, strings.Join(config.Brokers, ","), "OK")
			return nil
		}
	}
}

func mustJsonMarshal(val interface{}) string {
	res, _ := json.Marshal(val)
	return string(res)
}
