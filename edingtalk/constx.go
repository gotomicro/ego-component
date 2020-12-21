package edingtalk

const (
	Addr = "https://oapi.dingtalk.com"
	// https://ding-doc.dingtalk.com/document#/org-dev-guide/obtain-access_token
	ApiGetToken = "/gettoken?appkey=%s&appsecret=%s"
	// https://ding-doc.dingtalk.com/document#/org-dev-guide/userid
	ApiGetUserInfo = "/user/getuserinfo?access_token=%s&code=%s"

	ApiOauth2Redirect = "/connect/oauth2/sns_authorize?appid=%s&response_type=code&scope=snsapi_auth&state=%s&redirect_uri=%s"

	// https://ding-doc.dingtalk.com/document#/org-dev-guide/obtain-the-user-information-based-on-the-sns-temporary-authorization
	ApiGetUserInfoByCode = "/sns/getuserinfo_bycode?accessKey=%s&timestamp=%s&signature=%s"
	// https://ding-doc.dingtalk.com/document#/org-dev-guide/get-Userid-By-Unionid
	ApiGetUserIdByUnionId = "/user/getUseridByUnionid?access_token=%s&unionid=%s"
	// https://ding-doc.dingtalk.com/document#/org-dev-guide/get-user-detail
	ApiGetUserDetail = "/user/get?access_token=%s&userid=%s"
)
