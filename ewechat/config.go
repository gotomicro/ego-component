package wechat

type Option func(c *Container)

// ModName ..
const ModName = "contrib.wechat"

// Config
type Config struct {
	Debug          bool `toml:"debug"`
	AppID          string
	AppSecret      string
	Token          string
	EncodingAESKey string
	PayMchID       string
	PayKey         string
	PayNotifyURL   string
}

// DefaultConfig ...
func DefaultConfig() *Config {
	return &Config{
	}
}