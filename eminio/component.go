package eminio

import (
	"context"
	"fmt"
	"strings"

	"github.com/gotomicro/ego/core/elog"
	"github.com/minio/minio-go/v6"
)

const packageName = "component.eminio"

type Component struct {
	config *config
	client *minio.Client
	logger *elog.Component
}

func newComponent(compName string, config *config, logger *elog.Component) *Component {
	var (
		client          *minio.Client
		endpoint        = config.Endpoint
		region          = config.Region
		accessKeyID     = config.AccessKeyID
		secretAccessKey = config.SecretAccessKey
		ssl             = config.Ssl
		err             error
	)
	if region != "" {
		if !checkRegion(region) {
			panic("无效的region:" + region)
		}
		client, err = minio.NewWithRegion(endpoint, accessKeyID, secretAccessKey, ssl, region)
		if err != nil {
			panic("new minioClient with region failed:" + err.Error())
		}
	} else {
		client, err = minio.New(endpoint, accessKeyID, secretAccessKey, ssl)
		if err != nil {
			panic("new minioClient failed:" + err.Error())
		}
	}
	return &Component{
		config: config,
		client: client,
		logger: logger,
	}
}

// Client 暴露 minio 原生 client
func (c *Component) Client() *minio.Client {
	return c.client
}

// MakeBucketWithContext 创建一个存储桶
// bucketName: 存储桶名称
// location: 存储桶被创建的region(地区)，默认为us-east-1(美国东一区)
// 如果指定region的话，则region值必须在 regionMap 中存在
func (c *Component) MakeBucketWithContext(ctx context.Context, bucketName string, location string) error {
	bucketName = strings.TrimSpace(bucketName)
	location = strings.TrimSpace(location)
	if bucketName == "" {
		return errBucketNameEmpty
	}
	if location == "" {
		location = usEast1
	}
	if !checkRegion(location) {
		return fmt.Errorf("invalid location:%s", location)
	}
	if err := c.client.MakeBucketWithContext(ctx, bucketName, location); err != nil {
		return err
	}
	return nil
}

// ListBucketsWithContext 列出所有的存储桶
func (c *Component) ListBucketsWithContext(ctx context.Context) ([]minio.BucketInfo, error) {
	return c.client.ListBucketsWithContext(ctx)
}

// BucketExistsWithContext 检查存储桶是否存在
func (c *Component) BucketExistsWithContext(ctx context.Context, bucketName string) (bool, error) {
	bucketName = strings.TrimSpace(bucketName)
	if bucketName == "" {
		return false, errBucketNameEmpty
	}
	return c.client.BucketExistsWithContext(ctx, bucketName)
}

// RemoveBucket 删除一个存储桶，存储桶必须为空才能被成功删除
func (c *Component) RemoveBucket(ctx context.Context, bucketName string) error {
	bucketName = strings.TrimSpace(bucketName)
	if bucketName == "" {
		return errBucketNameEmpty
	}
	return c.client.RemoveBucket(bucketName)
}

// ListObjects 列举存储桶里的对象
func (c *Component) ListObjects(ctx context.Context, params *ListObjectsRequest) (<-chan minio.ObjectInfo, error) {
	if err := params.Valid(); err != nil {
		return nil, err
	}
	ch := c.client.ListObjects(params.BucketName, params.ObjectPrefix, params.Recursive, params.DoneCh)
	return ch, nil
}
