package ekafka

import (
	"time"

	"github.com/segmentio/kafka-go"
)

type Config struct {
	// Brokers brokers地址
	Brokers []string `json:"brokers" toml:"brokers"`
	// Debug 是否开启debug模式
	Debug bool `json:"debug" toml:"debug"`
	// Client 用于创建topic等
	Client ClientConfig `json:"client" toml:"client"`
	// Producers 多个消费者，用于生产消息
	Producers map[string]ProducerConfig `json:"producers" toml:"producers"`
	// Consumers 多个生产者，用于消费消息
	Consumers    map[string]ConsumerConfig `json:"consumers" toml:"consumers"`
	interceptors []Interceptor
	balancers    map[string]Balancer
}

type ClientConfig struct {
	// Timeout 超时时间
	Timeout time.Duration `json:"timeout" toml:"timeout"`
}

type ProducerConfig struct {
	// Topic 指定生产的消息推送到哪个topic
	Topic string `json:"topic" toml:"topic"`
	// Balancer 指定使用哪种Balancer，可选：hash\roundRobin
	Balancer string `json:"balancer" toml:"balancer"`
}

type ConsumerConfig struct {
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
	ReadLagInterval time.Duration `json:"readLagInterval" toml:"readLagInterval"`
}

const (
	balancerHash       = "hash"
	balancerRoundRobin = "roundRobin"
)

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		Debug: true,
		balancers: map[string]Balancer{
			balancerHash:       &kafka.Hash{},
			balancerRoundRobin: &kafka.RoundRobin{},
		},
	}
}
