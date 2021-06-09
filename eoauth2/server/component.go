package server

import (
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/gotomicro/ego/core/elog"
)

// Component ...
type Component struct {
	name   string
	config *Config
	logger *elog.Component
}

func newComponent(name string, config *Config, logger *elog.Component) *Component {
	cron := &Component{
		config: config,
		name:   name,
		logger: logger,
	}
	return cron
}

type AuthorizeRequestParam struct {
	ClientId            string
	RedirectUri         string
	Scope               string
	State               string
	ResponseType        string
	CodeChallenge       string
	CodeChallengeMethod string
}

// HandleAuthorizeRequest is the main http.HandlerFunc for handling
// authorizaion requests
func (c *Component) HandleAuthorizeRequest(param AuthorizeRequestParam) *AuthorizeRequest {
	ret := &AuthorizeRequest{
		State: param.State,
		Scope: param.Scope,
		Context: &Context{
			logger: c.logger,
			output: make(ResponseData),
		},
		storage:           c.config.storage,
		accessTokenGen:    c.config.accessTokenGen,
		authorizeTokenGen: c.config.authorizeTokenGen,
		config:            c.config,
	}

	ret.Context.SetOutput("state", param.State)

	// create the authorization request
	unescapedUri, err := url.QueryUnescape(param.RedirectUri)
	if err != nil {
		ret.SetError(E_INVALID_REQUEST, err, "")
		return ret
	}

	ret.RedirectUri = unescapedUri

	// must have a valid client
	ret.Client, err = ret.storage.GetClient(param.ClientId)
	if err == ErrNotFound {
		ret.SetError(E_UNAUTHORIZED_CLIENT, nil, "")
		return ret
	}
	if err != nil {
		ret.SetError(E_SERVER_ERROR, err, ret.State)
		return ret
	}
	if ret.Client == nil {
		ret.SetError(E_UNAUTHORIZED_CLIENT, nil, "")
		return ret
	}
	if ret.Client.GetRedirectUri() == "" {
		ret.SetError(E_UNAUTHORIZED_CLIENT, nil, "")
		return ret
	}

	// check redirect uri, if there are multiple client redirect uri's
	// don't set the uri
	if ret.RedirectUri == "" && FirstUri(ret.Client.GetRedirectUri(), c.config.RedirectUriSeparator) == ret.Client.GetRedirectUri() {
		ret.RedirectUri = FirstUri(ret.Client.GetRedirectUri(), c.config.RedirectUriSeparator)
	}

	if realRedirectUri, err := ValidateUriList(ret.Client.GetRedirectUri(), ret.RedirectUri, c.config.RedirectUriSeparator); err != nil {
		ret.SetError(E_INVALID_REQUEST, err, ret.State)
		return ret
	} else {
		ret.RedirectUri = realRedirectUri
	}

	requestType := AuthorizeRequestType(param.ResponseType)
	// 如果不存在该类型，直接返回错误
	if !c.config.AllowedAuthorizeTypes.Exists(requestType) {
		ret.SetError(E_UNSUPPORTED_RESPONSE_TYPE, nil, ret.State)
		return ret
	}

	switch requestType {
	case CODE:
		ret.Type = CODE
		ret.Expiration = c.config.AuthorizationExpiration
		codeChallenge := param.CodeChallenge
		if len(codeChallenge) != 0 {
			codeChallengeMethod := param.CodeChallengeMethod
			// allowed values are "plain" (default) and "S256", per https://tools.ietf.org/html/rfc7636#section-4.3
			if len(codeChallengeMethod) == 0 {
				codeChallengeMethod = PKCE_PLAIN
			}
			if codeChallengeMethod != PKCE_PLAIN && codeChallengeMethod != PKCE_S256 {
				// https://tools.ietf.org/html/rfc7636#section-4.4.1
				ret.SetError(E_INVALID_REQUEST, fmt.Errorf("code_challenge_method transform algorithm not supported (rfc7636)"), "")
				return ret
			}

			// https://tools.ietf.org/html/rfc7636#section-4.2
			if matched := pkceMatcher.MatchString(codeChallenge); !matched {
				ret.SetError(E_INVALID_REQUEST, fmt.Errorf("code_challenge invalid (rfc7636)"), ret.State)
				return ret
			}

			ret.CodeChallenge = codeChallenge
			ret.CodeChallengeMethod = codeChallengeMethod
			return ret
		}

		// Optional PKCE support (https://tools.ietf.org/html/rfc7636)
		if c.config.RequirePKCEForPublicClients && CheckClientSecret(ret.Client, "") {
			// https://tools.ietf.org/html/rfc7636#section-4.4.1
			ret.SetError(E_INVALID_REQUEST, fmt.Errorf("code_challenge (rfc7636) required for public clients"), ret.State)
			return ret
		}
	case TOKEN:
		ret.Type = TOKEN
		ret.Expiration = c.config.AccessExpiration
	}
	return ret

}

type ParamAccessRequest struct {
	Method    string
	GrantType string
	AccessRequestParam
}

// HandleAccessRequest is the http.HandlerFunc for handling access token requests
func (c *Component) HandleAccessRequest(param ParamAccessRequest) *AccessRequest {
	ret := &AccessRequest{
		Context: &Context{
			logger: c.logger,
			output: make(ResponseData),
		},
		config: c.config,
	}
	// Only allow GET or POST
	if param.Method == "GET" {
		if !c.config.AllowGetAccessRequest {
			ret.SetError(E_INVALID_REQUEST, errors.New("Request must be POST"), "access_request=%s", "GET request not allowed")
			return ret
		}
	} else if param.Method != "POST" {
		ret.SetError(E_INVALID_REQUEST, errors.New("Request must be POST"), "access_request=%s", "request must be POST")
		return ret
	}

	grantType := AccessRequestType(param.GrantType)
	if !c.config.AllowedAccessTypes.Exists(grantType) {
		ret.SetError(E_UNSUPPORTED_GRANT_TYPE, nil, "access_request=%s", "unknown grant type")
		return ret
	}
	switch grantType {
	case AUTHORIZATION_CODE:
		return ret.handleAuthorizationCodeRequest(param.AccessRequestParam)
		// todo
		//case REFRESH_TOKEN:
		//	return s.handleRefreshTokenRequest(w, r)
		//case PASSWORD:
		//	return s.handlePasswordRequest(w, r)
		//case CLIENT_CREDENTIALS:
		//	return s.handleClientCredentialsRequest(w, r)
		//case ASSERTION:
		//	return s.handleAssertionRequest(w, r)
	}
	return ret
}

func (ar *AuthorizeRequest) FinishAuthorizeRequest() {
	// don't process if is already an error
	if ar.IsError() {
		return
	}

	// 设置跳转地址
	ar.SetRedirect(ar.RedirectUri)

	if !ar.authorized {
		// redirect with error
		ar.SetError(E_ACCESS_DENIED, nil, ar.State)
		return
	}

	// todo 未验证过
	if ar.Type == TOKEN {
		// generate token directly
		ret := &AccessRequest{
			Type:            IMPLICIT,
			Code:            "",
			Client:          ar.Client,
			RedirectUri:     ar.RedirectUri,
			Scope:           ar.Scope,
			GenerateRefresh: false, // per the RFC, should NOT generate a refresh token in this case
			Authorized:      true,
			Expiration:      ar.Expiration,
			UserData:        ar.userData,
			Context:         ar.Context,
			config:          ar.config,
		}
		ret.SetRedirectFragment(true)
		ret.FinishAccessRequest()
		return
	}

	// 已验证过
	// generate authorization token
	ret := &AuthorizeData{
		Client:      ar.Client,
		CreatedAt:   time.Now(),
		ExpiresIn:   ar.Expiration,
		RedirectUri: ar.RedirectUri,
		State:       ar.State,
		Scope:       ar.Scope,
		UserData:    ar.userData,
		// Optional PKCE challenge
		CodeChallenge:       ar.CodeChallenge,
		CodeChallengeMethod: ar.CodeChallengeMethod,
		Context:             ar.Context,
		storage:             ar.storage,
		authorizeTokenGen:   ar.authorizeTokenGen,
	}

	// generate token code
	code, err := ret.authorizeTokenGen.GenerateAuthorizeToken(ret)
	if err != nil {
		ret.SetError(E_SERVER_ERROR, err, ar.State)
		return
	}
	ret.Code = code

	// save authorization token
	if err = ret.storage.SaveAuthorize(ret); err != nil {
		ret.SetError(E_SERVER_ERROR, err, ar.State)
		return
	}

	// redirect with code
	ar.SetOutput("code", ret.Code)
	ar.SetOutput("state", ret.State)
	return
}
