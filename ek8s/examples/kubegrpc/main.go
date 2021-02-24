package main

import (
	"fmt"
	ossv1 "git.shimo.im/gopkg/pb/infra/oss/v1"
	"github.com/gotomicro/ego"
	"github.com/gotomicro/ego-component/ekubernetes"
	"github.com/gotomicro/ego-component/ekubernetes/registry"
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
	kubClient := ek8s.Load("kubernetes").Build()
	kubeRegistry := registry.Load("registry").Build(registry.WithClientKubernetes(kubClient))
	resolver.Register("kubernetes", kubeRegistry)
	grpcConn := egrpc.Load("grpc.test").Build()
	client := ossv1.NewOssClient(grpcConn.ClientConn)
	fmt.Printf("client--------------->"+"%+v\n", client)
	return nil
}
