package server

// AuthorizeRequestOption 可选项
type AuthorizeRequestOption func(ar *AuthorizeRequest)

// WithAuthorizeRequestUserData 设置authorize的user data信息
func WithAuthorizeRequestUserData(userData interface{}) AuthorizeRequestOption {
	return func(c *AuthorizeRequest) {
		c.userData = userData
	}
}

// WithAuthorizeRequestAuthorized 设置authorize的flag信息
func WithAuthorizeRequestAuthorized(flag bool) AuthorizeRequestOption {
	return func(c *AuthorizeRequest) {
		c.authorized = flag
	}
}
