package main

import (
	"context"
	"fmt"

	"github.com/gotomicro/ego"
	"github.com/gotomicro/ego/core/elog"

	"github.com/gotomicro/ego-component/eminio"
)

func main() {
	err := ego.New().Invoker(invokerMinio).Run()
	if err != nil {
		elog.Error("startup", elog.Any("err", err))
	}
}

func invokerMinio() error {
	comp := eminio.Load("minio").Build()
	ctx := context.Background()
	bucketName := "my-images"
	// 使用默认的region创建 bucket
	if err := comp.MakeBucketWithContext(ctx, bucketName, ""); err != nil {
		return fmt.Errorf("make bucket failed:%w", err)
	}
	// 查询所有的bucket
	buckets, err := comp.ListBucketsWithContext(ctx)
	if err != nil {
		return fmt.Errorf("list bucket failed:%w", err)
	}
	fmt.Printf("all buckets:%v\n", buckets)
	// 删除指定的bucket
	if err := comp.RemoveBucket(ctx, bucketName); err != nil {
		return fmt.Errorf("remove bucket failed:%w", err)
	}
	return nil
}
