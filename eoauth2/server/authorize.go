package server

import (
	"regexp"
	"time"
)

// AuthorizeRequestType is the type for OAuth param `response_type`
type AuthorizeRequestType string

const (
	CODE  AuthorizeRequestType = "code"
	TOKEN AuthorizeRequestType = "token"

	PKCE_PLAIN = "plain"
	PKCE_S256  = "S256"
)

var (
	pkceMatcher = regexp.MustCompile("^[a-zA-Z0-9~._-]{43,128}$")
)

// AuthorizeRequest information
type AuthorizeRequest struct {
	Type        AuthorizeRequestType
	Client      Client
	Scope       string
	State       string
	userData    interface{} // Data to be passed to storage. Not used by the library.
	authorized  bool        // Set if request is authorized
	redirectUri string

	// Token expiration in seconds. Change if different from default.
	// If type = TOKEN, this expiration will be for the ACCESS token.
	Expiration int32

	// Optional code_challenge as described in rfc7636
	CodeChallenge string
	// Optional code_challenge_method as described in rfc7636
	CodeChallengeMethod string
	*Context
	storage           Storage
	accessTokenGen    AccessTokenGen
	authorizeTokenGen AuthorizeTokenGen
	config            *Config
}

// AuthorizeData ...
type AuthorizeData struct {
	Client              Client      // Client information
	Code                string      // Authorization code
	ExpiresIn           int32       // Token expiration in seconds
	Scope               string      // Requested scope
	RedirectUri         string      // Redirect Uri from request
	State               string      // State data from request
	CreatedAt           time.Time   // Date created
	UserData            interface{} // Data to be passed to storage. Not used by the library.
	CodeChallenge       string      // Optional code_challenge as described in rfc7636
	CodeChallengeMethod string      // Optional code_challenge_method as described in rfc7636
	*Context
	storage           Storage
	authorizeTokenGen AuthorizeTokenGen
}

// IsExpired is true if authorization expired
func (d *AuthorizeData) IsExpired() bool {
	return d.IsExpiredAt(time.Now())
}

// IsExpiredAt is true if authorization expires at time 't'
func (d *AuthorizeData) IsExpiredAt(t time.Time) bool {
	return d.ExpireAt().Before(t)
}

// ExpireAt returns the expiration date
func (d *AuthorizeData) ExpireAt() time.Time {
	return d.CreatedAt.Add(time.Duration(d.ExpiresIn) * time.Second)
}

// AuthorizeTokenGen is the token generator interface
type AuthorizeTokenGen interface {
	GenerateAuthorizeToken(data *AuthorizeData) (string, error)
}
