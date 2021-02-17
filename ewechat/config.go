package wechat

import (
	"github.com/go-resty/resty/v2"
	"github.com/gotomicro/ego/core/econf"
	"github.com/gotomicro/ego/core/elog"

	"github.com/gotomicro/ego-component/eredis"
	"github.com/gotomicro/ego-component/ewechat/cache"
	"github.com/gotomicro/ego-component/ewechat/context"
	"github.com/gotomicro/ego-component/ewechat/miniprogram"

	"sync"
)

type Option func(c *Config)

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

	Context *context.Context
	client  cache.Cache
	logger  *elog.Component
}

// DefaultConfig ...
func DefaultConfig() *Config {
	return &Config{
		logger: elog.EgoLogger.With(elog.FieldComponent(ModName)),
	}
}

// Invoker ...
func Load(key string) *Config {
	var config = DefaultConfig()
	if err := econf.UnmarshalKey(key, &config); err != nil {
		config.logger.Panic("parse wechat config panic", elog.FieldErr(err), elog.FieldKey(key), elog.FieldValueAny(config))
	}
	config.logger = config.logger.With(elog.FieldComponentName(key))
	return config
}

// Build
func (cfg *Config) Build(options ...Option) *Config {
	ctx := new(context.Context)
	ctx.AppID = cfg.AppID
	ctx.AppSecret = cfg.AppSecret
	ctx.Token = cfg.Token
	ctx.EncodingAESKey = cfg.EncodingAESKey
	ctx.PayMchID = cfg.PayMchID
	ctx.PayKey = cfg.PayKey
	ctx.PayNotifyURL = cfg.PayNotifyURL
	cfg.Context = ctx
	ctx.SetAccessTokenLock(new(sync.RWMutex))
	ctx.SetJsAPITicketLock(new(sync.RWMutex))
	ctx.RestyClient = resty.New().SetDebug(cfg.Debug)

	for _, option := range options {
		option(cfg)
	}

	ctx.Cache = cfg.client

	return cfg
}

func WithRedis(client *eredis.Component) Option {
	return func(c *Config) {
		c.client = client
	}
}

// GetMiniProgram 获取小程序的实例
func (cfg *Config) GetMiniProgram() *miniprogram.MiniProgram {
	return miniprogram.NewMiniProgram(cfg.Context)
}
