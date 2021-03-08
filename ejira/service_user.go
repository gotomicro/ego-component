package ejira

import (
	"encoding/json"
	"fmt"

	"github.com/gotomicro/ego-component/ejira/entity"
)

// GetUserInfoByUsername 获取用户信息
func (c *Component) GetUserInfoByUsername(username string) (*entity.User, error) {
	var user entity.User
	resp, err := c.ehttp.R().SetBasicAuth(c.config.Username, c.config.Password).SetResult(&user).Get(fmt.Sprintf(APIGetUserInfo, username))
	if err != nil {
		return nil, fmt.Errorf("userinfo get request fail, %w", err)
	}

	var respError entity.Error
	_ = json.Unmarshal(resp.Body(), &respError)
	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("userinfo get fail, %s", respError.LongError())
	}
	return &user, err
}
