package esession

// config
type config struct {
	Mode       string // session模式，默认redis，目前只支持redis、eredis 和 memstore
	Name       string // session名称
	Size       int
	Debug      bool   // debug变量
	Network    string // 协议
	Addr       string
	Password   string
	Keypairs   string
	RedisMode  string   // eredis 下生效 Redis模式 cluster|stub|sentinel，默认 stub
	MasterName string   // eredis sentinel  模式下用到
	Addrs      []string // eredis  用到
}

// DefaultConfig 定义了esession默认配置
func DefaultConfig() *config {
	return &config{
		Mode: "redis",
	}
}
