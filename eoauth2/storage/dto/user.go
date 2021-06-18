package dto

import (
	"encoding/json"
	"fmt"
)

// User 用户信息
type User struct {
	// 用户uid
	Uid int64 `json:"uid"`
	// 用户昵称，中文名
	Nickname string `json:"nickname"`
	// 用户名，拼音
	Username string `json:"username"`
	// 头像
	Avatar string `json:"avatar"`
	// 邮箱
	Email string `json:"email"`
}

func (u *User) Marshal() (string, error) {
	if u == nil {
		return "", fmt.Errorf("user is nil")
	}

	bytes, err := json.Marshal(u)
	return string(bytes), err
}

// GithubUser Github OAuth 协议登录用户结构
type GithubUser struct {
	Id        int    `json:"id,omitempty"`
	Login     string `json:"login,omitempty"`
	Name      string `json:"name,omitempty"`
	AvatarURL string `json:"avatar_url,omitempty"`
	HTMLURL   string `json:"html_url,omitempty"`
	Type      string `json:"type,omitempty"` // "Type" must be "user", "team", oder "org"
}

func (u *User) ToGithubUser() GithubUser {
	return GithubUser{
		Id:        int(u.Uid),
		Login:     u.Username,
		Name:      u.Nickname,
		AvatarURL: u.Avatar,
		Type:      "user",
	}
}

// GithubTeam Github OAuth 协议登录Team结构
type GithubTeam struct {
	Id           int                    `json:"id,omitempty"`
	Organization map[string]interface{} `json:"organization,omitempty"`
	Name         string                 `json:"name,omitempty"`
	Slug         string                 `json:"slug,omitempty"`
}
