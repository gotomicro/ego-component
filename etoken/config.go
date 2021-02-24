package etoken

type Option func(c *Container)

// PackageName ..
const PackageName = "contrib.token"

// config
type config struct {
	AccessTokenIss            string
	AccessTokenKey            string
	AccessTokenExpireInterval int64
	TokenPrefix               string
}

// DefaultConfig ...
func DefaultConfig() *config {
	return &config{
		AccessTokenIss:            "git.yitum.com/mygomod/yitea-contib",
		AccessTokenKey:            "ecologysK#xo",
		AccessTokenExpireInterval: 24 * 3600,
		TokenPrefix:               "/yitea",
	}
}
