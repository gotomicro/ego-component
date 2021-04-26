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
	config          *config
	logger          *elog.Component
	client          *Client
	consumers       map[string]*Consumer
	producers       map[string]*Producer
	consumerGroups  map[string]*ConsumerGroup
	clientOnce      sync.Once
	consumerMu      sync.RWMutex
	producerMu      sync.RWMutex
	consumerGroupMu sync.RWMutex
}

func (cmp *Component) interceptorChain() func(oldProcess processFn) processFn {
	return InterceptorChain(cmp.config.interceptors...)
}

// Producer 返回指定名称的kafka Producer
func (cmp *Component) Producer(name string) *Producer {
	cmp.producerMu.RLock()

	if producer, ok := cmp.producers[name]; ok {
		cmp.producerMu.RUnlock()
		return producer
	}

	cmp.producerMu.RUnlock()
	cmp.producerMu.Lock()

	if producer, ok := cmp.producers[name]; ok {
		cmp.producerMu.Unlock()
		return producer
	}

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
		cmp.producerMu.Unlock()
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

	cmp.producerMu.Unlock()

	return cmp.producers[name]
}

// Consumer 返回指定名称的kafka Consumer
func (cmp *Component) Consumer(name string) *Consumer {
	cmp.consumerMu.RLock()

	if consumer, ok := cmp.consumers[name]; ok {
		cmp.consumerMu.RUnlock()
		return consumer
	}

	cmp.consumerMu.RUnlock()
	cmp.consumerMu.Lock()

	if consumer, ok := cmp.consumers[name]; ok {
		cmp.consumerMu.Unlock()
		return consumer
	}

	config, ok := cmp.config.Consumers[name]
	if !ok {
		cmp.consumerMu.Unlock()
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

	cmp.consumerMu.Unlock()

	return cmp.consumers[name]
}

// ConsumerGroup 返回指定名称的 ConsumerGroup
func (cmp *Component) ConsumerGroup(name string) *ConsumerGroup {
	cmp.consumerGroupMu.RLock()

	if consumerGroup, ok := cmp.consumerGroups[name]; ok {
		cmp.consumerGroupMu.RUnlock()
		return consumerGroup
	}

	cmp.consumerGroupMu.RUnlock()
	cmp.consumerGroupMu.Lock()

	if consumerGroup, ok := cmp.consumerGroups[name]; ok {
		cmp.consumerGroupMu.Unlock()
		return consumerGroup
	}

	config, ok := cmp.config.ConsumerGroups[name]
	if !ok {
		cmp.consumerGroupMu.Unlock()
		cmp.logger.Panic("consumerGroup config not exists", elog.String("name", name))
	}
	consumerGroup, err := NewConsumerGroup(ConsumerGroupOptions{
		Logger:                 cmp.logger,
		Brokers:                cmp.config.Brokers,
		GroupID:                config.GroupID,
		Topic:                  config.Topic,
		HeartbeatInterval:      config.HeartbeatInterval,
		PartitionWatchInterval: config.PartitionWatchInterval,
		WatchPartitionChanges:  config.WatchPartitionChanges,
		SessionTimeout:         config.SessionTimeout,
		RebalanceTimeout:       config.RebalanceTimeout,
		JoinGroupBackoff:       config.JoinGroupBackoff,
		StartOffset:            config.StartOffset,
		RetentionTime:          config.RetentionTime,
		Reader: readerOptions{
			MinBytes:        config.MinBytes,
			MaxBytes:        config.MaxBytes,
			MaxWait:         config.MaxWait,
			ReadLagInterval: config.ReadLagInterval,
			CommitInterval:  config.CommitInterval,
			ReadBackoffMin:  config.ReadBackoffMin,
			ReadBackoffMax:  config.ReadBackoffMax,
		},
	})
	if err != nil {
		cmp.logger.Panic("create ConsumerGroup failed", elog.FieldErr(err))
	}
	// TODO: wrapProcessor
	cmp.consumerGroups[name] = consumerGroup

	cmp.consumerGroupMu.Unlock()

	return cmp.consumerGroups[name]
}

// Client 返回kafka Client
func (cmp *Component) Client() *Client {
	cmp.clientOnce.Do(func() {
		cmp.client = &Client{
			cc:        &kafka.Client{Addr: kafka.TCP(cmp.config.Brokers...), Timeout: cmp.config.Client.Timeout},
			processor: defaultProcessor,
			logMode:   cmp.config.Debug,
		}
		cmp.client.wrapProcessor(cmp.interceptorChain())
	})

	return cmp.client
}
