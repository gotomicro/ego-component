package ali

import (
	"io"
	"time"

	"github.com/gotomicro/ego/core/econf"
	"github.com/gotomicro/ego/core/elog"
	"go.uber.org/zap/zapcore"
)

const (
	defaultAliFallbackCorePath = "ali.log"
)

func init() {
	elog.Register(&aliWriterBuilder{})
}

var _ elog.WriterBuilder = &aliWriterBuilder{}

type aliWriterBuilder struct {
	zapcore.Core
	io.Closer
}

// Config ...
type Config struct {
	FlushBufferSize           int           // 缓冲大小，默认256 * 1024B
	FlushBufferInterval       time.Duration // 缓冲时间，默认5秒
	AliAccessKeyID            string        // [aliWriter]阿里云sls AKID，必填
	AliAccessKeySecret        string        // [aliWriter]阿里云sls AKSecret，必填
	AliEndpoint               string        // [aliWriter]阿里云sls endpoint，必填
	AliProject                string        // [aliWriter]阿里云sls Project名称，必填
	AliLogstore               string        // [aliWriter]阿里云sls logstore名称，必填
	AliMaxQueueSize           int           // [aliWriter]阿里云sls单实例logs等待队列最大值，默认4096
	AliAPIBulkSize            int           // [aliWriter]阿里云sls API单次请求发送最大日志条数，最少256条，默认256条
	AliAPITimeout             time.Duration // [aliWriter]阿里云sls API接口超时，默认3秒
	AliAPIRetryCount          int           // [aliWriter]阿里云sls API接口重试次数，默认3次
	AliAPIRetryWaitTime       time.Duration // [aliWriter]阿里云sls API接口重试默认等待间隔，默认1秒
	AliAPIRetryMaxWaitTime    time.Duration // [aliWriter]阿里云sls API接口重试最大等待间隔，默认3秒
	AliAPIMaxIdleConnsPerHost int           // [aliWriter]阿里云sls 单个Host HTTP最大空闲连接数，应当大于AliApiMaxIdleConns
	AliAPIMaxIdleConns        int           // [aliWriter]阿里云sls HTTP最大空闲连接数
	AliAPIIdleConnTimeout     time.Duration // [aliWriter]阿里云sls HTTP空闲连接保活时间
}

func defaultConfig() *Config {
	return &Config{
		FlushBufferSize:           256 * 1024,
		FlushBufferInterval:       5 * time.Second,
		AliMaxQueueSize:           4096,
		AliAPIBulkSize:            256,
		AliAPITimeout:             3 * time.Second,
		AliAPIRetryCount:          3,
		AliAPIRetryWaitTime:       1 * time.Second,
		AliAPIRetryMaxWaitTime:    3 * time.Second,
		AliAPIMaxIdleConnsPerHost: 20,
		AliAPIMaxIdleConns:        25,
		AliAPIIdleConnTimeout:     30 * time.Second,
	}
}

// Build constructs a zapcore.Core with stderr syncer
func (*aliWriterBuilder) Build(key string, commonConfig *elog.Config) elog.Writer {
	c := defaultConfig()
	if err := econf.UnmarshalKey(key, &c); err != nil {
		panic(err)
	}

	commonConfig.Name = defaultAliFallbackCorePath
	fallbackCore := elog.Provider("file").Build(key, commonConfig)
	core, cf := NewCore(
		WithEncoder(newMapObjEncoder(*commonConfig.EncoderConfig())),
		WithEndpoint(c.AliEndpoint),
		WithAccessKeyID(c.AliAccessKeyID),
		WithAccessKeySecret(c.AliAccessKeySecret),
		WithProject(c.AliProject),
		WithLogstore(c.AliLogstore),
		WithMaxQueueSize(c.AliMaxQueueSize),
		WithLevelEnabler(commonConfig.AtomicLevel()),
		WithFlushBufferSize(c.FlushBufferSize),
		WithFlushBufferInterval(c.FlushBufferInterval),
		WithAPIBulkSize(c.AliAPIBulkSize),
		WithAPITimeout(c.AliAPITimeout),
		WithAPIRetryCount(c.AliAPIRetryCount),
		WithAPIRetryWaitTime(c.AliAPIRetryWaitTime),
		WithAPIRetryMaxWaitTime(c.AliAPIRetryMaxWaitTime),
		WithAPIMaxIdleConns(c.AliAPIMaxIdleConns),
		WithAPIIdleConnTimeout(c.AliAPIIdleConnTimeout),
		WithAPIMaxIdleConnsPerHost(c.AliAPIMaxIdleConnsPerHost),
		WithFallbackCore(fallbackCore),
	)
	return &aliWriterBuilder{
		Core:   core,
		Closer: elog.CloseFunc(cf),
	}
}

func (*aliWriterBuilder) Scheme() string {
	return "ali"
}
