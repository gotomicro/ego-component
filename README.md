# ego-component
<!-- ALL-CONTRIBUTORS-BADGE:START - Do not remove or modify this section -->
[![All Contributors](https://img.shields.io/badge/all_contributors-11-orange.svg?style=flat-square)](#contributors-)
<!-- ALL-CONTRIBUTORS-BADGE:END -->

为了后续更好的维护EGO的Component，我们将该项目拆到了[https://github.com/ego-component](https://github.com/ego-component)，以下是EGO和Component的相关资料。

| Component Name            | Code                                                                  | Example                                                                        | Doc                                                                                                                 |
|---------------------------|-----------------------------------------------------------------------|--------------------------------------------------------------------------------|---------------------------------------------------------------------------------------------------------------------|
| HTTP Server               | [Code](https://github.com/gotomicro/ego/tree/master/server/egin)      | [Example](https://github.com/gotomicro/ego/tree/master/examples/server/http)   | [Doc](https://ego.gocn.vip/frame/server/http.html)                                                                  |
| gRPC Server               | [Code](./server/egrpc)                                                | [Example](https://github.com/gotomicro/ego/tree/master/examples/server/grpc)   | [Doc](https://ego.gocn.vip/frame/server/grpc.html#example)                                                          |
| Governance Service        | [Code](https://github.com/gotomicro/ego/tree/master/server/egovernor) | [Example](https://github.com/gotomicro/ego/tree/master/examples/server/governor) | [Doc](https://ego.gocn.vip/frame/server/governor.html)                                                              |
| Job                       | [Code](https://github.com/gotomicro/ego/tree/master/task/ejob)        | [Example](.https://github.com/gotomicro/ego/tree/master/examples/task/job)     | [Doc](https://ego.gocn.vip/frame/task/job.html)                                                                     |
| Corn job                  | [Code](https://github.com/gotomicro/ego/tree/master/task/ecron)       | [Example](https://github.com/gotomicro/ego/tree/master/examples/task/cron)     | [Doc](https://ego.gocn.vip/frame/task/cron.html#_3-%E5%B8%B8%E8%A7%84%E5%AE%9A%E6%97%B6%E4%BB%BB%E5%8A%A1)          |
| Distributed Scheduled Job | [Code](https://github.com/gotomicro/ego/tree/master/task/ecron)       | [Example](https://github.com/gotomicro/ego/tree/master/examples/task/cron)     | [Doc](https://ego.gocn.vip/frame/task/cron.html#_4-%E5%88%86%E5%B8%83%E5%BC%8F%E5%AE%9A%E6%97%B6%E4%BB%BB%E5%8A%A1) |
| HTTP Client               | [Code](https://github.com/gotomicro/ego/tree/master/client/ehttp)     | [Example](https://github.com/gotomicro/ego/tree/master/examples/http/client)   | [Doc](https://ego.gocn.vip/frame/client/http.html#example)                                                          |
| gRPC Client               | [Code](https://github.com/gotomicro/ego/tree/master/client/egrpc)     | [Example](https://github.com/gotomicro/ego/tree/master/examples/grpc/direct)   | [Doc](https://ego.gocn.vip/frame/client/grpc.html#_4-%E7%9B%B4%E8%BF%9Egrpc)                                        |
| gRPC Client using ETCD    | [Code](https://github.com/ego-component/tree/master/eetcd)            | [Example](https://github.com/ego-component/tree/master/eetcd/examples)         | [Doc](https://ego.gocn.vip/frame/client/grpc.html#_5-%E4%BD%BF%E7%94%A8etcd%E7%9A%84grpc)                           |
| gRPC Client using k8s     | [Code](https://github.com/ego-component/tree/master/ek8s)             | [Example](https://github.com/ego-component/tree/master/ek8s/examples)          | [Doc](https://ego.gocn.vip/frame/client/grpc.html#_6-%E4%BD%BF%E7%94%A8k8s%E7%9A%84grpc)                            |
| Sentinel                  | [Code](https://github.com/gotomicro/ego/tree/master/core/esentinel)   | [Example](https://github.com/gotomicro/ego/tree/master/examples/sentinel/http) | [Doc](https://ego.gocn.vip/frame/client/sentinel.html)                                                              |
| MySQL                     | [Code](https://github.com/ego-component/tree/master/egorm)            | [Example](https://github.com/ego-component/tree/master/egorm/examples)         | [Doc](https://ego.gocn.vip/frame/client/gorm.html#example)                                                          |
| Redis                     | [Code](https://github.com/ego-component/tree/master/eredis)           | [Example](https://github.com/ego-component/tree/master/eredis/examples)        | [Doc](https://ego.gocn.vip/frame/client/redis.html#example)                                                         |
| Redis Distributed lock    | [Code](https://github.com/ego-component/tree/master/eredis)           | [Example](https://github.com/ego-component/tree/master/eredis/examples)        | [Doc](https://ego.gocn.vip/frame/client/redis.html#example)                                                         |
| Mongo                     | [Code](https://github.com/ego-component/tree/master/emongo)           | [Example](https://github.com/ego-component/tree/master/emongo/examples)        | [Doc](https://ego.gocn.vip/frame/client/mongo.html)                                                                 |
| Kafka                     | [Code](https://github.com/ego-component/tree/master/ekafka)           | [Example](https://github.com/ego-component/tree/master/ekafka/examples)        | [Doc](https://ego.gocn.vip/frame/client/kafka.html)                                                                 |
| ETCD                      | [Code](https://github.com/ego-component/tree/master/eetcd)            | [Example](https://github.com/ego-component/tree/master/eetcd/examples)         | [Doc](https://ego.gocn.vip/frame/client/eetcd.html)                                                                 |
| K8S                       | [Code](https://github.com/ego-component/tree/master/ek8s)             | [Example](https://github.com/ego-component/tree/master/ek8s/examples)          | [Doc](https://ego.gocn.vip/frame/client/ek8s.html)                                                                  |
| Oauth2                    | [Code](https://github.com/ego-component/tree/master/eoauth2)          | [Example](https://github.com/ego-component/tree/master/eoauth2/examples)       ||


## Contributors

Thanks for these wonderful people:

<!-- ALL-CONTRIBUTORS-LIST:START - Do not remove or modify this section -->
<!-- prettier-ignore-start -->
<!-- markdownlint-disable -->
<table>
  <tr>
    <td align="center"><a href="https://github.com/askuy"><img src="https://avatars.githubusercontent.com/u/14119383?v=4" width="64px;" alt=""/><br /><sub><b>askuy</b></sub></a></td>
    <td align="center"><a href="https://github.com/sevennt"><img src="https://avatars.githubusercontent.com/u/10843736?v=4" width="64px;" alt=""/><br /><sub><b>Wei Zheng</b></sub></a></td>
    <td align="center"><a href="https://www.jianshu.com/u/f2b47e5528d8"><img src="https://avatars.githubusercontent.com/u/9923838?v=4" width="64px;" alt=""/><br /><sub><b>Ming Deng</b></sub></a></td>
    <td align="center"><a href="https://github.com/AaronJan"><img src="https://avatars.githubusercontent.com/u/4630940?v=4" width="64px;" alt=""/><br /><sub><b>AaronJan</b></sub></a></td>
    <td align="center"><a href="https://blog.gaoqixhb.com/"><img src="https://avatars.githubusercontent.com/u/4217102?v=4" width="64px;" alt=""/><br /><sub><b>yanjixiong</b></sub></a></td>
    <td align="center"><a href="http://blog.lincolnzhou.com/"><img src="https://avatars.githubusercontent.com/u/3911154?v=4" width="64px;" alt=""/><br /><sub><b>LincolnZhou</b></sub></a></td>
    <td align="center"><a href="https://www.duanlv.ltd"><img src="https://avatars.githubusercontent.com/u/20787331?v=4" width="64px;" alt=""/><br /><sub><b>Link Duan</b></sub></a></td>
  </tr>
  <tr>
    <td align="center"><a href="https://github.com/dandyhuang"><img src="https://avatars.githubusercontent.com/u/12603054?v=4" width="64px;" alt=""/><br /><sub><b>dandyhuang</b></sub></a></td>
    <td align="center"><a href="https://github.com/NeoyeElf"><img src="https://avatars.githubusercontent.com/u/6872731?v=4" width="64px;" alt=""/><br /><sub><b>刘文哲</b></sub></a></td>
    <td align="center"><a href="https://github.com/UnparalleledBeauty"><img src="https://avatars.githubusercontent.com/u/37238372?v=4" width="64px;" alt=""/><br /><sub><b>Carlos</b></sub></a></td>
    <td align="center"><a href="https://github.com/livepo"><img src="https://avatars.githubusercontent.com/u/6700352?v=4" width="64px;" alt=""/><br /><sub><b>qiandiao</b></sub></a></td>
    <td align="center"><a href="https://github.com/livepo"><img src="https://avatars.githubusercontent.com/u/17000001?s=400&u=0336f4726acdf82e13f4c19f28d6e1de39ea0b9b&v=4" width="64px;" alt=""/><br /><sub><b>soeluc</b></sub></a></td>
  </tr>
</table>

<!-- markdownlint-restore -->
<!-- prettier-ignore-end -->

<!-- ALL-CONTRIBUTORS-LIST:END -->
