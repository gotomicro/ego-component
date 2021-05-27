package ealiyun

import (
	"errors"
	openapi "github.com/alibabacloud-go/darabonba-openapi/client"
	ram20150501 "github.com/alibabacloud-go/ram-20150501/client"
	"github.com/alibabacloud-go/tea/tea"
	"github.com/gotomicro/ego/core/elog"
	"go.uber.org/zap"
)

var (
	errEmptyResult = errors.New("empty result")
)

const PackageName = "component.ealiyun"

type Component struct {
	config    *config
	logger    *elog.Component
	ramClient *ram20150501.Client // openApi Client 内嵌在该Client中
}

func newComponent(config *config, logger *elog.Component) *Component {
	openApiConfig := &openapi.Config{
		AccessKeyId:     tea.String(config.AccessKeyId),
		AccessKeySecret: tea.String(config.AccessKeySecret),
		Endpoint:        tea.String(config.Endpoint),
	}
	rawClient, err := ram20150501.NewClient(openApiConfig)
	if err != nil {
		panic("newClient fail:" + err.Error())
	}
	return &Component{
		config:    config,
		ramClient: rawClient,
		logger:    logger,
	}
}

func (c *Component) CreateRamUser(req SaveRamUserRequest) (*RamUserResponse, error) {
	res, err := c.ramClient.CreateUser(&ram20150501.CreateUserRequest{
		UserName:    tea.String(req.UserName),
		DisplayName: tea.String(req.DisplayName),
		MobilePhone: tea.String(req.MobilePhone),
		Email:       tea.String(req.Email),
		Comments:    tea.String(req.Comments),
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
			UserID:      *body.User.UserId,
			CreateDate:  *body.User.CreateDate,
			UserName:    *body.User.UserName,
			DisplayName: *body.User.DisplayName,
			MobilePhone: *body.User.MobilePhone,
			Email:       *body.User.Email,
			Comments:    *body.User.Comments,
		},
	}, nil
}

func (c *Component) GetRamUser(userName string) (*RamUserResponse, error) {
	res, err := c.ramClient.GetUser(&ram20150501.GetUserRequest{
		UserName: tea.String(userName),
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
			UserID:        *body.User.UserId,
			CreateDate:    *body.User.CreateDate,
			UserName:      *body.User.UserName,
			DisplayName:   *body.User.DisplayName,
			MobilePhone:   *body.User.MobilePhone,
			Email:         *body.User.Email,
			Comments:      *body.User.Comments,
			LastLoginDate: *body.User.LastLoginDate,
		},
	}, nil
}
