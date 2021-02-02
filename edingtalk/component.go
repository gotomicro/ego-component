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
	"net/url"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
	"github.com/gotomicro/ego-component/eredis"
	"github.com/gotomicro/ego/client/ehttp"
	"github.com/gotomicro/ego/core/elog"
	"go.uber.org/zap"
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
func (c *Component) Oauth2SnsAuthorize(state string) string {
	// 安全验证，生成随机state，防止获取oa系统的url，登录该系统
	// state, err := genRandState()
	// if err != nil {
	//	elog.Error("Generating state string failed", zap.Error(err))
	//	return
	// }
	// hashedState := c.hashStateCode(state, c.Config.Oauth2AppSecret)
	// 最大300s
	// ctx.SetCookie(c.Config.Oauth2StateCookieName, url.QueryEscape(hashedState), 300, "/", "", false, true)
	// ctx.Redirect(http.StatusFound, fmt.Sprintf(Addr+ApiOauth2Redirect, c.Config.Oauth2AppKey, state, c.Config.Oauth2RedirectUri))
	return fmt.Sprintf(Addr+ApiOauth2SnsAuthorize, c.Config.Oauth2AppKey, state, c.Config.Oauth2RedirectUri)
}

func (c *Component) Oauth2Qrconnect(state string) string {
	return fmt.Sprintf(Addr+ApiOauth2Qrconnect, c.Config.Oauth2AppKey, state, c.Config.Oauth2RedirectUri)
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
	var data Oauth2UserUnionInfoResponse
	err = json.Unmarshal(resp.Body(), &data)
	if err != nil {
		return UserInfoDetail{}, fmt.Errorf("oauth2 user info json unmarlshal err %w", err)
	}

	if data.ErrCode != 0 {
		return UserInfoDetail{}, fmt.Errorf("oauth2 user info errcode error err %s", data.ErrMsg)
	}

	token, err := c.GetAccessToken()
	if err != nil {
		return UserInfoDetail{}, fmt.Errorf("oauth2 user info token err %w", err)
	}

	// 获取用户的userid信息
	oauth2UseridInfoResp, err := c.ehttp.R().Get(fmt.Sprintf(ApiGetUserIdByUnionId, token, data.UserInfo.UnionId))
	if err != nil {
		return UserInfoDetail{}, fmt.Errorf("oauth2 user info get http err2 %w", err)
	}

	var oauth2UseridInfo Oauth2UseridInfo
	err = json.Unmarshal(oauth2UseridInfoResp.Body(), &oauth2UseridInfo)
	if err != nil {
		return UserInfoDetail{}, fmt.Errorf("oauth2 user info json unmarlshal2 err %w", err)
	}

	if oauth2UseridInfo.ErrCode != 0 {
		return UserInfoDetail{}, fmt.Errorf("oauth2 user info errcode error err2 %s", oauth2UseridInfo.ErrMsg)
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

	if userInfoDetail.ErrCode != 0 {
		return UserInfoDetail{}, fmt.Errorf("oauth2 user info errcode error err2 %s", userInfoDetail.ErrMsg)
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

// 查询用户
// 接口文档 https://ding-doc.dingtalk.com/document/app/create-a-department-v2
// 调试文档 https://open-dev.dingtalk.com/apiExplorer#/jsapi?api=runtime.permission.requestAuthCode
func (c *Component) UserGet(uid string) (user *User, err error) {
	token, err := c.GetAccessToken()
	if err != nil {
		return nil, fmt.Errorf("get user info token err %w", err)
	}
	var res userGetRes
	resp, err := c.ehttp.R().SetBody(payload{"userid": uid}).SetResult(&res).Post(fmt.Sprintf(ApiUserGet, token))
	if err != nil {
		return nil, fmt.Errorf("user get request fail, %w", err)
	}
	if resp.StatusCode() != 200 || res.ErrCode != 0 {
		return nil, fmt.Errorf("user get fail, %s", res)
	}
	return &res.Result, nil
}

// 创建用户
// 接口文档 https://ding-doc.dingtalk.com/document/app/create-a-department-v2
// 调试文档 https://open-dev.dingtalk.com/apiExplorer#/jsapi?api=runtime.permission.requestAuthCode
func (c *Component) UserCreate(req userCreateReq) (userId string, err error) {
	token, err := c.GetAccessToken()
	if err != nil {
		return "", fmt.Errorf("get user info token err %w", err)
	}
	var res userCreateRes
	resp, err := c.ehttp.R().SetBody(req).SetResult(&res).Post(fmt.Sprintf(ApiUserCreate, token))
	if err != nil {
		return "", fmt.Errorf("user create request fail, %w", err)
	}
	if resp.StatusCode() != 200 || res.ErrCode != 0 {
		return "", fmt.Errorf("user create fail, %s", res)
	}
	return res.Result.UserId, nil
}

// 更新用户
// 接口文档 https://ding-doc.dingtalk.com/document/app/update-a-department-v2
// 调试文档 https://open-dev.dingtalk.com/apiExplorer#/jsapi?api=runtime.permission.requestAuthCode
func (c *Component) UserUpdate(req *userUpdateReq) (err error) {
	token, err := c.GetAccessToken()
	if err != nil {
		return fmt.Errorf("get user info token err %w", err)
	}
	var res OpenAPIResponse
	resp, err := c.ehttp.R().SetBody(req).SetResult(&res).Post(fmt.Sprintf(ApiUserUpdate, token))
	if err != nil {
		return fmt.Errorf("user update request fail, %w", err)
	}
	if resp.StatusCode() != 200 || res.ErrCode != 0 {
		return fmt.Errorf("user update fail, %d,%s", resp.StatusCode(), res)
	}
	return
}

// 删除用户
// 接口文档 https://ding-doc.dingtalk.com/document/app/update-a-department-v2
// 调试文档 https://open-dev.dingtalk.com/apiExplorer#/jsapi?api=runtime.permission.requestAuthCode
func (c *Component) UserDelete(uid string) (err error) {
	token, err := c.GetAccessToken()
	if err != nil {
		return fmt.Errorf("get user info token err %w", err)
	}
	var res OpenAPIResponse
	resp, err := c.ehttp.R().SetBody(payload{"userid": uid}).SetResult(&res).Post(fmt.Sprintf(ApiUserDelete, token))
	if err != nil {
		return fmt.Errorf("user update request fail, %w", err)
	}
	if resp.StatusCode() != 200 || res.ErrCode != 0 {
		return fmt.Errorf("user delete fail, %d,%s", resp.StatusCode(), res)
	}
	return
}

// 获取部门用户userid列表
// 接口文档 https://ding-doc.dingtalk.com/document/app/update-a-department-v2
// 调试文档 https://open-dev.dingtalk.com/apiExplorer#/jsapi?api=runtime.permission.requestAuthCode
func (c *Component) UserListID(did int) (userIds []string, err error) {
	token, err := c.GetAccessToken()
	if err != nil {
		return nil, fmt.Errorf("get user info token err %w", err)
	}
	var res userListIDRes
	resp, err := c.ehttp.R().SetBody(payload{"dept_id": did}).SetResult(&res).Post(fmt.Sprintf(ApiUserListID, token))
	if err != nil {
		return nil, fmt.Errorf("user listid request fail, %w", err)
	}
	if resp.StatusCode() != 200 || res.ErrCode != 0 {
		return nil, fmt.Errorf("user listid fail, %d,%s", resp.StatusCode(), res)
	}
	return res.Result.UserIDList, nil
}

// 获取部门详情
// 接口文档 https://ding-doc.dingtalk.com/document/app/create-a-department-v2
// 调试文档 https://open-dev.dingtalk.com/apiExplorer#/jsapi?api=runtime.permission.requestAuthCode
func (c *Component) DepartmentGet(did int) (dep *Department, err error) {
	token, err := c.GetAccessToken()
	if err != nil {
		return nil, fmt.Errorf("get user info token err %w", err)
	}
	var res departmentGetRes
	resp, err := c.ehttp.R().SetBody(payload{"dept_id": did}).SetResult(&res).Post(fmt.Sprintf(ApiDepartmentGet, token))
	if err != nil {
		return nil, fmt.Errorf("department get request fail, %w", err)
	}
	if resp.StatusCode() != 200 || res.ErrCode != 0 {
		return nil, fmt.Errorf("department get fail, %s", res)
	}
	return &res.Result, nil
}

// 创建部门
// 接口文档 https://ding-doc.dingtalk.com/document/app/create-a-department-v2
// 调试文档 https://open-dev.dingtalk.com/apiExplorer#/jsapi?api=runtime.permission.requestAuthCode
func (c *Component) DepartmentCreate(req departmentCreateReq) (deptId int, err error) {
	token, err := c.GetAccessToken()
	if err != nil {
		return 0, fmt.Errorf("get user info token err %w", err)
	}
	var res DepartmentCreateRes
	resp, err := c.ehttp.R().SetBody(req).SetResult(&res).Post(fmt.Sprintf(ApiDepartmentCreate, token))
	if err != nil {
		return 0, fmt.Errorf("department create request fail, %w", err)
	}
	if resp.StatusCode() != 200 || res.ErrCode != 0 {
		return 0, fmt.Errorf("department create fail, %s", res)
	}
	return res.Result.DeptId, nil
}

// 创建部门
// 接口文档 https://ding-doc.dingtalk.com/document/app/update-a-department-v2
// 调试文档 https://open-dev.dingtalk.com/apiExplorer#/jsapi?api=runtime.permission.requestAuthCode
func (c *Component) DepartmentUpdate(req *DepartmentUpdateReq) (err error) {
	token, err := c.GetAccessToken()
	if err != nil {
		return fmt.Errorf("get user info token err %w", err)
	}
	var res OpenAPIResponse
	resp, err := c.ehttp.R().SetBody(req).SetResult(&res).Post(fmt.Sprintf(ApiDepartmentUpdate, token))
	if err != nil {
		return fmt.Errorf("department update request fail, %w", err)
	}
	if resp.StatusCode() != 200 || res.ErrCode != 0 {
		return fmt.Errorf("department update fail, %d,%s", resp.StatusCode(), res)
	}
	return
}

// 创建部门
// 接口文档 https://ding-doc.dingtalk.com/document/app/delete-a-department-v2
// 调试文档 https://open-dev.dingtalk.com/apiExplorer#/jsapi?api=runtime.permission.requestAuthCode
func (c *Component) DepartmentDelete(did int) error {
	token, err := c.GetAccessToken()
	if err != nil {
		return fmt.Errorf("get user info token fail, %w", err)
	}
	var res OpenAPIResponse
	resp, err := c.ehttp.R().SetBody(payload{"dept_id": did}).SetResult(&res).Post(fmt.Sprintf(ApiDepartmentDelete, token))
	if err != nil {
		return fmt.Errorf("department delete request fail, %w", err)
	}
	if resp.StatusCode() != 200 || res.ErrCode != 0 {
		return fmt.Errorf("department delete fail, %s", res)
	}
	return nil
}
