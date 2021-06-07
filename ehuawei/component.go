package ehuawei

import (
	"errors"

	"github.com/gotomicro/ego/core/elog"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/auth/global"
	iam "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/iam/v3"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/iam/v3/model"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/iam/v3/region"
	"go.uber.org/zap"
)

var (
	errEmptyResult = errors.New("empty result")
)

const packageName = "component.ehuawei"

type Component struct {
	config    *config
	logger    *elog.Component
	IamClient *iam.IamClient
}

func newComponent(config *config, logger *elog.Component) *Component {
	auth := global.NewCredentialsBuilder().
		WithAk(config.AK).
		WithSk(config.SK).
		Build()
	iamClient := iam.NewIamClient(
		iam.IamClientBuilder().
			WithRegion(region.ValueOf("cn-east-3")).
			WithCredential(auth).
			Build())

	return &Component{
		config:    config,
		logger:    logger,
		IamClient: iamClient,
	}
}

// KeystoneListGroups  查询用户组列表
// domainId 用户组所属账号ID
// link :  https://apiexplorer.developer.huaweicloud.com/apiexplorer/sdk?product=IAM&api=KeystoneListGroups
func (c *Component) KeystoneListGroups(domainId string) (*model.KeystoneListGroupsResponse, error) {
	res, err := c.IamClient.KeystoneListGroups(&model.KeystoneListGroupsRequest{
		DomainId: &domainId,
	})
	if err != nil {
		return nil, err
	}
	if res == nil {
		return nil, errEmptyResult
	}
	c.logger.Info("Component-ehuawei", zap.Any("keystoneListGroups-res", res))
	return res, nil
}
