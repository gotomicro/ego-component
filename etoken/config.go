package etoken

type Option func(c *Container)

// PackageName ..
const PackageName = "component.token"

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
		AccessTokenIss:            "github.com/gotomicro/ego/etoken",
		AccessTokenKey:            "ecologysK#xo",
		AccessTokenExpireInterval: 24 * 3600,
		TokenPrefix:               "/egotoken",
	}
}
