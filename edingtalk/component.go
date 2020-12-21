package edingtalk

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-redis/redis"
	"github.com/gotomicro/ego-component/eredis"
	"github.com/gotomicro/ego/client/ehttp"
	"github.com/gotomicro/ego/core/elog"
	"sync"
	"time"
)

const PackageName = "component.edingtalk"

type Component struct {
	Config      *Config
	ehttp       *ehttp.Component
	eredis      *eredis.Component
	logger      *elog.Component
	locker      sync.Mutex
	accessToken string
}

// newComponent ...
func newComponent(compName string, config *Config, logger *elog.Component) *Component {
	ehttpClient := ehttp.DefaultContainer().Build(
		ehttp.WithDebug(config.Debug),
		ehttp.WithRawDebug(config.RawDebug),
		ehttp.WithAddr(Addr),
		ehttp.WithReadTimeout(config.ReadTimeout),
		ehttp.WithSlowLogThreshold(config.SlowLogThreshold),
		ehttp.WithEnableAccessInterceptor(config.EnableAccessInterceptor),
		ehttp.WithEnableAccessInterceptorReply(config.EnableAccessInterceptorReply),
	)

	return &Component{
		Config: config,
		ehttp:  ehttpClient,
		logger: logger,
		eredis: config.eredis,
	}
}

// 获取access_token
// https://ding-doc.dingtalk.com/document#/org-dev-guide/obtain-access_token
func (c *Component) GetAccessToken() (token string, err error) {
	var data AccessTokenResponse
	accessTokenBytes, err := c.eredis.GetBytes(c.Config.RedisPrefix + "/token")
	// 系统错误返回
	if err != nil && !errors.Is(err, redis.Nil) {
		err = fmt.Errorf("refresh access token get redis %w", err)
		return
	}

	// 如果redis没数据，说明过期，重新获取数据
	if errors.Is(err, redis.Nil) {
		resp, err := c.ehttp.R().Get(fmt.Sprintf(ApiGetToken, c.Config.AppKey, c.Config.AppSecret))
		if err != nil {
			err = fmt.Errorf("refresh access token get dingding err %w", err)
			return "", err
		}
		// todo 存在并发问题
		err = json.Unmarshal(resp.Body(), &data)
		if err != nil {
			err = fmt.Errorf("refresh access token json unmarshal %w", err)
			return "", err
		}

		bytes, err := json.Marshal(data)
		if err != nil {
			err = fmt.Errorf("refresh access token json marshal %w", err)
			return "", err
		}
		// -60，可以提前过期，更新token数据
		c.eredis.Set(c.Config.RedisPrefix+"/token", string(bytes), time.Duration(data.ExpireTime-60)*time.Second)
		return data.AccessToken, err
	}

	err = json.Unmarshal(accessTokenBytes, &data)
	if err != nil {
		err = fmt.Errorf("refresh access token json unmarshal2 %w", err)
		return "", err
	}
	token = data.AccessToken
	return
}

// 获取用户信息
// 接口文档 https://ding-doc.dingtalk.com/document#/org-dev-guide/userid
// 调试文档 https://open-dev.dingtalk.com/apiExplorer#/jsapi?api=runtime.permission.requestAuthCode
func (c *Component) GetUserInfo(code string) (user UserInfo, err error) {
	token, err := c.GetAccessToken()
	if err != nil {
		return UserInfo{}, fmt.Errorf("get user info token err %w", err)
	}
	resp, err := c.ehttp.R().Get(fmt.Sprintf(ApiGetUserInfo, token, code))
	if err != nil {
		return UserInfo{}, fmt.Errorf("get user info token err2 %w", err)
	}
	err = json.Unmarshal(resp.Body(), &user)
	if err != nil {
		err = fmt.Errorf("refresh access token json unmarshal %w", err)
		return UserInfo{}, err
	}
	return
}
