package edingtalk

const (
	Addr = "https://oapi.dingtalk.com"
	// https://ding-doc.dingtalk.com/document#/org-dev-guide/obtain-access_token
	ApiGetToken = "/gettoken?appkey=%s&appsecret=%s"
	// https://ding-doc.dingtalk.com/document#/org-dev-guide/userid
	ApiGetUserInfo = "/user/getuserinfo?access_token=%s&code=%s"
)
