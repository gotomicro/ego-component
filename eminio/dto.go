package eminio

import (
	"errors"
	"strings"
)

var (
	errBucketNameEmpty   = errors.New("bucket name is empty")
	errObjectPrefixEmpty = errors.New("bucket name is empty")
	errDoneChannelEmpty  = errors.New("done channel is empty")
	errNullPoint         = errors.New("null point exception")
)

// ListObjectsRequest 列举存储桶里的对象请求结构体
type ListObjectsRequest struct {
	BucketName   string        `json:"bucketName"`   // 存储桶名称
	ObjectPrefix string        `json:"objectPrefix"` // 要列举的对象前缀
	Recursive    bool          `json:"recursive"`    // true代表递归查找，false代表类似文件夹查找，以'/'分隔，不查子文件夹
	DoneCh       chan struct{} `json:"doneCh"`       // 在该channel上结束ListObjects iterator的一个message`
}

func (l *ListObjectsRequest) Valid() error {
	if l == nil {
		return errNullPoint
	}
	l.BucketName = strings.TrimSpace(l.BucketName)
	if l.BucketName == "" {
		return errBucketNameEmpty
	}
	l.ObjectPrefix = strings.TrimSpace(l.ObjectPrefix)
	if l.ObjectPrefix == "" {
		return errObjectPrefixEmpty
	}
	if l.DoneCh == nil {
		return errDoneChannelEmpty
	}
	return nil
}
