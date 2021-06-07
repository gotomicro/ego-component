## 查看全部租期
```
# 查看全部租期
etcdctl lease list
found 1 leases
694d79ada4e6c82c
# 查看某个租期信息
etcdctl lease timetolive 694d79ada4e6c82c --keys
lease 694d79ada4e6c82c granted with TTL(10s), remaining(9s), attached keys([/ego/main/providers/grpc://0.0.0.0:9003])
```

