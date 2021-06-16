package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"golang.org/x/oauth2"
)

type OauthToken struct {
	*oauth2.Token
	Client *oauth2.Config
	config *Config
}

func (o *OauthToken) UserInfo(ctx context.Context, user interface{}) (err error) {
	resp, err := o.Client.Client(ctx, o.Token).Get(o.config.UserInfoURL)
	if err != nil {
		return fmt.Errorf("get client resp error, err: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("get client resp status not ok, code: %v", resp.StatusCode)
	}

	err = json.NewDecoder(resp.Body).Decode(&user)
	if err != nil {
		return fmt.Errorf("json decode error, err: %w", err)
	}
	return nil
}
