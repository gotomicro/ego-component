package miniprogram

import (
	"encoding/json"
	"fmt"

	"github.com/gotomicro/ego-component/ewechat/util"
)

const (
	code2SessionURL = "https://api.weixin.qq.com/sns/jscode2session?appid=%s&secret=%s&js_code=%s&grant_type=authorization_code"
)

// ResCode2Session 登录凭证校验的返回结果
type ResCode2Session struct {
	util.CommonError

	OpenID     string `json:"openid"`      // 用户唯一标识
	SessionKey string `json:"session_key"` // 会话密钥
	UnionID    string `json:"unionid"`     // 用户在开放平台的唯一标识符，在满足UnionID下发条件的情况下会返回
}

// Code2Session 登录凭证校验
func (wxa *MiniProgram) Code2Session(jsCode string) (result ResCode2Session, err error) {
	urlStr := fmt.Sprintf(code2SessionURL, wxa.AppID, wxa.AppSecret, jsCode)
	var response []byte
	response, err = wxa.Context.HTTPGet(urlStr)
	if err != nil {
		return
	}
	err = json.Unmarshal(response, &result)
	if err != nil {
		return
	}
	if result.ErrCode != 0 {
		err = fmt.Errorf("Code2Session error : errcode=%v , errmsg=%v", result.ErrCode, result.ErrMsg)
		return
	}
	return
}

// WexLogin 微信小程序登录 直接登录获取用户信息
func (m *MiniProgram) Login(code, encryptedData, iv string) (sessionKey string, wxUserInfo *UserInfo, err error) {
	wXBizDataCrypt, err := m.Code2Session(code)
	if err != nil {
		return "", nil, err
	}
	wxUserInfo, err = m.Decrypt(wXBizDataCrypt.SessionKey, encryptedData, iv)
	if err == nil {
		// 在新版本的微信API里面，将无法从encrypted里面解码出来这两个，而是只能从前面步骤里面拿到
		wxUserInfo.OpenID = wXBizDataCrypt.OpenID
		wxUserInfo.UnionID = wXBizDataCrypt.UnionID
	}
	sessionKey = wXBizDataCrypt.SessionKey
	return
}
