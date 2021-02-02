package edingtalk

import (
	"fmt"
)

type payload map[string]interface{}

type AccessTokenResponse struct {
	OpenAPIResponse
	AccessToken string `json:"access_token"`
	ExpireTime  int64  `json:"expires_in"`
	CreateTime  int64
}

type OpenAPIResponse struct {
	ErrCode   int    `json:"errcode"`
	ErrMsg    string `json:"errmsg"`
	RequestId string `json:"request_id"`
}

func (o OpenAPIResponse) String() string {
	return fmt.Sprintf(`{"errcode":%d,"errmsg":"%s"}`, o.ErrCode, o.ErrMsg)
}

type UserInfo struct {
	OpenAPIResponse
	UserId   string `json:"userid"`
	Name     string `json:"name"`
	DeviceId string `json:"deviceId"`
	IsSys    bool   `json:"is_sys"`
	SysLevel int    `json:"sys_level"`
}

type Oauth2UserUnionInfoResponse struct {
	OpenAPIResponse
	UserInfo Oauth2UserUnionInfo `json:"user_info"`
}

type Oauth2UserUnionInfo struct {
	Nick                 string `json:"nick"`                     // 用户在钉钉上面的昵称
	UnionId              string `json:"unionid"`                  // 用户在当前开放应用所属企业的唯一标识。
	OpenId               string `json:"openid"`                   // 用户在当前开放应用内的唯一标识。
	MainOrgAuthHighLevel bool   `json:"main_org_auth_high_level"` // 用户主企业是否达到高级认证级别
}

type Oauth2UseridInfo struct {
	OpenAPIResponse
	UserId      string `json:"userid"`      // 用户userid
	ContactType int    `json:"contactType"` // 联系人类型：	0：表示企业内部员工	1：表示企业外部联系人
}

type UserInfoDetail struct {
	OpenAPIResponse
	UserID          string `json:"userid"`
	OpenID          string `json:"openid"`
	Name            string `json:"name"`
	Tel             string
	WorkPlace       string
	Remark          string
	Mobile          string
	Email           string `json:"email"`
	OrgEmail        string
	Active          bool
	IsAdmin         bool
	IsBoos          bool
	DingID          string
	UnionID         string
	IsHide          bool
	Department      []int  `json:"department"`
	Position        string `json:"position"`
	Avatar          string `json:"avatar"`
	Jobnumber       string `json:"jobnumber"`
	IsSenior        bool
	StateCode       string
	OrderInDepts    string
	IsLeaderInDepts string
	Extattr         interface{}
}
