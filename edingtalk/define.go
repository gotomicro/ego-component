package edingtalk

type AccessTokenResponse struct {
	OpenAPIResponse
	AccessToken string `json:"access_token"`
	ExpireTime  int64  `json:"expires_in"`
	CreateTime  int64
}

type OpenAPIResponse struct {
	ErrCode int    `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}

type UserInfo struct {
	OpenAPIResponse
	UserId   string `json:"userid"`
	Name     string `json:"name"`
	DeviceId string `json:"deviceId"`
	IsSys    bool   `json:"is_sys"`
	SysLevel int    `json:"sys_level"`
}
