package emongo

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/gotomicro/ego/core/eapp"
	"github.com/gotomicro/ego/core/elog"
	"github.com/gotomicro/ego/core/emetric"
	"github.com/gotomicro/ego/core/util/xdebug"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	metricType = "mongo"
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

func debugInterceptor(compName string, c *config) func(processFn) processFn {
	return func(oldProcess processFn) processFn {
		return func(cmd *cmd) error {
			beg := time.Now()
			err := oldProcess(cmd)
			cost := time.Since(beg)
			if eapp.IsDevelopmentMode() {
				if err != nil {
					log.Println("[emongo.response]", xdebug.MakeReqResError(compName,
						fmt.Sprintf("%v", c.DSN), cost, fmt.Sprintf("%s %v", cmd.name, mustJsonMarshal(cmd.req)), err.Error()),
					)
				} else {
					log.Println("[emongo.response]", xdebug.MakeReqResInfo(compName,
						fmt.Sprintf("%v", c.DSN), cost, fmt.Sprintf("%s %v", cmd.name, mustJsonMarshal(cmd.req)), fmt.Sprintf("%v", cmd.res)),
					)
				}
			} else {
				// todo log debug info
			}
			return err
		}
	}
}

func metricInterceptor(compName string, c *config, logger *elog.Component) func(processFn) processFn {
	return func(oldProcess processFn) processFn {
		return func(cmd *cmd) error {
			beg := time.Now()
			err := oldProcess(cmd)
			cost := time.Since(beg)
			if err != nil {
				if errors.Is(err, mongo.ErrNoDocuments) {
					emetric.ClientHandleCounter.Inc(metricType, compName, cmd.name, c.DSN, "Empty")
				} else {
					emetric.ClientHandleCounter.Inc(metricType, compName, cmd.name, c.DSN, "Error")
				}
			} else {
				emetric.ClientHandleCounter.Inc(metricType, compName, cmd.name, c.DSN, "OK")
			}
			emetric.ClientHandleHistogram.WithLabelValues(metricType, compName, cmd.name, c.DSN).Observe(cost.Seconds())
			return err
		}
	}
}

func accessInterceptor(compName string, c *config, logger *elog.Component) func(processFn) processFn {
	return func(oldProcess processFn) processFn {
		return func(cmd *cmd) error {
			beg := time.Now()
			err := oldProcess(cmd)
			cost := time.Since(beg)

			var fields = make([]elog.Field, 0, 15)
			fields = append(fields,
				elog.FieldComponentName(compName),
				elog.FieldMethod(cmd.name),
				elog.FieldCost(cost),
			)
			if c.EnableAccessInterceptorReq {
				fields = append(fields, elog.Any("req", cmd.req))
			}
			if c.EnableAccessInterceptorRes && err == nil {
				fields = append(fields, elog.Any("res", cmd.res))
			}

			if c.SlowLogThreshold > time.Duration(0) && cost > c.SlowLogThreshold {
				logger.Warn("slow", fields...)
			}

			if err != nil {
				fields = append(fields, elog.FieldEvent("error"), elog.FieldErr(err))
				if errors.Is(err, mongo.ErrNoDocuments) {
					logger.Warn("access", fields...)
					return err
				}
				logger.Error("access", fields...)
				return err
			}

			if c.EnableAccessInterceptor {
				fields = append(fields, elog.FieldEvent("normal"))
				logger.Info("access", fields...)
			}
			return nil
		}
	}
}

func mustJsonMarshal(val interface{}) string {
	res, _ := json.Marshal(val)
	return string(res)
}
