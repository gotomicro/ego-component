package etoken

type Option func(c *Container)

// PackageName ..
const PackageName = "contrib.token"

// Config
type Config struct {
	AccessTokenIss            string
	AccessTokenKey            string
	AccessTokenExpireInterval int64
	TokenPrefix               string
}

// DefaultConfig ...
func DefaultConfig() *Config {
	return &Config{
		AccessTokenIss:            "git.yitum.com/mygomod/yitea-contib",
		AccessTokenKey:            "ecologysK#xo",
		AccessTokenExpireInterval: 24 * 3600,
		TokenPrefix:               "/yitea",
	}
}