package server

// AccessRequestOption 可选项
type AccessRequestOption func(ar *AccessRequest)

// WithAccessRequestAuthorized 设置authorized flag
func WithAccessRequestAuthorized(flag bool) AccessRequestOption {
	return func(c *AccessRequest) {
		c.authorized = flag
	}
}
