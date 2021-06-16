package client

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/gotomicro/ego/core/elog"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
)

type Component struct {
	Client *oauth2.Config
	Config *Config
	logger *elog.Component
	name   string
}

func newComponent(name string, config *Config, logger *elog.Component) *Component {
	client := &oauth2.Config{
		ClientID:     config.ClientID,
		ClientSecret: config.ClientSecret,
		Endpoint: oauth2.Endpoint{
			AuthURL:  config.AuthURL,
			TokenURL: config.TokenURL,
		},
		RedirectURL: config.RedirectURL,
	}

	return &Component{
		name:   name,
		logger: logger,
		Config: config,
		Client: client,
	}
}

// OauthState base 64 编码 referer和state信息
type OauthState struct {
	State   string `json:"state"`
	Referer string `json:"referer"`
}

type OauthLoginParams struct {
	Referer string
}

// OauthLogin 登录Handler
func (c *Component) OauthLogin(w http.ResponseWriter, r *http.Request, req OauthLoginParams) {
	// todo 安全验证来源

	// 安全验证，生成随机state，防止获取服务端系统url，登录客户端
	state, err := genRandState()
	if err != nil {
		return
	}

	oauthState := OauthState{
		State:   state,
		Referer: req.Referer,
	}
	oauthStateStr, err := json.Marshal(oauthState)
	if err != nil {
		return
	}
	sEnc := base64.RawURLEncoding.EncodeToString(oauthStateStr)

	hashedState := c.hashStateCode(state, c.Config.ClientSecret)
	// 最大300s
	http.SetCookie(w, &http.Cookie{
		Name:     c.Config.OauthStateCookieName,
		Value:    url.QueryEscape(hashedState),
		MaxAge:   300,
		Path:     "/",
		Domain:   "",
		Secure:   false,
		HttpOnly: true,
	})
	http.Redirect(w, r, c.Client.AuthCodeURL(sEnc, oauth2.AccessTypeOnline), http.StatusFound)
	return
}

// OauthCode 获取code Handler
func (c *Component) OauthCode(w http.ResponseWriter, r *http.Request) (*OauthToken, error) {
	ot := &OauthToken{
		Client: c.Client,
		config: c.Config,
	}

	code := r.FormValue("code")
	if code == "" {
		return ot, fmt.Errorf("code is empty")
	}

	stateBase64 := r.FormValue("state")
	resBytes, err := base64.RawURLEncoding.DecodeString(stateBase64)
	if err != nil {
		return ot, fmt.Errorf("state decode error,err: %w", err)
	}
	oauthState := OauthState{}
	err = json.Unmarshal(resBytes, &oauthState)
	if err != nil {
		return ot, fmt.Errorf("state decode error,err: %w", err)
	}

	cookie, err := r.Cookie(c.Config.OauthStateCookieName)
	if err != nil {
		return ot, fmt.Errorf("get cookie error,err: %w", err)
	}
	cookieState, err := url.QueryUnescape(cookie.Value)
	if err != nil {
		return ot, fmt.Errorf("get cookie query unescape error,err: %w", err)
	}

	// delete cookie
	http.SetCookie(w, &http.Cookie{
		Name:     c.Config.OauthStateCookieName,
		Value:    "",
		MaxAge:   -1,
		Path:     "/",
		Domain:   "",
		Secure:   false,
		HttpOnly: true,
	})

	if cookieState == "" {
		return ot, fmt.Errorf("cookie state empty")

	}

	queryState := c.hashStateCode(oauthState.State, c.Config.ClientSecret)
	if cookieState != queryState {
		return ot, fmt.Errorf("state not equal")
	}

	jr, err := c.Client.Exchange(r.Context(), code)
	if err != nil {
		return ot, fmt.Errorf("code exchange error, err: %w", err)
	}
	ot.Token = jr
	return ot, nil
}

func genRandState() (string, error) {
	rnd := make([]byte, 32)
	if _, err := rand.Read(rnd); err != nil {
		elog.Error("failed to generate state string", zap.Error(err))
		return "", err
	}
	return base64.URLEncoding.EncodeToString(rnd), nil
}

func (c *Component) hashStateCode(code, seed string) string {
	hashBytes := sha256.Sum256([]byte(code + c.Config.ClientID + seed))
	return hex.EncodeToString(hashBytes[:])
}
