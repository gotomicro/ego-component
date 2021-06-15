package server

import (
	"context"
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

// HandleAuthorizeRequest for handling
func (c *Component) HandleAuthorizeRequest(ctx context.Context, param AuthorizeRequestParam) *AuthorizeRequest {
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

	if c.config.EnableAccessInterceptor {
		c.logger.Info("HandleAuthorizeRequest access", elog.FieldCtxTid(ctx), elog.FieldValueAny(param))
	}

	ret.Context.SetOutput("state", param.State)

	// create the authorization request
	unescapedUri, err := url.QueryUnescape(param.RedirectUri)
	if err != nil {
		ret.setError(E_INVALID_REQUEST, err, "")
		return ret
	}

	ret.redirectUri = unescapedUri

	// must have a valid client
	ret.Client, err = ret.storage.GetClient(ctx, param.ClientId)
	if err == ErrNotFound {
		ret.setError(E_UNAUTHORIZED_CLIENT, nil, "")
		return ret
	}
	if err != nil {
		ret.setError(E_SERVER_ERROR, err, ret.State)
		return ret
	}
	if ret.Client == nil {
		ret.setError(E_UNAUTHORIZED_CLIENT, nil, "")
		return ret
	}
	if ret.Client.GetRedirectUri() == "" {
		ret.setError(E_UNAUTHORIZED_CLIENT, nil, "")
		return ret
	}

	// check redirect uri, if there are multiple client redirect uri's
	// don't set the uri
	if ret.redirectUri == "" && FirstUri(ret.Client.GetRedirectUri(), c.config.RedirectUriSeparator) == ret.Client.GetRedirectUri() {
		ret.redirectUri = FirstUri(ret.Client.GetRedirectUri(), c.config.RedirectUriSeparator)
	}

	if realRedirectUri, err := ValidateUriList(ret.Client.GetRedirectUri(), ret.redirectUri, c.config.RedirectUriSeparator); err != nil {
		ret.setError(E_INVALID_REQUEST, err, ret.State)
		return ret
	} else {
		ret.redirectUri = realRedirectUri
	}

	requestType := AuthorizeRequestType(param.ResponseType)
	// 如果不存在该类型，直接返回错误，code、token类型
	if !c.config.AllowedAuthorizeTypes.Exists(requestType) {
		ret.setError(E_UNSUPPORTED_RESPONSE_TYPE, nil, ret.State)
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
				ret.setError(E_INVALID_REQUEST, fmt.Errorf("code_challenge_method transform algorithm not supported (rfc7636)"), "")
				return ret
			}

			// https://tools.ietf.org/html/rfc7636#section-4.2
			if matched := pkceMatcher.MatchString(codeChallenge); !matched {
				ret.setError(E_INVALID_REQUEST, fmt.Errorf("code_challenge invalid (rfc7636)"), ret.State)
				return ret
			}

			ret.CodeChallenge = codeChallenge
			ret.CodeChallengeMethod = codeChallengeMethod
			return ret
		}

		// Optional PKCE support (https://tools.ietf.org/html/rfc7636)
		if c.config.RequirePKCEForPublicClients && CheckClientSecret(ret.Client, "") {
			// https://tools.ietf.org/html/rfc7636#section-4.4.1
			ret.setError(E_INVALID_REQUEST, fmt.Errorf("code_challenge (rfc7636) required for public clients"), ret.State)
			return ret
		}
	case TOKEN:
		ret.Type = TOKEN
		ret.Expiration = c.config.AccessExpiration
	}
	return ret

}

// Build 处理authorize请求
func (r *AuthorizeRequest) Build(ctx context.Context, options ...AuthorizeRequestOption) error {
	// don't process if is already an error
	if r.IsError() {
		return fmt.Errorf("Build error1, err %w", r.responseErr)
	}

	for _, option := range options {
		option(r)
	}

	// 设置跳转地址
	r.setRedirect(r.redirectUri)

	if !r.authorized {
		// redirect with error
		r.setError(E_ACCESS_DENIED, nil, r.State)
		return fmt.Errorf("Build error2, err %w", r.responseErr)
	}

	// todo 未验证过
	if r.Type == TOKEN {
		// generate token directly
		ret := &AccessRequest{
			Type:            IMPLICIT,
			Code:            "",
			Client:          r.Client,
			RedirectUri:     r.redirectUri,
			Scope:           r.Scope,
			GenerateRefresh: false, // per the RFC, should NOT generate a refresh token in this case
			authorized:      true,
			Expiration:      r.Expiration,
			userData:        r.userData,
			Context:         r.Context,
			config:          r.config,
		}
		ret.setRedirectFragment(true)
		ret.Build(ctx)
		return nil
	}

	// 已验证过
	// generate authorization token
	ret := &AuthorizeData{
		Client:      r.Client,
		CreatedAt:   time.Now(),
		ExpiresIn:   r.Expiration,
		RedirectUri: r.redirectUri,
		State:       r.State,
		Scope:       r.Scope,
		UserData:    r.userData,
		// Optional PKCE challenge
		CodeChallenge:       r.CodeChallenge,
		CodeChallengeMethod: r.CodeChallengeMethod,
		Context:             r.Context,
		storage:             r.storage,
		authorizeTokenGen:   r.authorizeTokenGen,
	}

	// generate token code
	code, err := ret.authorizeTokenGen.GenerateAuthorizeToken(ret)
	if err != nil {
		ret.setError(E_SERVER_ERROR, err, r.State)
		return fmt.Errorf("Build error3, err %w", r.responseErr)
	}
	ret.Code = code

	// save authorization token
	if err = ret.storage.SaveAuthorize(ctx, ret); err != nil {
		ret.setError(E_SERVER_ERROR, err, r.State)
		return fmt.Errorf("Build error4, err %w", r.responseErr)
	}

	// redirect with code
	r.SetOutput("code", ret.Code)
	r.SetOutput("state", ret.State)
	return nil
}

type ParamAccessRequest struct {
	Method    string
	GrantType string
	AccessRequestParam
}

// HandleAccessRequest is the http.HandlerFunc for handling access token requests
func (c *Component) HandleAccessRequest(ctx context.Context, param ParamAccessRequest) *AccessRequest {
	ret := &AccessRequest{
		Context: &Context{
			logger: c.logger,
			output: make(ResponseData),
		},
		config: c.config,
	}

	if c.config.EnableAccessInterceptor {
		c.logger.Info("HandleAccessRequest access", elog.FieldCtxTid(ctx), elog.FieldAddr(param.ClientId))
	}

	// Only allow GET or POST
	if param.Method == "GET" {
		if !c.config.AllowGetAccessRequest {
			ret.setError(E_INVALID_REQUEST, errors.New("Request must be POST"), "access_request=%s", "GET request not allowed")
			return ret
		}
	} else if param.Method != "POST" {
		ret.setError(E_INVALID_REQUEST, errors.New("Request must be POST"), "access_request=%s", "request must be POST")
		return ret
	}

	grantType := AccessRequestType(param.GrantType)
	if !c.config.AllowedAccessTypes.Exists(grantType) {
		ret.setError(E_UNSUPPORTED_GRANT_TYPE, nil, "access_request=%s", "unknown grant type")
		return ret
	}
	switch grantType {
	case AUTHORIZATION_CODE:
		return ret.handleAuthorizationCodeRequest(ctx, param.AccessRequestParam)
	case REFRESH_TOKEN:
		return ret.handleRefreshTokenRequest(ctx, param.AccessRequestParam)
		//case PASSWORD:
		//	return s.handlePasswordRequest(w, r)
		//case CLIENT_CREDENTIALS:
		//	return s.handleClientCredentialsRequest(w, r)
		//case ASSERTION:
		//	return s.handleAssertionRequest(w, r)
	}
	return ret
}
