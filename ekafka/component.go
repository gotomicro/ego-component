package ekafka

import (
	"fmt"
	"sync"

	"github.com/gotomicro/ego/core/elog"
	"github.com/segmentio/kafka-go"
)

const PackageName = "component.ekafka"

// Component kafka 组件，包含Client、Producers、Consumers
type Component struct {
	config     *config
	logger     *elog.Component
	client     *Client
	consumers  map[string]*Consumer
	producers  map[string]*Producer
	clientMu   sync.Mutex
	consumerMu sync.Mutex
	producerMu sync.Mutex
}

func (cmp *Component) interceptorChain() func(oldProcess processFn) processFn {
	return InterceptorChain(cmp.config.interceptors...)
}

// Producer 返回指定名称的kafka Producer
func (cmp *Component) Producer(name string) *Producer {
	if producer, ok := cmp.producers[name]; ok {
		return producer
	}

	cmp.producerMu.Lock()
	defer cmp.producerMu.Unlock()

	if _, ok := cmp.producers[name]; !ok {
		config, ok := cmp.config.Producers[name]
		if !ok {
			cmp.logger.Panic("producer config not exists", elog.String("name", name))
		}

		// 如果未设置balancer，则使用roundRobin
		if config.Balancer == "" {
			config.Balancer = balancerRoundRobin
		}
		balancer, ok := cmp.config.balancers[config.Balancer]
		if !ok {
			panic(fmt.Sprintf(
				"producer.Balancer is not in registered balancers, %s, %v",
				config.Balancer,
				cmp.config.balancers,
			))
		}
		producer := &Producer{
			w: &kafka.Writer{
				Addr:         kafka.TCP(cmp.config.Brokers...),
				Topic:        config.Topic,
				Balancer:     balancer,
				MaxAttempts:  config.MaxAttempts,
				BatchSize:    config.BatchSize,
				BatchBytes:   config.BatchBytes,
				BatchTimeout: config.BatchTimeout,
				ReadTimeout:  config.ReadTimeout,
				WriteTimeout: config.WriteTimeout,
				RequiredAcks: config.RequiredAcks,
				Async:        config.Async,
			},
			processor: defaultProcessor,
			logMode:   cmp.config.Debug,
		}
		producer.wrapProcessor(cmp.interceptorChain())

		cmp.producers[name] = producer
	}

	return cmp.producers[name]
}

// Consumer 返回指定名称的kafka Consumer
func (cmp *Component) Consumer(name string) *Consumer {
	if consumer, ok := cmp.consumers[name]; ok {
		return consumer
	}

	cmp.consumerMu.Lock()
	defer cmp.consumerMu.Unlock()

	if _, ok := cmp.consumers[name]; !ok {
		config, ok := cmp.config.Consumers[name]
		if !ok {
			cmp.logger.Panic("consumer config not exists", elog.String("name", name))
		}

		logger := newKafkaLogger(cmp.logger)
		errorLogger := newKafkaErrorLogger(cmp.logger)
		consumer := &Consumer{
			r: kafka.NewReader(kafka.ReaderConfig{
				Brokers:                cmp.config.Brokers,
				Topic:                  config.Topic,
				GroupID:                config.GroupID,
				Partition:              config.Partition,
				MinBytes:               config.MinBytes,
				MaxBytes:               config.MaxBytes,
				WatchPartitionChanges:  config.WatchPartitionChanges,
				PartitionWatchInterval: config.PartitionWatchInterval,
				RebalanceTimeout:       config.RebalanceTimeout,
				MaxWait:                config.MaxWait,
				ReadLagInterval:        config.ReadLagInterval,
				Logger:                 logger,
				ErrorLogger:            errorLogger,
				HeartbeatInterval:      config.HeartbeatInterval,
				CommitInterval:         config.CommitInterval,
				SessionTimeout:         config.SessionTimeout,
				JoinGroupBackoff:       config.JoinGroupBackoff,
				RetentionTime:          config.RetentionTime,
				StartOffset:            config.StartOffset,
				ReadBackoffMin:         config.ReadBackoffMin,
				ReadBackoffMax:         config.ReadBackoffMax,
			}),
			processor: defaultProcessor,
			logMode:   cmp.config.Debug,
		}
		consumer.wrapProcessor(cmp.interceptorChain())

		cmp.consumers[name] = consumer
	}

	return cmp.consumers[name]
}

// Client 返回kafka Client
func (cmp *Component) Client() *Client {
	if cmp.client != nil {
		return cmp.client
	}

	cmp.clientMu.Lock()
	defer cmp.clientMu.Unlock()

	if cmp.client == nil {
		cmp.client = &Client{
			cc:        &kafka.Client{Addr: kafka.TCP(cmp.config.Brokers...), Timeout: cmp.config.Client.Timeout},
			processor: defaultProcessor,
			logMode:   cmp.config.Debug,
		}
		cmp.client.wrapProcessor(cmp.interceptorChain())
	}

	return cmp.client
}
