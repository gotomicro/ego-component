package eredis

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/go-redis/redis"
	"github.com/gotomicro/ego/core/eapp"
	"github.com/gotomicro/ego/core/elog"
	"github.com/gotomicro/ego/core/emetric"
	"github.com/gotomicro/ego/core/util/xdebug"
)

type CmdHandler func(cmd redis.Cmder) error

type Interceptor func(oldProcess CmdHandler) CmdHandler

func InterceptorChain(interceptors ...Interceptor) func(oldProcess func(cmd redis.Cmder) error) func(cmd redis.Cmder) error {
	build := func(interceptor Interceptor, oldProcess CmdHandler) CmdHandler {
		return interceptor(oldProcess)
	}

	return func(oldProcess func(cmd redis.Cmder) error) func(cmd redis.Cmder) error {
		chain := oldProcess
		for i := len(interceptors) - 1; i >= 0; i-- {
			chain = build(interceptors[i], chain)
		}
		return chain
	}
}

func debugInterceptor(compName string, config *Config, logger *elog.Component) Interceptor {
	return func(oldProcess CmdHandler) CmdHandler {
		return func(cmd redis.Cmder) error {
			beg := time.Now()
			err := oldProcess(cmd)
			cost := time.Since(beg)
			if eapp.IsDevelopmentMode() {
				if err != nil {
					log.Println("[eredis.response]",
						xdebug.MakeReqResError(compName, fmt.Sprintf("%v", config.Addrs), cost, fmt.Sprintf("%v", cmd.Args()), err.Error()),
					)
				} else {
					log.Println("[eredis.response]",
						xdebug.MakeReqResInfo(compName, fmt.Sprintf("%v", config.Addrs), cost, fmt.Sprintf("%v", cmd.Args()), response(cmd)),
					)
				}
			} else {
				// todo log debug info
			}
			return err
		}
	}
}

func metricInterceptor(compName string, config *Config, logger *elog.Component) Interceptor {
	return func(oldProcess CmdHandler) CmdHandler {
		return func(cmd redis.Cmder) error {
			beg := time.Now()
			err := oldProcess(cmd)
			cost := time.Since(beg)
			// error metric
			if err != nil {
				if errors.Is(err, redis.Nil) {
					emetric.ClientHandleCounter.Inc(emetric.TypeRedis, compName, cmd.Name(), strings.Join(config.Addrs, ","), "Empty")
				} else {
					emetric.ClientHandleCounter.Inc(emetric.TypeRedis, compName, cmd.Name(), strings.Join(config.Addrs, ","), "Error")
				}
			} else {
				emetric.ClientHandleCounter.Inc(emetric.TypeRedis, compName, cmd.Name(), strings.Join(config.Addrs, ","), "OK")
			}
			emetric.ClientHandleHistogram.WithLabelValues(emetric.TypeRedis, compName, cmd.Name(), strings.Join(config.Addrs, ",")).Observe(cost.Seconds())
			return err
		}
	}
}

func accessInterceptor(compName string, config *Config, logger *elog.Component) Interceptor {
	return func(oldProcess CmdHandler) CmdHandler {
		return func(cmd redis.Cmder) error {
			beg := time.Now()
			err := oldProcess(cmd)
			cost := time.Since(beg)

			var fields = make([]elog.Field, 0, 15)
			fields = append(fields, elog.FieldComponentName(compName), elog.FieldMethod(cmd.Name()), elog.FieldCost(cost))

			if config.EnableAccessInterceptorReq {
				fields = append(fields, elog.Any("req", cmd.Args()))
			}
			if config.EnableAccessInterceptorRes && cmd.Err() == nil {
				fields = append(fields, elog.Any("res", response(cmd)))
			}
			isErrLog := false
			isSlowLog := false
			// error metric
			if err != nil {
				fields = append(fields, elog.FieldEvent("error"), elog.FieldErr(err))
				isErrLog = true
				if errors.Is(err, redis.Nil) {
					logger.Warn("access", fields...)
				} else {
					logger.Error("access", fields...)
				}
			}

			if config.SlowLogThreshold > time.Duration(0) && cost > config.SlowLogThreshold {
				fields = append(fields, elog.FieldEvent("slow"))
				logger.Info("access", fields...)
				isSlowLog = true
			}

			if config.EnableAccessInterceptor && !isSlowLog && !isErrLog {
				fields = append(fields, elog.FieldEvent("normal"))
				logger.Info("access", fields...)
			}
			return err
		}
	}
}

func response(cmd redis.Cmder) string {
	switch cmd.(type) {
	case *redis.Cmd:
		return fmt.Sprintf("%v", cmd.(*redis.Cmd).Val())
	case *redis.StringCmd:
		return fmt.Sprintf("%v", cmd.(*redis.StringCmd).Val())
	case *redis.StatusCmd:
		return fmt.Sprintf("%v", cmd.(*redis.StatusCmd).Val())
	case *redis.IntCmd:
		return fmt.Sprintf("%v", cmd.(*redis.IntCmd).Val())
	case *redis.DurationCmd:
		return fmt.Sprintf("%v", cmd.(*redis.DurationCmd).Val())
	case *redis.BoolCmd:
		return fmt.Sprintf("%v", cmd.(*redis.BoolCmd).Val())
	case *redis.CommandsInfoCmd:
		return fmt.Sprintf("%v", cmd.(*redis.CommandsInfoCmd).Val())
	case *redis.StringSliceCmd:
		return fmt.Sprintf("%v", cmd.(*redis.StringSliceCmd).Val())
	default:
		return ""
	}
}
