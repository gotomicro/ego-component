package esession

// config
type config struct {
	Mode     string // session模式，默认redis，目前只支持redis和memstore
	Name     string // session名称
	Size     int
	Debug    bool   // debug变量
	Network  string // 协议
	Addr     string
	Password string
	Keypairs string
}

// DefaultConfig 定义了esession默认配置
func DefaultConfig() *config {
	return &config{
		Mode: "redis",
	}
}
