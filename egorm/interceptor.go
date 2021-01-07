package egorm

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/gotomicro/ego/core/eapp"
	"github.com/gotomicro/ego/core/elog"
	"github.com/gotomicro/ego/core/emetric"
	"github.com/gotomicro/ego/core/etrace"
	"github.com/gotomicro/ego/core/util/xdebug"

)

// Handler ...
type Handler func(*Scope)

// Interceptor ...
type Interceptor func(string, *DSN, string, *Config, *elog.Component) func(next Handler) Handler

func debugInterceptor(compName string, dsn *DSN, op string, options *Config, logger *elog.Component) func(Handler) Handler {
	return func(next Handler) Handler {
		return func(scope *Scope) {
			beg := time.Now()
			next(scope)
			cost := time.Since(beg)
			if eapp.IsDevelopmentMode() {
				if scope.HasError() {
					log.Println("[egorm.response]",
						xdebug.MakeReqResError(compName, fmt.Sprintf("%v", dsn.Addr+"/"+dsn.DBName), cost, logSQL(scope.SQL, scope.SQLVars, true), scope.DB().Error.Error()),
					)
				} else {
					log.Println("[egorm.response]",
						xdebug.MakeReqResInfo(compName, fmt.Sprintf("%v", dsn.Addr+"/"+dsn.DBName), cost, logSQL(scope.SQL, scope.SQLVars, true), fmt.Sprintf("%v", scope.Value)),
					)
				}
			} else {
				// todo log debug info
			}
		}
	}
}

func metricInterceptor(compName string, dsn *DSN, op string, config *Config, logger *elog.Component) func(Handler) Handler {
	return func(next Handler) Handler {
		return func(scope *Scope) {
			beg := time.Now()
			next(scope)
			cost := time.Since(beg)
			var fields = make([]elog.Field, 0, 15)
			fields = append(fields,
				elog.FieldMethod(op),
				elog.FieldName(dsn.DBName+"."+scope.TableName()),
				elog.FieldCost(cost),
			)
			if config.EnableAccessInterceptorReq {
				fields = append(fields, elog.String("req", logSQL(scope.SQL, scope.SQLVars, config.EnableDetailSQL)))
			}

			if config.EnableAccessInterceptorRes {
				fields = append(fields, elog.Any("res", scope.Value))
			}

			isErrLog := false
			isSlowLog := false
			// error metric
			if scope.HasError() {
				fields = append(fields,
					elog.FieldEvent("error"),
					elog.FieldErr(scope.DB().Error),
				)
				if errors.Is(scope.DB().Error, ErrRecordNotFound) {
					logger.Warn("access", fields...)
					emetric.ClientHandleCounter.Inc(emetric.TypeGorm, compName, dsn.DBName+"."+scope.TableName(), dsn.Addr, "Empty")
				} else {
					logger.Error("access", fields...)
					emetric.ClientHandleCounter.Inc(emetric.TypeGorm, compName, dsn.DBName+"."+scope.TableName(), dsn.Addr, "Error")
				}
				isErrLog = true
			} else {
				emetric.ClientHandleCounter.Inc(emetric.TypeGorm, compName, dsn.DBName+"."+scope.TableName(), dsn.Addr, "OK")
			}

			emetric.ClientHandleHistogram.WithLabelValues(emetric.TypeGorm, compName, dsn.DBName+"."+scope.TableName(), dsn.Addr).Observe(cost.Seconds())

			if config.SlowLogThreshold > time.Duration(0) && config.SlowLogThreshold < cost {
				fields = append(fields,
					elog.FieldEvent("slow"),
				)
				logger.Warn("access", fields...)
			}

			if config.EnableAccessInterceptor && !isSlowLog && !isErrLog {
				fields = append(fields,
					elog.FieldEvent("normal"),
				)
				logger.Info("access", fields...)
			}

		}
	}
}

func logSQL(sql string, args []interface{}, containArgs bool) string {
	if containArgs {
		return bindSQL(sql, args)
	}
	return sql
}

func traceInterceptor(compName string, dsn *DSN, op string, options *Config, logger *elog.Component) func(Handler) Handler {
	return func(next Handler) Handler {
		return func(scope *Scope) {
			if val, ok := scope.Get("_context"); ok {
				if ctx, ok := val.(context.Context); ok {
					span, _ := etrace.StartSpanFromContext(
						ctx,
						"GORM", // TODO this op value is op or GORM
						etrace.TagComponent("mysql"),
						etrace.TagSpanKind("client"),
					)
					defer span.Finish()

					// 延迟执行 scope.CombinedConditionSql() 避免sqlVar被重复追加
					next(scope)

					span.SetTag("sql.inner", dsn.DBName)
					span.SetTag("sql.addr", dsn.Addr)
					span.SetTag("span.kind", "client")
					span.SetTag("peer.service", "mysql")
					span.SetTag("db.instance", dsn.DBName)
					span.SetTag("peer.address", dsn.Addr)
					span.SetTag("peer.statement", logSQL(scope.SQL, scope.SQLVars, options.EnableDetailSQL))
					return
				}
			}

			next(scope)
		}
	}
}
