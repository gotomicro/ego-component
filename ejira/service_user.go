package ejira

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
)

// AvatarUrls represents different dimensions of avatars / images
type AvatarUrls struct {
	Four8X48  string `json:"48x48,omitempty"`
	Two4X24   string `json:"24x24,omitempty"`
	One6X16   string `json:"16x16,omitempty"`
	Three2X32 string `json:"32x32,omitempty"`
}

// User represents a Jira user.
type User struct {
	Self            string     `json:"self,omitempty"`
	AccountID       string     `json:"accountId,omitempty"`
	AccountType     string     `json:"accountType,omitempty"`
	Name            string     `json:"name,omitempty"`
	Key             string     `json:"key,omitempty"`
	Password        string     `json:"-"`
	EmailAddress    string     `json:"emailAddress,omitempty"`
	AvatarUrls      AvatarUrls `json:"avatarUrls,omitempty"`
	DisplayName     string     `json:"displayName,omitempty"`
	Active          bool       `json:"active,omitempty"`
	TimeZone        string     `json:"timeZone,omitempty"`
	Locale          string     `json:"locale,omitempty"`
	ApplicationKeys []string   `json:"applicationKeys,omitempty"`
}

// UserSearchOption 查询参数
type UserSearchOption struct {
	Username        string
	StartAt         int
	MaxResults      int
	IncludeActive   *bool
	IncludeInactive *bool
}

// GetUserInfoByUsername 获取用户信息
// Jira api docs：https://docs.atlassian.com/software/jira/docs/api/REST/8.8.0/#api/2/user-getUser
func (c *Component) GetUserInfoByUsername(username string) (*User, error) {
	var user User
	resp, err := c.ehttp.R().SetBasicAuth(c.config.Username, c.config.Password).SetResult(&user).Get(fmt.Sprintf(APIGetUserInfo, username))
	if err != nil {
		return nil, fmt.Errorf("userinfo get request fail, %w", err)
	}

	var respError Error
	_ = json.Unmarshal(resp.Body(), &respError)
	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("userinfo get fail, %s", respError.LongError())
	}
	return &user, err
}

// FindUsers 查找用户
// Jira API docs: https://docs.atlassian.com/software/jira/docs/api/REST/8.8.0/#api/2/user-findUsers
func (c *Component) FindUsers(options *UserSearchOption) (*[]User, error) {
	uv := url.Values{}
	if options != nil {
		if options.Username != "" {
			uv.Add("username", options.Username)
		} else {
			uv.Add("username", ".")
		}

		if options.StartAt != 0 {
			uv.Add("startAt", strconv.Itoa(options.StartAt))
		}
		if options.MaxResults != 0 {
			uv.Add("maxResults", strconv.Itoa(options.MaxResults))
		}

		if options.IncludeActive != nil {
			if *options.IncludeActive {
				uv.Add("includeActive", "true")
			} else {
				uv.Add("includeActive", "false")
			}
		}

		if options.IncludeInactive != nil {
			if *options.IncludeInactive {
				uv.Add("includeActive", "true")
			} else {
				uv.Add("includeActive", "false")
			}
		}
	}

	var users []User
	resp, err := c.ehttp.R().SetBasicAuth(c.config.Username, c.config.Password).SetQueryParamsFromValues(uv).SetResult(&users).Get(fmt.Sprintf(APIFindUsers))
	if err != nil {
		return nil, fmt.Errorf("userlist get request fail, %w", err)
	}

	var respError Error
	_ = json.Unmarshal(resp.Body(), &respError)
	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("userlist get fail, %s", respError.LongError())
	}

	return &users, err
}
