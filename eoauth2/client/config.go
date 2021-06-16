package client

// Config oauth2配置
type Config struct {
	ClientID             string
	ClientSecret         string
	AuthURL              string
	TokenURL             string
	RedirectURL          string
	UserInfoURL          string
	OauthStateCookieName string
}

// DefaultConfig 定义默认配置
func DefaultConfig() *Config {
	return &Config{
		OauthStateCookieName: "ego_oauth_state",
	}
}
