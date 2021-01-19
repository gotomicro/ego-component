# ekafka 组件使用指南


## 简介 

对 [kafka-go](https://github.com/segmentio/kafka-go) 进行了轻量封装，并提供了以下功能：

- 规范了标准配置格式，提供了统一的 Load().Build() 方法。
- 支持自定义拦截器
- 提供了默认的 Debug 拦截器，开启 Debug 后可输出 Request、Response 至终端。

## 快速上手

生产者消费者使用样例可参考 [example](examples/main.go)

