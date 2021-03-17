package eredis

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/gotomicro/ego/core/elog"
	"github.com/gotomicro/ego/core/emetric"
	"github.com/gotomicro/ego/core/util/xdebug"
)

const ctxBegKey = "_cmdResBegin_"

type interceptor struct {
	beforeProcess         func(ctx context.Context, cmd redis.Cmder) (context.Context, error)
	afterProcess          func(ctx context.Context, cmd redis.Cmder) error
	beforeProcessPipeline func(ctx context.Context, cmds []redis.Cmder) (context.Context, error)
	afterProcessPipeline  func(ctx context.Context, cmds []redis.Cmder) error
}

func (i *interceptor) BeforeProcess(ctx context.Context, cmd redis.Cmder) (context.Context, error) {
	return i.beforeProcess(ctx, cmd)
}

func (i *interceptor) AfterProcess(ctx context.Context, cmd redis.Cmder) error {
	return i.afterProcess(ctx, cmd)
}

func (i *interceptor) BeforeProcessPipeline(ctx context.Context, cmds []redis.Cmder) (context.Context, error) {
	return i.beforeProcessPipeline(ctx, cmds)
}

func (i *interceptor) AfterProcessPipeline(ctx context.Context, cmds []redis.Cmder) error {
	return i.afterProcessPipeline(ctx, cmds)
}

func newInterceptor(compName string, config *config, logger *elog.Component) *interceptor {
	return &interceptor{
		beforeProcess: func(ctx context.Context, cmd redis.Cmder) (context.Context, error) {
			return ctx, nil
		},
		afterProcess: func(ctx context.Context, cmd redis.Cmder) error {
			return nil
		},
		beforeProcessPipeline: func(ctx context.Context, cmds []redis.Cmder) (context.Context, error) {
			return ctx, nil
		},
		afterProcessPipeline: func(ctx context.Context, cmds []redis.Cmder) error {
			return nil
		},
	}
}

func (i *interceptor) setBeforeProcess(p func(ctx context.Context, cmd redis.Cmder) (context.Context, error)) *interceptor {
	i.beforeProcess = p
	return i
}

func (i *interceptor) setAfterProcess(p func(ctx context.Context, cmd redis.Cmder) error) *interceptor {
	i.afterProcess = p
	return i
}

func (i *interceptor) setBeforeProcessPipeline(p func(ctx context.Context, cmds []redis.Cmder) (context.Context, error)) *interceptor {
	i.beforeProcessPipeline = p
	return i
}

func (i *interceptor) setAfterProcessPipeline(p func(ctx context.Context, cmds []redis.Cmder) error) *interceptor {
	i.afterProcessPipeline = p
	return i
}

func fixedInterceptor(compName string, config *config, logger *elog.Component) *interceptor {
	return newInterceptor(compName, config, logger).
		setBeforeProcess(func(ctx context.Context, cmd redis.Cmder) (context.Context, error) {
			return context.WithValue(ctx, ctxBegKey, time.Now()), nil
		}).
		setAfterProcess(func(ctx context.Context, cmd redis.Cmder) error {
			var err = cmd.Err()
			// go-redis script的error做了prefix处理
			// https://github.com/go-redis/redis/blob/master/script.go#L61
			if err != nil && !strings.HasPrefix(err.Error(), "NOSCRIPT ") {
				err = fmt.Errorf("eredis exec command %s fail, %w", cmd.Name(), err)
			}
			return err
		})
}

func debugInterceptor(compName string, config *config, logger *elog.Component) *interceptor {
	return newInterceptor(compName, config, logger).setAfterProcess(
		func(ctx context.Context, cmd redis.Cmder) error {
			cost := time.Since(ctx.Value(ctxBegKey).(time.Time))
			err := cmd.Err()
			if err != nil {
				log.Println("[eredis.response]",
					xdebug.MakeReqResError(compName, fmt.Sprintf("%v", config.Addrs), cost, fmt.Sprintf("%v", cmd.Args()), err.Error()),
				)
			} else {
				log.Println("[eredis.response]",
					xdebug.MakeReqResInfo(compName, fmt.Sprintf("%v", config.Addrs), cost, fmt.Sprintf("%v", cmd.Args()), response(cmd)),
				)
			}
			return err
		},
	)
}

func metricInterceptor(compName string, config *config, logger *elog.Component) *interceptor {
	return newInterceptor(compName, config, logger).setAfterProcess(
		func(ctx context.Context, cmd redis.Cmder) error {
			cost := time.Since(ctx.Value(ctxBegKey).(time.Time))
			err := cmd.Err()
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
		},
	)
}

func accessInterceptor(compName string, config *config, logger *elog.Component) *interceptor {
	return newInterceptor(compName, config, logger).setAfterProcess(
		func(ctx context.Context, cmd redis.Cmder) error {
			var fields = make([]elog.Field, 0, 15)
			var err = cmd.Err()
			cost := time.Since(ctx.Value(ctxBegKey).(time.Time))
			fields = append(fields, elog.FieldComponentName(compName), elog.FieldMethod(cmd.Name()), elog.FieldCost(cost))

			if config.EnableAccessInterceptorReq {
				fields = append(fields, elog.Any("req", cmd.Args()))
			}
			if config.EnableAccessInterceptorRes && err == nil {
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
		},
	)
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
