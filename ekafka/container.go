package ekafka

import (
	"fmt"

	"github.com/gotomicro/ego/core/econf"
	"github.com/gotomicro/ego/core/elog"
	"github.com/segmentio/kafka-go"
)

type Option func(c *Container)

type Container struct {
	config *config
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
	if err := econf.UnmarshalKey(key, &c.config, econf.WithWeaklyTypedInput(true)); err != nil {
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
		config: c.config,
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
		// 如果未设置balancer，则使用roundRobin
		if producer.Balancer == "" {
			producer.Balancer = balancerRoundRobin
		}
		b, ok := c.config.balancers[producer.Balancer]
		if !ok {
			panic(fmt.Sprintf("producer.Balancer is not in registered balancers, %s, %v", producer.Balancer, c.config.balancers))
		}
		w := &Producer{
			w: &kafka.Writer{
				Addr:         kafka.TCP(c.config.Brokers...),
				Topic:        producer.Topic,
				Balancer:     b,
				MaxAttempts:  producer.MaxAttempts,
				BatchSize:    producer.BatchSize,
				BatchBytes:   producer.BatchBytes,
				BatchTimeout: producer.BatchTimeout,
				ReadTimeout:  producer.ReadTimeout,
				WriteTimeout: producer.WriteTimeout,
				RequiredAcks: producer.RequiredAcks,
				Async:        producer.Async,
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
	el := &errorLogger{cmp.logger}
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
				Logger:                 l,
				ErrorLogger:            el,
				HeartbeatInterval:      cg.HeartbeatInterval,
				CommitInterval:         cg.CommitInterval,
				SessionTimeout:         cg.SessionTimeout,
				JoinGroupBackoff:       cg.JoinGroupBackoff,
				RetentionTime:          cg.RetentionTime,
				StartOffset:            cg.StartOffset,
				ReadBackoffMin:         cg.ReadBackoffMin,
				ReadBackoffMax:         cg.ReadBackoffMax,
			}),
			processor: defaultProcessor,
			logMode:   c.config.Debug,
		}
		r.wrapProcessor(ic)
		cmp.consumers[name] = r
	}

	return cmp
}

type logger struct {
	*elog.Component
}

func (l *logger) Printf(tmpl string, args ...interface{}) {
	l.Debugf(tmpl, args...)
}

type errorLogger struct {
	*elog.Component
}

func (l *errorLogger) Printf(tmpl string, args ...interface{}) {
	l.Errorf(tmpl, args...)
}
