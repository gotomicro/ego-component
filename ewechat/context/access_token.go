package context

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"

	"github.com/gotomicro/ego-component/ewechat/util"
)

const (
	// AccessTokenURL 获取access_token的接口
	AccessTokenURL = "https://api.weixin.qq.com/cgi-bin/token"
)

// ResAccessToken struct
type ResAccessToken struct {
	util.CommonError

	AccessToken string `json:"access_token"`
	ExpiresIn   int64  `json:"expires_in"`
}

// GetAccessTokenFunc 获取 access token 的函数签名
type GetAccessTokenFunc func(ctx *Context) (accessToken string, err error)

// SetAccessTokenLock 设置读写锁（一个appID一个读写锁）
func (ctx *Context) SetAccessTokenLock(l *sync.RWMutex) {
	ctx.accessTokenLock = l
}

// SetGetAccessTokenFunc 设置自定义获取accessToken的方式, 需要自己实现缓存
func (ctx *Context) SetGetAccessTokenFunc(f GetAccessTokenFunc) {
	ctx.accessTokenFunc = f
}

// GetAccessToken 获取access_token
func (ctx *Context) GetAccessToken() (accessToken string, err error) {
	ctx.accessTokenLock.Lock()
	defer ctx.accessTokenLock.Unlock()

	if ctx.accessTokenFunc != nil {
		return ctx.accessTokenFunc(ctx)
	}
	accessTokenCacheKey := fmt.Sprintf("access_token_%s", ctx.AppID)
	accessToken, err = ctx.Cache.Get(context.Background(), accessTokenCacheKey)
	if err != nil && !errors.Is(err, redis.Nil) {
		return "", err
	}
	if accessToken != "" {
		return accessToken, nil
	}

	// 从微信服务器获取
	var resAccessToken ResAccessToken
	resAccessToken, err = ctx.GetAccessTokenFromServer()
	if err != nil {
		err = fmt.Errorf("get access token err %w", err)
		return
	}

	accessToken = resAccessToken.AccessToken
	return
}

// GetAccessTokenFromServer 强制从微信服务器获取token
func (ctx *Context) GetAccessTokenFromServer() (resAccessToken ResAccessToken, err error) {
	url := fmt.Sprintf("%s?grant_type=client_credential&appid=%s&secret=%s", AccessTokenURL, ctx.AppID, ctx.AppSecret)
	var body []byte
	body, err = ctx.HTTPGet(url)
	if err != nil {
		return
	}
	err = json.Unmarshal(body, &resAccessToken)
	if err != nil {
		err = fmt.Errorf("access token from server parse json err %w", err)
		return
	}
	if resAccessToken.ErrMsg != "" {
		err = fmt.Errorf("get access_token error : errcode=%v , errormsg=%v", resAccessToken.ErrCode, resAccessToken.ErrMsg)
		return
	}

	accessTokenCacheKey := fmt.Sprintf("access_token_%s", ctx.AppID)
	expires := resAccessToken.ExpiresIn - 1500
	err = ctx.Cache.Set(context.Background(), accessTokenCacheKey, resAccessToken.AccessToken, time.Duration(expires)*time.Second)
	if err != nil {
		err = fmt.Errorf("set token error %w", err)
		return
	}
	return
}
