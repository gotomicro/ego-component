package main

import (
	"fmt"
	"github.com/gotomicro/ego"
	"github.com/gotomicro/ego-component/ek8s"
	"github.com/gotomicro/ego-component/ek8s/examples/kubegrpc/helloworld"
	"github.com/gotomicro/ego-component/ek8s/registry"
	"github.com/gotomicro/ego/client/egrpc"
	"github.com/gotomicro/ego/client/egrpc/resolver"
	"github.com/gotomicro/ego/core/elog"
)

func main() {
	if err := ego.New().Invoker(
		invokerGrpc,
	).Run(); err != nil {
		elog.Error("startup", elog.FieldErr(err))
	}
}

func invokerGrpc() error {
	resolver.Register("k8s",  registry.Load("registry").Build(registry.WithClientK8s(ek8s.Load("k8s").Build())))
	grpcConn := egrpc.Load("grpc.test").Build()
	grpcComp := helloworld.NewGreeterClient(grpcConn.ClientConn)
	fmt.Printf("client--------------->"+"%+v\n", grpcComp)
	return nil
}
