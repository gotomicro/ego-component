package edingtalk

const (
	Addr = "https://oapi.dingtalk.com"
	// https://ding-doc.dingtalk.com/document#/org-dev-guide/obtain-access_token
	ApiGetToken = "/gettoken?appkey=%s&appsecret=%s"
	// https://ding-doc.dingtalk.com/document#/org-dev-guide/userid
	ApiGetUserInfo = "/user/getuserinfo?access_token=%s&code=%s"

	// https://ding-doc.dingtalk.com/document#/org-dev-guide/etaarr
	ApiOauth2SnsAuthorize = "/connect/oauth2/sns_authorize?appid=%s&response_type=code&scope=snsapi_auth&state=%s&redirect_uri=%s"
	// https://ding-doc.dingtalk.com/document#/org-dev-guide/kymkv6
	ApiOauth2Qrconnect = "/connect/qrconnect?appid=%s&response_type=code&scope=snsapi_login&state=%s&redirect_uri=%s"

	// https://ding-doc.dingtalk.com/document#/org-dev-guide/obtain-the-user-information-based-on-the-sns-temporary-authorization
	ApiGetUserInfoByCode = "/sns/getuserinfo_bycode?accessKey=%s&timestamp=%s&signature=%s"
	// https://ding-doc.dingtalk.com/document#/org-dev-guide/get-Userid-By-Unionid
	ApiGetUserIdByUnionId = "/user/getUseridByUnionid?access_token=%s&unionid=%s"
	// https://ding-doc.dingtalk.com/document#/org-dev-guide/get-user-detail
	ApiGetUserDetail = "/user/get?access_token=%s&userid=%s"

	// https://ding-doc.dingtalk.com/document/app/create-a-department-v2
	ApiDepartmentGet = "/topapi/v2/department/get?access_token=%s"
	// https://ding-doc.dingtalk.com/document/app/create-a-department-v2
	ApiDepartmentCreate = "/topapi/v2/department/create?access_token=%s"
	// https://ding-doc.dingtalk.com/document/app/update-a-department-v2
	ApiDepartmentUpdate = "/topapi/v2/department/update?access_token=%s"
	// https://ding-doc.dingtalk.com/document/app/delete-a-department-v2
	ApiDepartmentDelete = "/topapi/v2/department/delete?access_token=%s"
	// https://ding-doc.dingtalk.com/document/app/delete-a-department-v2
	ApiDepartmentListsub = "/topapi/v2/department/listsub?access_token=%s"
	// https://ding-doc.dingtalk.com/document/app/delete-a-department-v2
	ApiDepartmentList = "/department/list?access_token=%s"

	// https://ding-doc.dingtalk.com/document/app/query-user-details
	ApiUserGet = "/topapi/v2/user/get?access_token=%s"
	// https://ding-doc.dingtalk.com/document/app/user-information-creation
	ApiUserCreate = "/topapi/v2/user/create?access_token=%s"
	// https://ding-doc.dingtalk.com/document/app/user-information-update
	ApiUserUpdate = "/topapi/v2/user/update?access_token=%s"
	// https://ding-doc.dingtalk.com/document/app/user-information-delete
	ApiUserDelete = "/topapi/v2/user/delete?access_token=%s"
	// https://ding-doc.dingtalk.com/document/app/query-the-list-of-department-userids
	ApiUserListID = "/topapi/user/listid?access_token=%s"
	// https://ding-doc.dingtalk.com/document/app/queries-the-complete-information-of-a-department-user
	ApiUserList = "/topapi/v2/user/list?access_token=%s"
	// https://developers.dingtalk.com/document/app/asynchronous-sending-of-enterprise-session-messages
	CorpconversationAsyncsendV2 = "/topapi/message/corpconversation/asyncsend_v2?access_token=%s"
)

const (
	// MsgTypeLink 链接消息
	MsgTypeLink = "link"
	// MsgTypeImage 图片消息
	MsgTypeImage = "image"
	// MsgTypeText 文本消息
	MsgTypeText = "text"
	// MsgTypeVoice 语音消息
	MsgTypeVoice = "voice"
	// MsgTypeFile 文件消息
	MsgTypeFile = "file"
	// MsgTypeOA oa消息
	MsgTypeOA = "oa"
	// MsgTypeMD markdown消息
	MsgTypeMD = "markdown"
	// MsgTypeCard 卡片消息
	MsgTypeCard = "action_card"
)
