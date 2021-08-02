package ekafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/gotomicro/ego/core/eapp"
	"github.com/gotomicro/ego/core/elog"
	"github.com/gotomicro/ego/core/emetric"
	"github.com/gotomicro/ego/core/etrace"
	"github.com/gotomicro/ego/core/util/xdebug"
	"github.com/gotomicro/ego/core/util/xstring"
	"github.com/opentracing/opentracing-go"
	"github.com/segmentio/kafka-go"
)

type serverProcessFn func(context.Context, Messages, *cmd) error

type ServerInterceptor func(oldProcessFn serverProcessFn) (newProcessFn serverProcessFn)

func InterceptorServerChain(interceptors ...ServerInterceptor) ServerInterceptor {
	return func(p serverProcessFn) serverProcessFn {
		chain := p
		for i := len(interceptors) - 1; i >= 0; i-- {
			chain = buildServerInterceptor(interceptors[i], chain)
		}
		return chain
	}
}

func buildServerInterceptor(interceptor ServerInterceptor, oldProcess serverProcessFn) serverProcessFn {
	return interceptor(oldProcess)
}

func fixedServerInterceptor(_ string, _ *config) ServerInterceptor {
	return func(next serverProcessFn) serverProcessFn {
		return func(ctx context.Context, msgs Messages, cmd *cmd) error {
			start := time.Now()
			ctx = context.WithValue(ctx, ctxStartTimeKey{}, start)
			err := next(ctx, msgs, cmd)
			return err
		}
	}
}

func traceServerInterceptor(compName string, c *config) ServerInterceptor {
	return func(next serverProcessFn) serverProcessFn {
		return func(ctx context.Context, msgs Messages, cmd *cmd) error {
			_, ctx = etrace.StartSpanFromContext(
				ctx,
				"kafka",
			)
			md := etrace.MetadataReaderWriter{MD: map[string][]string{}}
			span := opentracing.SpanFromContext(ctx)
			_ = opentracing.GlobalTracer().Inject(span.Context(), opentracing.HTTPHeaders, md)
			headers := make([]kafka.Header, 0)
			md.ForeachKey(func(key, val string) error {
				headers = append(headers, kafka.Header{
					Key:   key,
					Value: []byte(val),
				})
				return nil
			})
			for _, value := range msgs {
				value.Headers = headers
				value.Time = time.Now()
			}
			err := next(ctx, msgs, cmd)
			return err
		}
	}
}

func accessServerInterceptor(compName string, c *config) ServerInterceptor {
	return func(next serverProcessFn) serverProcessFn {
		return func(ctx context.Context, msgs Messages, cmd *cmd) error {
			err := next(ctx, msgs, cmd)
			// kafka 比较坑爹，合在一起处理链路
			if c.EnableTraceInterceptor {
				mds := make(map[string][]string)
				for _, value := range cmd.msg.Headers {
					mds[value.Key] = []string{string(value.Value)}
				}
				md := etrace.MetadataReaderWriter{MD: mds}
				sc, _ := opentracing.GlobalTracer().Extract(opentracing.HTTPHeaders, md)

				// 重新赋值ctx
				_, ctx = etrace.StartSpanFromContext(
					ctx,
					"kafka",
					opentracing.ChildOf(sc),
				)
			}

			cost := time.Since(ctx.Value(ctxStartTimeKey{}).(time.Time))
			if c.EnableAccessInterceptor {
				var fields = make([]elog.Field, 0, 10)

				fields = append(fields,
					elog.FieldMethod(cmd.name),
					elog.FieldCost(cost),
				)

				// 开启了链路，那么就记录链路id
				if c.EnableTraceInterceptor && opentracing.IsGlobalTracerRegistered() {
					fields = append(fields, elog.FieldTid(etrace.ExtractTraceID(ctx)))
				}
				if c.EnableAccessInterceptorReq {
					fields = append(fields, elog.Any("req", json.RawMessage(xstring.JSON(msgs.ToLog()))))
				}
				if c.EnableAccessInterceptorRes {
					fields = append(fields, elog.Any("res", json.RawMessage(xstring.JSON(messageToLog(cmd.msg)))))
				}
				elog.Info("access", fields...)
			}

			if !eapp.IsDevelopmentMode() {
				return err
			}
			if err != nil {
				log.Println("[ekafka.response]", xdebug.MakeReqResError(compName,
					fmt.Sprintf("%v", c.Brokers), cost, fmt.Sprintf("%s %v", cmd.name, xstring.JSON(msgs.ToLog())), err.Error()),
				)
			} else {
				log.Println("[ekafka.response]", xdebug.MakeReqResInfo(compName,
					fmt.Sprintf("%v", c.Brokers), cost, fmt.Sprintf("%s %v", cmd.name, xstring.JSON(msgs)), xstring.JSON(messageToLog(cmd.msg))),
				)
			}
			return err
		}
	}
}

func metricServerInterceptor(compName string, config *config) ServerInterceptor {
	return func(next serverProcessFn) serverProcessFn {
		return func(ctx context.Context, msgs Messages, cmd *cmd) error {
			err := next(ctx, msgs, cmd)
			cost := time.Since(ctx.Value(ctxStartTimeKey{}).(time.Time))
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
