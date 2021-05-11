# egorm 组件使用指南


## 简介

对 [gorm](https://github.com/jinzhu/gorm) 进行了轻量封装，并提供了以下功能：

- 规范了标准配置格式，提供了统一的 Load().Build() 方法。
- 支持自定义拦截器
- 提供了默认的 Debug 拦截器，开启 Debug 后可输出 Request、Response 至终端。
- 提供了默认的 Metric 拦截器，开启后可采集 Prometheus 指标数据
- 提供了默认的 OpenTracing 拦截器，开启后可采集 Tracing Span 数据

## 快速上手

数据库使用样例可参考 [example](examples/main.go)

