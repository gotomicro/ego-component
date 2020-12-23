package main

import (
	"context"
	"github.com/gotomicro/ego"
	"github.com/gotomicro/ego-component/eetcd"
	"github.com/gotomicro/ego-component/eetcd/examples/helloworld"
	"github.com/gotomicro/ego-component/eetcd/registry"
	"github.com/gotomicro/ego/client/egrpc"
	"github.com/gotomicro/ego/client/egrpc/resolver"
	"github.com/gotomicro/ego/core/elog"
)

func main() {
	if err := ego.New().Invoker(
		invokerGrpc,
		callGrpc,
	).Run(); err != nil {
		elog.Error("startup", elog.FieldErr(err))
	}
}

var grpcComp helloworld.GreeterClient

func invokerGrpc() error {
	EtcdClient := eetcd.Load("etcd").Build()
	EtcdRegistry := registry.Load("registry").Build(registry.WithClientEtcd(EtcdClient))
	// 必须注册在grpc前面
	resolver.Register("etcd", EtcdRegistry)
	grpcConn := egrpc.Load("grpc.test").Build()
	grpcComp = helloworld.NewGreeterClient(grpcConn.ClientConn)
	return nil
}

func callGrpc() error {
	_, err := grpcComp.SayHello(context.Background(), &helloworld.HelloRequest{
		Name: "i am client",
	})
	if err != nil {
		return err
	}

	_, err = grpcComp.SayHello(context.Background(), &helloworld.HelloRequest{
		Name: "error",
	})
	if err != nil {
		return err
	}
	return nil
}
