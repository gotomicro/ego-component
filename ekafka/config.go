package ekafka

import (
	"time"

	"github.com/segmentio/kafka-go"
)

type config struct {
	// Brokers brokers地址
	Brokers []string `json:"brokers" toml:"brokers"`
	// Debug 是否开启debug模式
	Debug bool `json:"debug" toml:"debug"`
	// Client 用于创建topic等
	Client clientConfig `json:"client" toml:"client"`
	// Producers 多个消费者，用于生产消息
	Producers map[string]producerConfig `json:"producers" toml:"producers"`
	// Consumers 多个生产者，用于消费消息
	Consumers map[string]consumerConfig `json:"consumers" toml:"consumers"`
	// ConsumerGroups 多个消费组，用于消费消息
	ConsumerGroups             map[string]consumerGroupConfig `json:"consumerGroups" toml:"consumerGroups"`
	clientInterceptors         []ClientInterceptor
	serverInterceptors         []ServerInterceptor
	balancers                  map[string]Balancer
	EnableTraceInterceptor     bool // 是否开启链路追踪，默认开启
	EnableAccessInterceptor    bool // 是否开启记录请求数据，默认不开启
	EnableAccessInterceptorReq bool // 是否开启记录请求参数，默认不开启
	EnableAccessInterceptorRes bool // 是否开启记录响应参数，默认不开启
	EnableMetricInterceptor    bool // 是否开启监控，默认开启
}

type clientConfig struct {
	// Timeout 超时时间
	Timeout time.Duration `json:"timeout" toml:"timeout"`
}

type producerConfig struct {
	// Topic 指定生产的消息推送到哪个topic
	Topic string `json:"topic" toml:"topic"`
	// Balancer 指定使用哪种Balancer，可选：hash\roundRobin
	Balancer string `json:"balancer" toml:"balancer"`
	// MaxAttempts 最大重试次数，默认10次
	MaxAttempts int `json:"maxAttempts" toml:"maxAttempts"`
	// BatchSize 批量发送的消息数量，默认100条
	BatchSize int `json:"batchSize" toml:"batchSize"`
	// BatchBytes 批量发送的消息大小，默认1MB
	BatchBytes int64 `json:"batchBytes" toml:"batchBytes"`
	// BatchTimeout 批量发送消息的周期，默认1s
	BatchTimeout time.Duration `json:"batchTimeout" toml:"batchTimeout"`
	// ReadTimeout 读超时
	ReadTimeout time.Duration `json:"readTimeout" toml:"readTimeout"`
	// WriteTimeout 写超时
	WriteTimeout time.Duration `json:"writeTimeout" toml:"writeTimeout"`
	// RequiredAcks ACK配置
	// RequireNone (0) fire-and-forget，producer不等待来自broker同步完成的确认后，就可以发送下一批消息
	// RequireOne  (1) producer在leader已成功收到的数据并得到确认后，才发送下一批消息
	// RequireAll  (-1) producer在所有follower副本确认接收到数据后，才发送下一批消息
	RequiredAcks kafka.RequiredAcks `json:"requiredAcks" toml:"requiredAcks"`
	// Async 设置成true时会导致WriteMessages非阻塞，会导致调用WriteMessages方法获取不到error
	Async bool `json:"async" toml:"async"`
}

type consumerConfig struct {
	// Partition 指定分区ID，和GroupID不能同时配置
	Partition int `json:"partition" toml:"partition"`
	// GroupID 指定分组ID，和Partition不能同时配置，当配置了GroupID时，默认使用ConsumerGroup来消费
	GroupID string `json:"groupID" toml:"groupID"`
	// Topic 消费的topic
	Topic string `json:"topic" toml:"topic"`
	// MinBytes 向kafka发送请求的包最小值
	MinBytes int `json:"minBytes" toml:"minBytes"`
	// MaxBytes 向kafka发送请求的包最大值
	MaxBytes int `json:"maxBytes" toml:"maxBytes"`
	// WatchPartitionChanges 是否监听分区变化
	WatchPartitionChanges bool `json:"watchPartitionChanges" toml:"watchPartitionChanges"`
	// PartitionWatchInterval 监听分区变化时间周期
	PartitionWatchInterval time.Duration `json:"partitionWatchInterval" toml:"partitionWatchInterval"`
	// RebalanceTimeout rebalance 超时时间
	RebalanceTimeout time.Duration `json:"rebalanceTimeout" toml:"rebalanceTimeout"`
	// MaxWait 从kafka批量获取数据时，最大等待间隔
	MaxWait time.Duration `json:"maxWait" toml:"maxWait"`
	// ReadLagInterval 获取消费者滞后值的时间周期
	ReadLagInterval   time.Duration `json:"readLagInterval" toml:"readLagInterval"`
	HeartbeatInterval time.Duration `json:"heartbeatInterval" ,toml:"heartbeatInterval"`
	CommitInterval    time.Duration `json:"commitInterval" toml:"commitInterval"`
	SessionTimeout    time.Duration `json:"sessionTimeout" toml:"sessionTimeout"`
	JoinGroupBackoff  time.Duration `json:"joinGroupBackoff" toml:"joinGroupBackoff"`
	RetentionTime     time.Duration `json:"retentionTime" toml:"retentionTime"`
	StartOffset       int64         `json:"startOffset" toml:"startOffset"`
	ReadBackoffMin    time.Duration `json:"readBackoffMin" toml:"readBackoffMin"`
	ReadBackoffMax    time.Duration `json:"readBackoffMax" toml:"readBackoffMax"`
}

type consumerGroupConfig struct {
	GroupID                string                `json:"groupID" toml:"groupID"`
	Topic                  string                `json:"topic" toml:"topic"`
	GroupBalancers         []kafka.GroupBalancer `json:"groupBalancers" toml:"groupBalancers"`
	HeartbeatInterval      time.Duration         `json:"heartbeatInterval" toml:"heartbeatInterval"`
	PartitionWatchInterval time.Duration         `json:"partitionWatchInterval" toml:"partitionWatchInterval"`
	WatchPartitionChanges  bool                  `json:"watchPartitionChanges" toml:"watchPartitionChanges"`
	SessionTimeout         time.Duration         `json:"sessionTimeout" toml:"sessionTimeout"`
	RebalanceTimeout       time.Duration         `json:"rebalanceTimeout" toml:"rebalanceTimeout"`
	JoinGroupBackoff       time.Duration         `json:"joinGroupBackoff" toml:"joinGroupBackoff"`
	StartOffset            int64                 `json:"startOffset" toml:"startOffset"`
	RetentionTime          time.Duration         `json:"retentionTime" toml:"retentionTime"`
	// Reader otpions:
	MinBytes        int           `json:"minBytes" toml:"minBytes"`
	MaxBytes        int           `json:"maxBytes" toml:"maxBytes"`
	MaxWait         time.Duration `json:"maxWait" toml:"maxWait"`
	ReadLagInterval time.Duration `json:"readLagInterval" toml:"readLagInterval"`
	CommitInterval  time.Duration `json:"commitInterval" toml:"commitInterval"`
	ReadBackoffMin  time.Duration `json:"readBackoffMin" toml:"readBackoffMin"`
	ReadBackoffMax  time.Duration `json:"readBackoffMax" toml:"readBackoffMax"`
}

const (
	balancerHash       = "hash"
	balancerRoundRobin = "roundRobin"
)

// DefaultConfig 返回默认配置
func DefaultConfig() *config {
	return &config{
		Debug:                   true,
		EnableTraceInterceptor:  true,
		EnableMetricInterceptor: true,
		balancers: map[string]Balancer{
			balancerHash:       &kafka.Hash{},
			balancerRoundRobin: &kafka.RoundRobin{},
		},
	}
}
