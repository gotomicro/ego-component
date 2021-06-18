package edingtalk

import (
	"context"
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
	"github.com/go-redis/redis/v8"
	"github.com/gotomicro/ego-component/eredis"
	"github.com/gotomicro/ego/client/ehttp"
	"github.com/gotomicro/ego/core/elog"
	"go.uber.org/zap"
)

const PackageName = "component.edingtalk"

type Component struct {
	config      *config
	ehttp       *ehttp.Component
	eredis      *eredis.Component
	logger      *elog.Component
	locker      sync.Mutex
	accessToken string
}

// newComponent ...
func newComponent(compName string, config *config, logger *elog.Component) *Component {
	ehttpClient := ehttp.DefaultContainer().Build(
		ehttp.WithDebug(config.Debug),
		ehttp.WithRawDebug(config.RawDebug),
		ehttp.WithAddr(Addr),
		ehttp.WithReadTimeout(config.ReadTimeout),
		ehttp.WithSlowLogThreshold(config.SlowLogThreshold),
		ehttp.WithEnableAccessInterceptor(config.EnableAccessInterceptor),
		ehttp.WithEnableAccessInterceptorRes(config.EnableAccessInterceptorRes),
	)

	return &Component{
		config: config,
		ehttp:  ehttpClient,
		logger: logger,
		eredis: config.eredis,
	}
}

// 获取access_token
// https://ding-doc.dingtalk.com/document#/org-dev-guide/obtain-access_token
func (c *Component) GetAccessToken() (token string, err error) {
	var data AccessTokenResponse
	accessTokenBytes, err := c.eredis.GetBytes(context.Background(), c.config.RedisPrefix+c.config.RedisBaseToken)
	// 系统错误返回
	if err != nil && !errors.Is(err, redis.Nil) {
		return "", fmt.Errorf("refresh access token get redis %w", err)
	}

	// 如果redis没数据，说明过期，重新获取数据
	if errors.Is(err, redis.Nil) {
		_, err := c.ehttp.R().SetResult(&data).Get(fmt.Sprintf(ApiGetToken, c.config.AppKey, c.config.AppSecret))
		if err != nil {
			return "", fmt.Errorf("refresh access token get dingding fail, %w", err)
		}
		// todo 存在并发问题
		if data.ErrCode != 0 {
			return "", fmt.Errorf("get access token fail, %w", err)
		}
		bytes, err := json.Marshal(data)
		if err != nil {
			return "", fmt.Errorf("refresh access token json marshal fail, %w", err)
		}
		// -60，可以提前过期，更新token数据
		err = c.eredis.Set(context.Background(), c.config.RedisPrefix+c.config.RedisBaseToken, string(bytes), time.Duration(data.ExpireTime-60)*time.Second)
		if err != nil {
			return "", fmt.Errorf("set access token to redis fail, %w", err)
		}
		return data.AccessToken, err
	}

	if err = json.Unmarshal(accessTokenBytes, &data); err != nil {
		return "", fmt.Errorf("refresh access token json unmarshal fail, %w", err)
	}

	return data.AccessToken, nil
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
	// hashedState := c.hashStateCode(state, c.config.Oauth2AppSecret)
	// 最大300s
	// ctx.SetCookie(c.config.Oauth2StateCookieName, url.QueryEscape(hashedState), 300, "/", "", false, true)
	// ctx.Redirect(http.StatusFound, fmt.Sprintf(Addr+ApiOauth2Redirect, c.config.Oauth2AppKey, state, c.config.Oauth2RedirectUri))
	return fmt.Sprintf(Addr+ApiOauth2SnsAuthorize, c.config.Oauth2AppKey, state, c.config.Oauth2RedirectUri)
}

func (c *Component) Oauth2Qrconnect(state string) string {
	return fmt.Sprintf(Addr+ApiOauth2Qrconnect, c.config.Oauth2AppKey, state, c.config.Oauth2RedirectUri)
}

// 根据code，获取用户信息
// todo code state
func (c *Component) Oauth2UserInfo(code string) (user UserInfoDetail, err error) {
	timestamp := strconv.FormatInt(time.Now().UnixNano()/1000000, 10) // 毫秒时间戳
	signature := encryptHMAC(timestamp, c.config.Oauth2AppSecret)
	// 获取用户的union信息
	resp, err := c.ehttp.R().SetBody(gin.H{"tmp_auth_code": code}).Post(fmt.Sprintf(ApiGetUserInfoByCode, c.config.Oauth2AppKey, timestamp, signature))
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
	hashBytes := sha256.Sum256([]byte(code + c.config.Oauth2AppKey + seed))
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
func (c *Component) UserGet(uid string) (*User, error) {
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
func (c *Component) UserCreate(req UserCreateReq) (string, error) {
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
func (c *Component) UserUpdate(req *UserUpdateReq) error {
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
	return nil
}

// 删除用户
// 接口文档 https://ding-doc.dingtalk.com/document/app/update-a-department-v2
// 调试文档 https://open-dev.dingtalk.com/apiExplorer#/jsapi?api=runtime.permission.requestAuthCode
func (c *Component) UserDelete(uid string) error {
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
	return nil
}

// 获取部门用户userid列表
// 接口文档 https://ding-doc.dingtalk.com/document/app/update-a-department-v2
// 调试文档 https://open-dev.dingtalk.com/apiExplorer#/jsapi?api=runtime.permission.requestAuthCode
func (c *Component) UserListID(did int) ([]string, error) {
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

// 获取部门用户详情，注意size最大为100，超过100钉钉会报错
// 接口文档 https://ding-doc.dingtalk.com/document/app/update-a-department-v2
// 调试文档 https://open-dev.dingtalk.com/apiExplorer#/jsapi?api=runtime.permission.requestAuthCode
func (c *Component) UserList(did, cursor, size int) (*UserListRes, error) {
	token, err := c.GetAccessToken()
	if err != nil {
		return nil, fmt.Errorf("get user info token err %w", err)
	}
	var res userListRes
	resp, err := c.ehttp.R().SetBody(payload{
		"dept_id": did,
		"cursor":  cursor,
		"size":    size,
	}).SetResult(&res).Post(fmt.Sprintf(ApiUserList, token))
	if err != nil {
		return nil, fmt.Errorf("user listid request fail, %w", err)
	}
	if resp.StatusCode() != 200 || res.ErrCode != 0 {
		return nil, fmt.Errorf("user listid fail, %d,%s", resp.StatusCode(), res)
	}
	return res.Result, nil
}

// 获取部门详情
// 接口文档 https://ding-doc.dingtalk.com/document/app/create-a-department-v2
// 调试文档 https://open-dev.dingtalk.com/apiExplorer#/jsapi?api=runtime.permission.requestAuthCode
func (c *Component) DepartmentGet(did int) (*Department, error) {
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
func (c *Component) DepartmentCreate(req DepartmentCreateReq) (int, error) {
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
func (c *Component) DepartmentUpdate(req *DepartmentUpdateReq) error {
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
	return nil
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

// 获取部门下一级部门列表
// 接口文档 https://ding-doc.dingtalk.com/document/app/obtain-the-department-list-v2
// 调试文档 https://open-dev.dingtalk.com/apiExplorer#/jsapi?api=runtime.permission.requestAuthCode
func (c *Component) DepartmentListsub(did int) ([]Department, error) {
	token, err := c.GetAccessToken()
	if err != nil {
		return nil, fmt.Errorf("get user info token fail, %w", err)
	}
	var res DepartmentListsubRes
	resp, err := c.ehttp.R().SetBody(payload{"dept_id": did}).SetResult(&res).Post(fmt.Sprintf(ApiDepartmentListsub, token))
	if err != nil {
		return nil, fmt.Errorf("department listsub request fail, %w", err)
	}
	if resp.StatusCode() != 200 || res.ErrCode != 0 {
		return nil, fmt.Errorf("department listsub fail, %s", res)
	}
	return res.Result, nil
}

// 获取部门树，递归查询全部子部门
// NOTICE: 只能查询Name,DeptID,CreateDeptGroup,ParentID,SubDeptList字段
// 接口文档 https://ding-doc.dingtalk.com/document/app/obtain-the-department-list
// 调试文档 https://open-dev.dingtalk.com/apiExplorer#/jsapi?api=runtime.permission.requestAuthCode
func (c *Component) DepartmentTree(did int) (*Department, error) {
	token, err := c.GetAccessToken()
	if err != nil {
		return nil, fmt.Errorf("get user info token fail, %w", err)
	}
	var res departmentListRes
	resp, err := c.ehttp.R().SetBody(payload{"dept_id": did}).SetResult(&res).Get(fmt.Sprintf(ApiDepartmentList, token))
	if err != nil {
		return nil, fmt.Errorf("department listsub request fail, %w", err)
	}
	if resp.StatusCode() != 200 || res.ErrCode != 0 {
		return nil, fmt.Errorf("department listsub fail, %s", res)
	}

	return castDepV1ToDepV2(res.Result), nil
}

// CorpconversationAsyncsendV2 发送工作通知消息
// 接口文档: https://developers.dingtalk.com/document/app/asynchronous-sending-of-enterprise-session-messages
func (c *Component) CorpconversationAsyncsendV2(req CorpconversationAsyncsendV2Req) (CorpconversationAsyncsendV2Res, error) {
	var res CorpconversationAsyncsendV2Res
	token, err := c.GetAccessToken()
	if err != nil {
		return res, fmt.Errorf("CorpconversationAsyncsendV2-GetAccessToken err, %w", err)
	}
	req.AgentID = int64(c.config.AgentID)
	resp, err := c.ehttp.R().SetBody(req).SetResult(&res).Post(fmt.Sprintf(CorpconversationAsyncsendV2, token))
	if err != nil {
		return res, fmt.Errorf("CorpconversationAsyncsendV2-doRequest err, %w", err)
	}
	if resp.StatusCode() != 200 || res.ErrCode != 0 {
		return res, fmt.Errorf("CorpconversationAsyncsendV2-doRequest fail, %s", res)
	}
	return res, nil
}

func castDepV1ToDepV2(depv1 []departmentV1) *Department {
	depv1Map := make(map[int][]departmentV1)
	for _, v := range depv1 {
		if _, ok := depv1Map[v.ParentID]; !ok {
			depv1Map[v.ParentID] = make([]departmentV1, 0)
		}
		depv1Map[v.ParentID] = append(depv1Map[v.ParentID], v)
	}
	depv1Root := depv1Map[0][0]
	depv2Root := &Department{
		Name:            depv1Root.Name,
		DeptId:          depv1Root.ID,
		CreateDeptGroup: depv1Root.CreateDeptGroup,
		ParentId:        depv1Root.ParentID,
		SubDeptList:     make([]Department, 0),
	}
	buildDepTree(depv2Root, depv1Map)
	return depv2Root
}

func buildDepTree(d *Department, depv1Map map[int][]departmentV1) {
	// 说明无子节点
	if _, ok := depv1Map[d.DeptId]; !ok {
		return
	}
	// 有子节点
	for _, val := range depv1Map[d.DeptId] {
		newD := Department{
			Name:        val.Name,
			DeptId:      val.ID,
			ParentId:    val.ParentID,
			SubDeptList: make([]Department, 0),
		}
		buildDepTree(&newD, depv1Map)
		d.SubDeptList = append(d.SubDeptList, newD)
	}
	return
}
