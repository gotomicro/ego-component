package server

const PackageName = "component.eoauth2.server"

// Config contains server configuration information
type Config struct {
	EnableAccessInterceptor bool                  // 是否开启，记录请求数据
	AuthorizationExpiration int32                 // Authorization token expiration in seconds (default 5 minutes)
	AccessExpiration        int32                 // Access token expiration in seconds (default 1 hour)
	TokenType               string                // Token type to return
	AllowedAuthorizeTypes   AllowedAuthorizeTypes // List of allowed authorize types (only CODE by default)
	AllowedAccessTypes      AllowedAccessTypes    // List of allowed access types (only AUTHORIZATION_CODE by default)
	// HTTP status code to return for errors - default 200
	// Only used if response was created from server
	ErrorStatusCode int
	// If true allows client secret also in params, else only in
	// Authorization header - default false
	AllowClientSecretInParams bool
	// If true allows access request using GET, else only POST - default false
	AllowGetAccessRequest bool
	// Require PKCE for code flows for public OAuth clients - default false
	RequirePKCEForPublicClients bool
	// Separator to support multiple URIs in Client.GetRedirectUri().
	// If blank (the default), don't allow multiple URIs.
	RedirectUriSeparator string
	// RetainTokenAfter Refresh allows the server to retain the access and
	// refresh token for re-use - default false
	RetainTokenAfterRefresh bool
	storage                 Storage
	authorizeTokenGen       AuthorizeTokenGen
	accessTokenGen          AccessTokenGen
}

// DefaultConfig ...
func DefaultConfig() *Config {
	return &Config{
		AuthorizationExpiration:     300,
		AccessExpiration:            3600,
		TokenType:                   "Bearer",
		AllowedAuthorizeTypes:       AllowedAuthorizeTypes{CODE},
		AllowedAccessTypes:          AllowedAccessTypes{AUTHORIZATION_CODE},
		ErrorStatusCode:             200,
		AllowClientSecretInParams:   true,
		AllowGetAccessRequest:       false,
		RequirePKCEForPublicClients: false,
		RedirectUriSeparator:        "",
		RetainTokenAfterRefresh:     false,
		authorizeTokenGen:           &AuthorizeTokenGenDefault{},
		accessTokenGen:              &AccessTokenGenDefault{},
	}
}

// AllowedAuthorizeTypes is a collection of allowed auth request types
type AllowedAuthorizeTypes []AuthorizeRequestType

// Exists returns true if the auth type exists in the list
func (t AllowedAuthorizeTypes) Exists(rt AuthorizeRequestType) bool {
	for _, k := range t {
		if k == rt {
			return true
		}
	}
	return false
}

// AllowedAccessTypes is a collection of allowed access request types
type AllowedAccessTypes []AccessRequestType

// Exists returns true if the access type exists in the list
func (t AllowedAccessTypes) Exists(rt AccessRequestType) bool {
	for _, k := range t {
		if k == rt {
			return true
		}
	}
	return false
}
