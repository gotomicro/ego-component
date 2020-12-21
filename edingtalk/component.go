package edingtalk

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
	"github.com/gotomicro/ego-component/eredis"
	"github.com/gotomicro/ego/client/ehttp"
	"github.com/gotomicro/ego/core/elog"
	"go.uber.org/zap"
	"net/http"
	"net/url"
	"strconv"
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
	accessTokenBytes, err := c.eredis.GetBytes(c.Config.RedisPrefix + c.Config.RedisBaseToken)
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
		c.eredis.Set(c.Config.RedisPrefix+c.Config.RedisBaseToken, string(bytes), time.Duration(data.ExpireTime-60)*time.Second)
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

// 获取跳转地址
// https://ding-doc.dingtalk.com/document#/org-dev-guide/etaarr
func (c *Component) Oauth2RedirectUri(ctx *gin.Context) {
	// 安全验证，生成随机state，防止获取oa系统的url，登录该系统
	state, err := genRandState()
	if err != nil {
		elog.Error("Generating state string failed", zap.Error(err))
		return
	}
	hashedState := c.hashStateCode(state, c.Config.Oauth2AppSecret)
	// 最大300s
	ctx.SetCookie(c.Config.Oauth2StateCookieName, url.QueryEscape(hashedState), 300, "/", "", false, true)
	ctx.Redirect(http.StatusFound, fmt.Sprintf(Addr+ApiOauth2Redirect, c.Config.Oauth2AppKey, hashedState, c.Config.Oauth2RedirectUri))
}

// 根据code，获取用户信息
// todo code state
func (c *Component) Oauth2UserInfo(code string) (user UserInfoDetail, err error) {
	timestamp := strconv.FormatInt(time.Now().UnixNano()/1000000, 10) // 毫秒时间戳
	signature := encryptHMAC(timestamp, c.Config.Oauth2AppSecret)
	// 获取用户的union信息
	resp, err := c.ehttp.R().SetBody(gin.H{"tmp_auth_code": code}).Post(fmt.Sprintf(ApiGetUserInfoByCode, c.Config.Oauth2AppKey, timestamp, signature))
	if err != nil {
		return UserInfoDetail{}, fmt.Errorf("oauth2 user info get http err %w", err)
	}
	var data Oauth2UserUnionInfo
	err = json.Unmarshal(resp.Body(), &data)
	if err != nil {
		return UserInfoDetail{}, fmt.Errorf("oauth2 user info json unmarlshal err %w", err)
	}

	token, err := c.GetAccessToken()
	if err != nil {
		return UserInfoDetail{}, fmt.Errorf("oauth2 user info token err %w", err)
	}

	// 获取用户的userid信息
	oauth2UseridInfoResp, err := c.ehttp.R().Get(fmt.Sprintf(ApiGetUserIdByUnionId, token, data.UnionId))
	if err != nil {
		return UserInfoDetail{}, fmt.Errorf("oauth2 user info get http err2 %w", err)
	}

	var oauth2UseridInfo Oauth2UseridInfo
	err = json.Unmarshal(oauth2UseridInfoResp.Body(), &oauth2UseridInfo)
	if err != nil {
		return UserInfoDetail{}, fmt.Errorf("oauth2 user info json unmarlshal2 err %w", err)
	}

	// 获取用户的详细信息
	userInfoDetailResp, err := c.ehttp.R().Get(fmt.Sprintf(ApiGetUserDetail, token, oauth2UseridInfo.UserId))
	if err != nil {
		return UserInfoDetail{}, fmt.Errorf("oauth2 user info get http err3 %w", err)
	}
	var userInfoDetail UserInfoDetail
	err = json.Unmarshal(userInfoDetailResp.Body(), &userInfoDetail)
	if err != nil {
		return UserInfoDetail{}, fmt.Errorf("oauth2 user info json unmarlshal2 err %w", err)
	}
	return userInfoDetail, nil
}

func genRandState() (string, error) {
	rnd := make([]byte, 32)
	if _, err := rand.Read(rnd); err != nil {
		elog.Error("failed to generate state string", zap.Error(err))
		return "", err
	}
	return base64.URLEncoding.EncodeToString(rnd), nil
}

func (c *Component) hashStateCode(code, seed string) string {
	hashBytes := sha256.Sum256([]byte(code + c.Config.Oauth2AppKey + seed))
	return hex.EncodeToString(hashBytes[:])
}

//
//func encryptHMAC(paramsJoin string, secret string) []byte {
//	hHmac := hmac.New(md5.New, []byte(secret))
//	hHmac.Write([]byte(paramsJoin))
//	return hHmac.Sum([]byte(""))
//}

func encryptHMAC(message, secret string) string {
	// 钉钉签名算法实现
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(message))
	sum := h.Sum(nil) // 二进制流
	message1 := base64.StdEncoding.EncodeToString(sum)

	uv := url.Values{}
	uv.Add("0", message1)
	message2 := uv.Encode()[2:]
	return message2

}
