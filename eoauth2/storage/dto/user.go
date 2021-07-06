package dto

import (
	"encoding/json"
	"fmt"
)

// User 用户信息
type User struct {
	Uid      int64  `json:"uid"`      // 用户uid
	Nickname string `json:"nickname"` // 用户昵称，中文名
	Username string `json:"username"` // 用户名，拼音
	Avatar   string `json:"avatar"`   // 头像
	Email    string `json:"email"`    // 邮箱
	State    int    `json:"state"`    // 状态
}

func (u *User) Marshal() (string, error) {
	if u == nil {
		return "", fmt.Errorf("user is nil")
	}

	bytes, err := json.Marshal(u)
	return string(bytes), err
}

// GitlabUser Gitlab OAuth 协议登录用户结构
type GitlabUser struct {
	Id       int    `json:"Id"`
	Username string `json:"Username"`
	Email    string `json:"Email"`
	Name     string `json:"Name"`
	State    string `json:"State"`
}

func (u *User) ToGitlabUser() GitlabUser {
	activeState := "inactive"
	if u.State == 1 {
		activeState = "active"
	}

	return GitlabUser{
		Id:       int(u.Uid),
		Username: u.Username,
		Email:    u.Email,
		Name:     u.Nickname,
		State:    activeState,
	}
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
