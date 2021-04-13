package main

import (
	"fmt"

	"github.com/gotomicro/ego"
	"github.com/gotomicro/ego/client/egrpc"
	"github.com/gotomicro/ego/core/elog"

	"github.com/gotomicro/ego-component/ek8s"
	"github.com/gotomicro/ego-component/ek8s/examples/kubegrpc/helloworld"
	"github.com/gotomicro/ego-component/ek8s/registry"
)

func main() {
	if err := ego.New().Invoker(
		invokerGrpc,
	).Run(); err != nil {
		elog.Error("startup", elog.FieldErr(err))
	}
}

func invokerGrpc() error {
	// 构建k8s registry，并注册为grpc resolver
	registry.Load("registry").Build(
		registry.WithClient(ek8s.Load("k8s").Build()),
	)
	// 构建gRPC.ClientConn组件
	grpcConn := egrpc.Load("grpc.test").Build()
	// 构建gRPC Client组件
	grpcComp := helloworld.NewGreeterClient(grpcConn.ClientConn)
	fmt.Printf("client--------------->"+"%+v\n", grpcComp)
	return nil
}
