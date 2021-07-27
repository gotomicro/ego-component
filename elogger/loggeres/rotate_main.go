package loggeres

import (
	"io"
	"os"
	"time"

	"github.com/gotomicro/ego/core/elog"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/gotomicro/ego/core/econf"
)

type EsWriterBuilder struct{}

type rotateWriter struct {
	zapcore.Core
	io.Closer
}

var _ elog.WriterBuilder = &EsWriterBuilder{}

// config ...
type config struct {
	Dir                  string
	Name                 string
	MaxSize              int           // 日志输出文件最大长度，超过改值则截断，默认500M
	MaxAge               int           // 日志存储最大时间，默认最大保存天数为7天
	MaxBackup            int           // 日志存储最大数量，默认最大保存文件个数为10个
	RotateInterval       time.Duration // 日志轮转时间，默认1天
	FlushBufferSize      int           // 缓冲大小，默认256 * 1024B
	FlushBufferInterval  time.Duration // 缓冲时间，默认5秒
	EnableEs             bool          // 启动es
	Addrs                []string      // A list of Elasticsearch nodes to use.
	Username             string        // Username for HTTP Basic Authentication.
	Password             string        // Password for HTTP Basic Authentication.
	APIKey               string        // Base64-encoded token for authorization; if set, overrides username/password and service token.
	ServiceToken         string        // Service token for authorization; if set, overrides username/password.
	RetryOnStatus        []int         // List of status codes for retry. Default: 502, 503, 504.
	EnableRetry          bool          // Default: false.
	EnableRetryOnTimeout bool          // Default: false.
	MaxRetries           int           // Default: 3.
}

func defaultConfig() *config {
	return &config{
		MaxSize:              500, // 500M
		MaxAge:               7,   // 1 day
		MaxBackup:            10,  // 10 backup
		RotateInterval:       24 * time.Hour,
		FlushBufferSize:      256 * 1024,
		FlushBufferInterval:  5 * time.Second,
		Addrs:                nil,
		Username:             "",
		Password:             "",
		APIKey:               "",
		ServiceToken:         "",
		RetryOnStatus:        []int{502, 503, 504},
		EnableRetry:          false,
		EnableRetryOnTimeout: false,
		MaxRetries:           3,
		EnableEs:             true,
	}
}

const (
	writerRotateLogger = "es"
)

func (*EsWriterBuilder) Scheme() string {
	return writerRotateLogger
}

// Load constructs a zapcore.Core with stderr syncer
func (r *EsWriterBuilder) Build(key string, commonConfig *elog.Config) elog.Writer {
	c := defaultConfig()
	c.Name = commonConfig.Name
	if err := econf.UnmarshalKey(key, &c); err != nil {
		panic(err)
	}
	// NewRotateFileCore constructs a zapcore.Core with rotate file syncer
	// Debug output to console and file by default
	cf := noopCloseFunc
	var ws = zapcore.AddSync(&rLogger{
		Filename:   commonConfig.Filename(),
		MaxSize:    c.MaxSize,
		MaxAge:     c.MaxAge,
		MaxBackups: c.MaxBackup,
		LocalTime:  true,
		Compress:   false,
		Interval:   c.RotateInterval,
	})

	if commonConfig.Debug {
		ws = zap.CombineWriteSyncers(os.Stdout, ws)
	}
	if commonConfig.EnableAsync {
		ws, cf = bufferWriteSyncer(ws, c)
	}
	w := &rotateWriter{}
	w.Closer = elog.CloseFunc(cf)
	w.Core = zapcore.NewCore(
		func() zapcore.Encoder {
			if commonConfig.Debug {
				return zapcore.NewConsoleEncoder(*commonConfig.EncoderConfig())
			}
			return zapcore.NewJSONEncoder(*commonConfig.EncoderConfig())
		}(),
		ws,
		commonConfig.AtomicLevel(),
	)
	return w
}
