package ealiyun

import (
	"errors"

	openapi "github.com/alibabacloud-go/darabonba-openapi/client"
	ims20190815 "github.com/alibabacloud-go/ims-20190815/v2/client"
	"github.com/alibabacloud-go/tea/tea"
	"github.com/gotomicro/ego/core/elog"
	"go.uber.org/zap"
)

var (
	errEmptyResult = errors.New("empty result")
)

const packageName = "component.ealiyun"

type Component struct {
	config    *config
	logger    *elog.Component
	ramClient *ims20190815.Client // openApi Client 内嵌在该Client中
}

func newComponent(config *config, logger *elog.Component) *Component {
	openApiConfig := &openapi.Config{
		AccessKeyId:     tea.String(config.AccessKeyId),
		AccessKeySecret: tea.String(config.AccessKeySecret),
		Endpoint:        tea.String(config.Endpoint),
	}
	rawClient, err := ims20190815.NewClient(openApiConfig)
	if err != nil {
		panic("newClient fail:" + err.Error())
	}
	return &Component{
		config:    config,
		ramClient: rawClient,
		logger:    logger,
	}
}

// CreateRamUser ...
func (c *Component) CreateRamUser(req SaveRamUserRequest) (*RamUserResponse, error) {
	res, err := c.ramClient.CreateUser(&ims20190815.CreateUserRequest{
		UserPrincipalName: tea.String(req.UserPrincipalName),
		DisplayName:       tea.String(req.DisplayName),
		MobilePhone:       tea.String(req.MobilePhone),
		Email:             tea.String(req.Email),
		Comments:          tea.String(req.Comments),
	})
	if err != nil {
		return nil, err
	}
	if res == nil {
		return nil, errEmptyResult
	}
	c.logger.Info("Component-ealiyun", zap.Any("CreateRamUser-res", res))
	body := res.Body
	return &RamUserResponse{
		RequestID: *body.RequestId,
		User: RamUserInfo{
			UserID:            *body.User.UserId,
			CreateDate:        *body.User.CreateDate,
			UserPrincipalName: *body.User.UserPrincipalName,
			DisplayName:       *body.User.DisplayName,
			MobilePhone:       *body.User.MobilePhone,
			Email:             *body.User.Email,
			Comments:          *body.User.Comments,
			UpdateDate:        *body.User.UpdateDate,
		},
	}, nil
}

// DelRamUser 删除用户前，需要保证用户不拥有任何权限且不属于任何用户组。
func (c *Component) DelRamUser(userPrincipalName string) error {
	res, err := c.ramClient.DeleteUser(&ims20190815.DeleteUserRequest{
		UserPrincipalName: tea.String(userPrincipalName),
	})
	if err != nil {
		return err
	}
	c.logger.Info("Component-ealiyun", zap.Any("DelRamUser-res", res))
	return nil
}

// GetRamUser ...
func (c *Component) GetRamUser(userPrincipalName string) (*RamUserResponse, error) {
	res, err := c.ramClient.GetUser(&ims20190815.GetUserRequest{
		UserPrincipalName: tea.String(userPrincipalName),
	})
	if err != nil {
		return nil, err
	}
	if res == nil {
		return nil, errEmptyResult
	}
	c.logger.Info("Component-ealiyun", zap.Any("GetUser-res", res))
	body := res.Body
	return &RamUserResponse{
		RequestID: *body.RequestId,
		User: RamUserInfo{
			UserID:            *body.User.UserId,
			CreateDate:        *body.User.CreateDate,
			UserPrincipalName: *body.User.UserPrincipalName,
			DisplayName:       *body.User.DisplayName,
			MobilePhone:       *body.User.MobilePhone,
			Email:             *body.User.Email,
			Comments:          *body.User.Comments,
			LastLoginDate:     *body.User.LastLoginDate,
			UpdateDate:        *body.User.UpdateDate,
		},
	}, nil
}
