package server

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"
)

// AccessRequestType is the type for OAuth param `grant_type`
type AccessRequestType string

const (
	AUTHORIZATION_CODE AccessRequestType = "authorization_code"
	REFRESH_TOKEN      AccessRequestType = "refresh_token"
	PASSWORD           AccessRequestType = "password"
	CLIENT_CREDENTIALS AccessRequestType = "client_credentials"
	ASSERTION          AccessRequestType = "assertion"
	IMPLICIT           AccessRequestType = "__implicit"
)

// AccessRequest is a request for access tokens
type AccessRequest struct {
	Type          AccessRequestType
	Code          string
	Client        Client
	AuthorizeData *AuthorizeData
	AccessData    *AccessData

	// Force finish to use this access data, to allow access data reuse
	ForceAccessData *AccessData
	RedirectUri     string
	Scope           string
	Username        string
	Password        string
	AssertionType   string
	Assertion       string
	authorized      bool        // Set if request is authorized
	userData        interface{} // Data to be passed to storage. Not used by the library.
	Expiration      int32       // Token expiration in seconds. Change if different from default

	// Set if a refresh token should be generated
	GenerateRefresh bool

	// Optional code_verifier as described in rfc7636
	CodeVerifier string
	*Context
	config *Config
}

// ResponseData for response output
type ResponseData map[string]interface{}

type AccessRequestParam struct {
	Code         string
	Scope        string
	CodeVerifier string
	RedirectUri  string
	ClientAuthParam
}

func (ar *AccessRequest) handleAuthorizationCodeRequest(ctx context.Context, param AccessRequestParam) *AccessRequest {
	// get client authentication
	auth := ar.getClientAuth(param.ClientAuthParam, ar.config.AllowClientSecretInParams)
	if auth == nil {
		ar.setError(E_INVALID_GRANT, nil, "getClientAuth_request=%s", "getClientAuth is required")
		return ar
	}

	// generate access token
	ar.Type = AUTHORIZATION_CODE
	ar.Code = param.Code
	ar.CodeVerifier = param.CodeVerifier
	ar.RedirectUri = param.RedirectUri
	ar.GenerateRefresh = true
	ar.Expiration = ar.config.AccessExpiration

	// "code" is required
	if ar.Code == "" {
		ar.setError(E_INVALID_GRANT, nil, "auth_code_request=%s", "code is required")
		return ar
	}

	// must have a valid client
	if ar.Client = ar.getClient(ctx, auth); ar.Client == nil {
		ar.setError(E_UNAUTHORIZED_CLIENT, nil, "auth_code_request=%s", "client is nil")
		return ar
	}

	// must be a valid authorization code
	var err error
	ar.AuthorizeData, err = ar.config.storage.LoadAuthorize(ctx, ar.Code)
	if err != nil {
		ar.setError(E_INVALID_GRANT, err, "auth_code_request=%s", "error loading authorize data")
		return ar
	}
	if ar.AuthorizeData == nil {
		ar.setError(E_UNAUTHORIZED_CLIENT, nil, "auth_code_request=%s", "authorization data is nil")
		return ar
	}
	if ar.AuthorizeData.Client == nil {
		ar.setError(E_UNAUTHORIZED_CLIENT, nil, "auth_code_request=%s", "authorization client is nil")
		return ar
	}
	if ar.AuthorizeData.Client.GetRedirectUri() == "" {
		ar.setError(E_UNAUTHORIZED_CLIENT, nil, "auth_code_request=%s", "client redirect uri is empty")
		return ar
	}
	if ar.AuthorizeData.IsExpiredAt(time.Now()) {
		ar.setError(E_INVALID_GRANT, nil, "auth_code_request=%s", "authorization data is expired")
		return ar
	}

	// code must be from the client
	if ar.AuthorizeData.Client.GetId() != ar.Client.GetId() {
		ar.setError(E_INVALID_GRANT, nil, "auth_code_request=%s", "client code does not match")
		return ar
	}

	// check redirect uri
	if ar.RedirectUri == "" {
		ar.RedirectUri = FirstUri(ar.Client.GetRedirectUri(), ar.config.RedirectUriSeparator)
	}
	if realRedirectUri, err := ValidateUriList(ar.Client.GetRedirectUri(), ar.RedirectUri, ar.config.RedirectUriSeparator); err != nil {
		ar.setError(E_INVALID_REQUEST, err, "auth_code_request=%s", "error validating client redirect")
		return ar
	} else {
		ar.RedirectUri = realRedirectUri
	}
	if ar.AuthorizeData.RedirectUri != ar.RedirectUri {
		ar.setError(E_INVALID_REQUEST, errors.New("Redirect uri is different"), "auth_code_request=%s", "client redirect does not match authorization data")
		return ar
	}

	// Verify PKCE, if present in the authorization data
	if len(ar.AuthorizeData.CodeChallenge) > 0 {
		// https://tools.ietf.org/html/rfc7636#section-4.1
		if matched := pkceMatcher.MatchString(ar.CodeVerifier); !matched {
			ar.setError(E_INVALID_REQUEST, errors.New("code_verifier has invalid format"),
				"auth_code_request=%s", "pkce code challenge verifier does not match")
			return ar
		}

		// https: //tools.ietf.org/html/rfc7636#section-4.6
		codeVerifier := ""
		switch ar.AuthorizeData.CodeChallengeMethod {
		case "", PKCE_PLAIN:
			codeVerifier = ar.CodeVerifier
		case PKCE_S256:
			hash := sha256.Sum256([]byte(ar.CodeVerifier))
			codeVerifier = base64.RawURLEncoding.EncodeToString(hash[:])
		default:
			ar.setError(E_INVALID_REQUEST, nil,
				"auth_code_request=%s", "pkce transform algorithm not supported (rfc7636)")
			return ar
		}
		if codeVerifier != ar.AuthorizeData.CodeChallenge {
			ar.setError(E_INVALID_GRANT, errors.New("code_verifier failed comparison with code_challenge"),
				"auth_code_request=%s", "pkce code verifier does not match challenge")
			return ar
		}
	}

	// set rest of data
	ar.Scope = ar.AuthorizeData.Scope
	ar.userData = ar.AuthorizeData.UserData
	return ar
}

func (ar *AccessRequest) handleRefreshTokenRequest(ctx context.Context, param AccessRequestParam) *AccessRequest {
	// get client authentication
	auth := ar.getClientAuth(param.ClientAuthParam, ar.config.AllowClientSecretInParams)
	if auth == nil {
		return nil
	}

	// generate access token
	ar.Type = REFRESH_TOKEN
	ar.Code = param.Code
	ar.Scope = param.Scope
	ar.GenerateRefresh = true
	ar.Expiration = ar.config.AccessExpiration

	// "refresh_token" is required
	if ar.Code == "" {
		ar.setError(E_INVALID_GRANT, nil, "refresh_token=%s", "refresh_token is required")
		return ar
	}

	// must have a valid client
	if ar.Client = ar.getClient(ctx, auth); ar.Client == nil {
		ar.setError(E_UNAUTHORIZED_CLIENT, nil, "auth_code_request=%s", "client is nil")
		return ar
	}

	// must be a valid refresh code
	var err error
	ar.AccessData, err = ar.config.storage.LoadRefresh(ctx, ar.Code)
	if err != nil {
		ar.setError(E_INVALID_GRANT, err, "refresh_token=%s", "error loading access data")
		return ar
	}
	if ar.AccessData == nil {
		ar.setError(E_UNAUTHORIZED_CLIENT, nil, "refresh_token=%s", "access data is nil")
		return ar
	}
	if ar.AccessData.Client == nil {
		ar.setError(E_UNAUTHORIZED_CLIENT, nil, "refresh_token=%s", "access data client is nil")
		return ar
	}
	if ar.AccessData.Client.GetRedirectUri() == "" {
		ar.setError(E_UNAUTHORIZED_CLIENT, nil, "refresh_token=%s", "access data client redirect uri is empty")
		return ar
	}

	// client must be the same as the previous token
	if ar.AccessData.Client.GetId() != ar.Client.GetId() {
		ar.setError(E_INVALID_CLIENT, errors.New("Client id must be the same from previous token"), "refresh_token=%s, current=%v, previous=%v", "client mismatch", ar.Client.GetId(), ar.AccessData.Client.GetId())
		return nil

	}

	// set rest of data
	ar.RedirectUri = ar.AccessData.RedirectUri
	ar.userData = ar.AccessData.UserData
	if ar.Scope == "" {
		ar.Scope = ar.AccessData.Scope
	}

	if extraScopes(ar.AccessData.Scope, ar.Scope) {
		msg := "the requested scope must not include any scope not originally granted by the resource owner"
		ar.setError(E_ACCESS_DENIED, errors.New(msg), "refresh_token=%s", msg)
		return ar
	}

	return ar
}

// Helper Functions

// getClient looks up and authenticates the basic auth using the given
// storage. Sets an error on the response if auth fails or a server error occurs.
func (ar *AccessRequest) getClient(ctx context.Context, auth *BasicAuth) Client {
	client, err := ar.config.storage.GetClient(ctx, auth.Username)
	if err == ErrNotFound {
		ar.setError(E_UNAUTHORIZED_CLIENT, nil, "get_client=%s", "not found")
		return nil
	}
	if err != nil {
		ar.setError(E_SERVER_ERROR, err, "get_client=%s", "error finding client")
		return nil
	}
	if client == nil {
		ar.setError(E_UNAUTHORIZED_CLIENT, nil, "get_client=%s", "client is nil")
		return nil
	}

	if !CheckClientSecret(client, auth.Password) {
		ar.setError(E_UNAUTHORIZED_CLIENT, nil, "get_client=%s, client_id=%v", "client check failed", client.GetId())
		return nil
	}

	if client.GetRedirectUri() == "" {
		ar.setError(E_UNAUTHORIZED_CLIENT, nil, "get_client=%s", "client redirect uri is empty")
		return nil
	}
	return client
}

type ClientAuthParam struct {
	ClientId      string
	ClientSecret  string
	Authorization string
}

// getClientAuth checks client basic authentication in params if allowed,
// otherwise gets it from the header.
// Sets an error on the response if no auth is present or a server error occurs.
func (ar *AccessRequest) getClientAuth(param ClientAuthParam, allowQueryParams bool) *BasicAuth {
	if allowQueryParams {
		// Allow for auth without password
		if len(param.ClientSecret) > 0 {
			auth := &BasicAuth{
				Username: param.ClientId,
				Password: param.ClientSecret,
			}
			if auth.Username != "" {
				return auth
			}
		}
	}

	auth, err := CheckBasicAuth(BasicAuthParam{
		Authorization: param.Authorization,
	})
	if err != nil {
		ar.setError(E_INVALID_REQUEST, err, "get_client_auth=%s", "check auth error")
		return nil
	}
	if auth == nil {
		ar.setError(E_INVALID_REQUEST, errors.New("Client authentication not sent"), "get_client_auth=%s", "client authentication not sent")
		return nil
	}
	return auth
}

// AccessData represents an access grant (tokens, expiration, client, etc)
type AccessData struct {
	// Client information
	Client Client

	// Authorize data, for authorization code
	AuthorizeData *AuthorizeData

	// Previous access data, for refresh token
	AccessData *AccessData

	// Access token
	AccessToken string

	// Refresh Token. Can be blank
	RefreshToken string

	// Token expiration in seconds
	ExpiresIn int32

	// Requested scope
	Scope string

	// Redirect Uri from request
	RedirectUri string

	// Date created
	CreatedAt time.Time

	// Data to be passed to storage. Not used by the library.
	UserData interface{}
}

// IsExpired returns true if access expired
func (d *AccessData) IsExpired() bool {
	return d.IsExpiredAt(time.Now())
}

// IsExpiredAt returns true if access expires at time 't'
func (d *AccessData) IsExpiredAt(t time.Time) bool {
	return d.ExpireAt().Before(t)
}

// ExpireAt returns the expiration date
func (d *AccessData) ExpireAt() time.Time {
	return d.CreatedAt.Add(time.Duration(d.ExpiresIn) * time.Second)
}

// AccessTokenGen generates access tokens
type AccessTokenGen interface {
	GenerateAccessToken(data *AccessData, generaterefresh bool) (accesstoken string, refreshtoken string, err error)
}

// Build ...
func (ar *AccessRequest) Build(options ...AccessRequestOption) error {
	// don't process if is already an error
	if ar.IsError() {
		return fmt.Errorf("AccessRequest Build error1, err %w", ar.responseErr)
	}

	for _, option := range options {
		option(ar)
	}

	redirectUri := ""
	// Get redirect uri from AccessRequest if it's there (e.g., refresh token request)
	if ar.RedirectUri != "" {
		redirectUri = ar.RedirectUri
	}
	if !ar.authorized {
		ar.setError(E_ACCESS_DENIED, nil, "finish_access_request=%s", "authorization failed")
		return fmt.Errorf("Build error2, err %w", ar.responseErr)
	}
	var ret *AccessData
	var err error

	if ar.ForceAccessData == nil {
		// generate access token
		ret = &AccessData{
			Client:        ar.Client,
			AuthorizeData: ar.AuthorizeData,
			AccessData:    ar.AccessData,
			RedirectUri:   redirectUri,
			CreatedAt:     time.Now(),
			ExpiresIn:     ar.Expiration,
			UserData:      ar.userData,
			Scope:         ar.Scope,
		}

		// generate access token
		ret.AccessToken, ret.RefreshToken, err = ar.config.accessTokenGen.GenerateAccessToken(ret, ar.GenerateRefresh)
		if err != nil {
			ar.setError(E_SERVER_ERROR, err, "finish_access_request=%s", "error generating token")
			return fmt.Errorf("Build error3, err %w", ar.responseErr)
		}
	} else {
		ret = ar.ForceAccessData
	}

	// save access token
	if err = ar.config.storage.SaveAccess(ar.Ctx, ret); err != nil {
		ar.setError(E_SERVER_ERROR, err, "finish_access_request=%s", "error saving access token")
		return fmt.Errorf("Build error4, err %w", ar.responseErr)
	}

	// remove authorization token
	if ret.AuthorizeData != nil {
		ar.config.storage.RemoveAuthorize(ar.Ctx, ret.AuthorizeData.Code)
	}

	// remove previous access token
	if ret.AccessData != nil && !ar.config.RetainTokenAfterRefresh {
		if ret.AccessData.RefreshToken != "" {
			ar.config.storage.RemoveRefresh(ar.Ctx, ret.AccessData.RefreshToken)
		}
		ar.config.storage.RemoveAccess(ar.Ctx, ret.AccessData.AccessToken)
	}

	// output data
	ar.SetOutput("access_token", ret.AccessToken)
	ar.SetOutput("token_type", ar.config.TokenType)
	ar.SetOutput("expires_in", ret.ExpiresIn)
	if ret.RefreshToken != "" {
		ar.SetOutput("refresh_token", ret.RefreshToken)
	}
	if ret.Scope != "" {
		ar.SetOutput("scope", ret.Scope)
	}
	return nil
}

func extraScopes(access_scopes, refresh_scopes string) bool {
	access_scopes_list := strings.Split(access_scopes, " ")
	refresh_scopes_list := strings.Split(refresh_scopes, " ")

	access_map := make(map[string]int)

	for _, scope := range access_scopes_list {
		if scope == "" {
			continue
		}
		access_map[scope] = 1
	}

	for _, scope := range refresh_scopes_list {
		if scope == "" {
			continue
		}
		if _, ok := access_map[scope]; !ok {
			return true
		}
	}
	return false
}
