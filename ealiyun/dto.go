package ealiyun

// SaveRamUserRequest 具体注释参考: https://help.aliyun.com/document_detail/185726.html?spm=a2c4g.11186623.6.737.207d1f2fRXubtd
type SaveRamUserRequest struct {
	UserPrincipalName string `json:"user_name"`    // RAM用户登录名。
	DisplayName       string `json:"display_name"` // RAM用户显示名称。
	MobilePhone       string `json:"mobile_phone"` // RAM用户的手机号。
	Email             string `json:"email"`        // RAM用户的邮箱。
	Comments          string `json:"comments"`     // 备注
}
type RamUserInfo struct {
	UserID            string `json:"user_id"`
	CreateDate        string `json:"create_date"`
	UpdateDate        string `json:"update_date"`
	UserPrincipalName string `json:"user_principal_name"`
	DisplayName       string `json:"display_name"`
	MobilePhone       string `json:"mobile_phone"`
	Email             string `json:"email"`
	Comments          string `json:"comments"`
	LastLoginDate     string `json:"last_login_date"`
}
type RamUserResponse struct {
	RequestID string      `json:"request_id"`
	User      RamUserInfo `json:"user"`
}

type GroupInfo struct {
	DisplayName string `json:"display_name"`
	GroupName   string `json:"group_name"`
	GroupId     string `json:"group_id"`
	Comments    string `json:"comments"`
	JoinDate    string `json:"join_date"`
	UpdateDate  string `json:"update_date"`
	CreateDate  string `json:"create_date"`
}

type AddOrRemoveUserToGroupRequest struct {
	GroupName         string `json:"group_name"`
	UserPrincipalName string `json:"user_principal_name"`
}
