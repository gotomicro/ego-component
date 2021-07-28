package egorm

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/gotomicro/ego-component/egorm/dsn"
	"github.com/spf13/cast"

	"github.com/gotomicro/ego/core/eapp"
	"github.com/gotomicro/ego/core/elog"
	"github.com/gotomicro/ego/core/emetric"
	"github.com/gotomicro/ego/core/etrace"
	"github.com/gotomicro/ego/core/util/xdebug"
	"github.com/opentracing/opentracing-go"
	"gorm.io/gorm"
)

// Handler ...
type Handler func(*gorm.DB)

// Processor ...
type Processor interface {
	Get(name string) func(*gorm.DB)
	Replace(name string, handler func(*gorm.DB)) error
}

// Interceptor ...
type Interceptor func(string, *dsn.DSN, string, *config, *elog.Component) func(next Handler) Handler

func debugInterceptor(compName string, dsn *dsn.DSN, op string, options *config, logger *elog.Component) func(Handler) Handler {
	return func(next Handler) Handler {
		return func(db *gorm.DB) {
			if !eapp.IsDevelopmentMode() {
				return
			}
			beg := time.Now()
			next(db)
			cost := time.Since(beg)
			if db.Error != nil {
				log.Println("[egorm.response]",
					xdebug.MakeReqResError(compName, fmt.Sprintf("%v", dsn.Addr+"/"+dsn.DBName), cost, logSQL(db.Statement.SQL.String(), db.Statement.Vars, true), db.Error.Error()),
				)
			} else {
				log.Println("[egorm.response]",
					xdebug.MakeReqResInfo(compName, fmt.Sprintf("%v", dsn.Addr+"/"+dsn.DBName), cost, logSQL(db.Statement.SQL.String(), db.Statement.Vars, true), fmt.Sprintf("%v", db.Statement.Dest)),
				)
			}

		}
	}
}

func metricInterceptor(compName string, dsn *dsn.DSN, op string, config *config, logger *elog.Component) func(Handler) Handler {
	return func(next Handler) Handler {
		return func(db *gorm.DB) {
			beg := time.Now()
			next(db)
			cost := time.Since(beg)
			var fields = make([]elog.Field, 0, 15)
			fields = append(fields, elog.FieldMethod(op), elog.FieldName(dsn.DBName+"."+db.Statement.Table), elog.FieldCost(cost))
			if config.EnableAccessInterceptorReq {
				fields = append(fields, elog.String("req", logSQL(db.Statement.SQL.String(), db.Statement.Vars, config.EnableDetailSQL)))
			}
			if config.EnableAccessInterceptorRes {
				fields = append(fields, elog.Any("res", db.Statement.Dest))
			}

			// 开启了链路，那么就记录链路id
			if config.EnableTraceInterceptor && opentracing.IsGlobalTracerRegistered() {
				fields = append(fields, elog.FieldTid(etrace.ExtractTraceID(db.Statement.Context)))
			}

			// 支持自定义log
			for _, key := range eapp.EgoLogExtraKeys() {
				if value := getContextValue(db.Statement.Context, key); value != "" {
					fields = append(fields, elog.FieldCustomKeyValue(key, value))
				}
			}

			// 记录监控耗时
			emetric.ClientHandleHistogram.WithLabelValues(emetric.TypeGorm, compName, dsn.DBName+"."+db.Statement.Table, dsn.Addr).Observe(cost.Seconds())

			// 如果有慢日志，就记录
			if config.SlowLogThreshold > time.Duration(0) && config.SlowLogThreshold < cost {
				logger.Warn("slow", fields...)
			}

			// 如果有错误，记录错误信息
			if db.Error != nil {
				fields = append(fields, elog.FieldEvent("error"), elog.FieldErr(db.Error))
				if errors.Is(db.Error, ErrRecordNotFound) {
					logger.Warn("access", fields...)
					emetric.ClientHandleCounter.Inc(emetric.TypeGorm, compName, dsn.DBName+"."+db.Statement.Table, dsn.Addr, "Empty")
					return
				}
				logger.Error("access", fields...)
				emetric.ClientHandleCounter.Inc(emetric.TypeGorm, compName, dsn.DBName+"."+db.Statement.Table, dsn.Addr, "Error")
				return
			}

			emetric.ClientHandleCounter.Inc(emetric.TypeGorm, compName, dsn.DBName+"."+db.Statement.Table, dsn.Addr, "OK")
			// 开启了记录日志信息，那么就记录access
			// event normal和error，代表全部access的请求数
			if config.EnableAccessInterceptor {
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

func traceInterceptor(compName string, dsn *dsn.DSN, op string, options *config, logger *elog.Component) func(Handler) Handler {
	return func(next Handler) Handler {
		return func(db *gorm.DB) {
			if db.Statement.Context != nil {
				span, _ := etrace.StartSpanFromContext(
					db.Statement.Context,
					"GORM", // TODO this op value is op or GORM
					etrace.TagComponent("mysql"),
					etrace.TagSpanKind("client"),
				)

				defer span.Finish()

				// 延迟执行 scope.CombinedConditionSql() 避免sqlVar被重复追加
				next(db)

				span.SetTag("sql.inner", dsn.DBName)
				span.SetTag("sql.addr", dsn.Addr)
				span.SetTag("span.kind", "client")
				span.SetTag("peer.service", "mysql")
				span.SetTag("db.instance", dsn.DBName)
				span.SetTag("peer.address", dsn.Addr)
				span.SetTag("peer.statement", logSQL(db.Statement.SQL.String(), db.Statement.Vars, options.EnableDetailSQL))
				return
			}

			next(db)
		}
	}
}

func getContextValue(c context.Context, key string) string {
	if key == "" {
		return ""
	}
	return cast.ToString(c.Value(key))
}
