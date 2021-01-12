package ekafka

import (
	"fmt"

	"github.com/gotomicro/ego/core/econf"
	"github.com/gotomicro/ego/core/elog"
	"github.com/segmentio/kafka-go"
)

type Option func(c *Container)

type Container struct {
	config *Config
	name   string
	logger *elog.Component
}

// DefaultContainer 返回默认Container
func DefaultContainer() *Container {
	return &Container{
		config: DefaultConfig(),
		logger: elog.EgoLogger.With(elog.FieldComponent(PackageName)),
	}
}

// Load 载入配置，初始化Container
func Load(key string) *Container {
	c := DefaultContainer()
	if err := econf.UnmarshalKey(key, &c.config); err != nil {
		c.logger.Panic("parse config error", elog.FieldErr(err), elog.FieldKey(key))
		return c
	}

	c.logger = c.logger.With(elog.FieldComponentName(key))
	c.name = key
	return c
}

// Build 构建Container
func (c *Container) Build(options ...Option) *Component {
	if options == nil {
		options = make([]Option, 0)
	}
	if c.config.Debug {
		options = append(options, WithInterceptor(debugInterceptor(c.name, c.config)))
	}
	for _, option := range options {
		option(c)
	}

	c.logger = c.logger.With(elog.FieldAddr(fmt.Sprintf("%s", c.config.Brokers)))
	cmp := &Component{
		Config: c.config,
		logger: c.logger,
	}
	// 初始化client
	cmp.client = &Client{
		cc:        &kafka.Client{Addr: kafka.TCP(c.config.Brokers...), Timeout: c.config.Client.Timeout},
		processor: defaultProcessor,
		logMode:   c.config.Debug,
	}
	ic := InterceptorChain(c.config.interceptors...)
	cmp.client.wrapProcessor(ic)

	// 初始化producers
	cmp.producers = make(map[string]*Producer)
	for name, producer := range c.config.Producers {
		w := &Producer{
			w: &kafka.Writer{
				Addr:     kafka.TCP(c.config.Brokers...),
				Topic:    producer.Topic,
				Balancer: getBalancer(producer.Balancer),
			},
			processor: defaultProcessor,
			logMode:   c.config.Debug,
		}
		w.wrapProcessor(ic)
		cmp.producers[name] = w
	}

	// 初始化consumers
	cmp.consumers = make(map[string]*Consumer)
	l := &logger{cmp.logger}
	for name, cg := range c.config.Consumers {
		r := &Consumer{
			r: kafka.NewReader(kafka.ReaderConfig{
				Brokers:                c.config.Brokers,
				Topic:                  cg.Topic,
				GroupID:                cg.GroupID,
				Partition:              cg.Partition,
				MinBytes:               cg.MinBytes,
				MaxBytes:               cg.MaxBytes,
				WatchPartitionChanges:  cg.WatchPartitionChanges,
				PartitionWatchInterval: cg.PartitionWatchInterval,
				RebalanceTimeout:       cg.RebalanceTimeout,
				MaxWait:                cg.MaxWait,
				ReadLagInterval:        cg.ReadLagInterval,
				ErrorLogger:            l,
			}),
			processor: defaultProcessor,
			logMode:   c.config.Debug,
		}
		r.wrapProcessor(ic)
		cmp.consumers[name] = r
	}

	return cmp
}

func getBalancer(name string) kafka.Balancer {
	switch name {
	case "hash":
		return &kafka.Hash{}
	case "roundRobin":
		return &kafka.RoundRobin{}
	default:
		return &kafka.RoundRobin{}
	}
}

type logger struct {
	*elog.Component
}

func (l *logger) Printf(tmpl string, args ...interface{}) {
	l.Errorf(tmpl, args...)
}
