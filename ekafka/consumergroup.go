package ekafka

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/gotomicro/ego/core/elog"
	"github.com/segmentio/kafka-go"
)

type TopicPartition struct {
	Topic     string
	Partition int
	Offset    int64
}

type AssignedPartitions struct {
	Partitions []TopicPartition
}

type RevokedPartitions struct {
	Partitions []TopicPartition
}

type ConsumerGroup struct {
	logger     *elog.Component
	group      *kafka.ConsumerGroup
	events     chan interface{}
	options    *ConsumerGroupOptions
	currentGen *kafka.Generation
	genMu      sync.RWMutex
	readerWg   sync.WaitGroup
	processor  processor
}

func createTopicPartitionsFromGenAssignments(genAssignments map[string][]kafka.PartitionAssignment) []TopicPartition {
	topicPartitions := make([]TopicPartition, 0)
	for topic, assignments := range genAssignments {
		for _, assignment := range assignments {
			topicPartitions = append(topicPartitions, TopicPartition{
				Topic:     topic,
				Partition: assignment.ID,
				Offset:    assignment.Offset,
			})
		}
	}
	return topicPartitions
}

type readerOptions struct {
	MinBytes        int
	MaxBytes        int
	MaxWait         time.Duration
	ReadLagInterval time.Duration
	CommitInterval  time.Duration
	ReadBackoffMin  time.Duration
	ReadBackoffMax  time.Duration
}

type ConsumerGroupOptions struct {
	Logger                 *elog.Component
	Brokers                []string
	GroupID                string
	Topic                  string
	HeartbeatInterval      time.Duration
	PartitionWatchInterval time.Duration
	WatchPartitionChanges  bool
	SessionTimeout         time.Duration
	RebalanceTimeout       time.Duration
	JoinGroupBackoff       time.Duration
	StartOffset            int64
	RetentionTime          time.Duration
	Reader                 readerOptions
	logMode                bool
}

func NewConsumerGroup(options ConsumerGroupOptions) (*ConsumerGroup, error) {
	logger := newKafkaLogger(options.Logger)
	errorLogger := newKafkaErrorLogger(options.Logger)
	group, err := kafka.NewConsumerGroup(kafka.ConsumerGroupConfig{
		Brokers:                options.Brokers,
		ID:                     options.GroupID,
		Topics:                 []string{options.Topic},
		HeartbeatInterval:      options.HeartbeatInterval,
		PartitionWatchInterval: options.PartitionWatchInterval,
		WatchPartitionChanges:  options.WatchPartitionChanges,
		SessionTimeout:         options.SessionTimeout,
		RebalanceTimeout:       options.RebalanceTimeout,
		JoinGroupBackoff:       options.JoinGroupBackoff,
		StartOffset:            options.StartOffset,
		RetentionTime:          options.RetentionTime,
		Logger:                 logger,
		ErrorLogger:            errorLogger,
	})
	if err != nil {
		return nil, err
	}

	cg := &ConsumerGroup{
		logger:    options.Logger,
		group:     group,
		events:    make(chan interface{}, 100),
		processor: defaultProcessor,
		options:   &options,
	}
	go cg.run()

	return cg, nil
}

func (cg *ConsumerGroup) wrapProcessor(wrapFn Interceptor) {
	cg.processor = func(fn processFn) error {
		return wrapFn(fn)(&cmd{req: make([]interface{}, 0, 1), ctx: context.Background()})
	}
}

func (cg *ConsumerGroup) run() {
	cg.readerWg.Add(1)
	defer cg.readerWg.Done()

	for {
		gen, err := cg.group.Next(context.TODO())
		cg.genMu.Lock()
		cg.currentGen = gen
		cg.genMu.Unlock()

		if err != nil {
			if errors.Is(err, kafka.ErrGroupClosed) {
				return
			}

			cg.events <- err
			return
		}

		// Organize partitions
		topicPartitions := createTopicPartitionsFromGenAssignments(gen.Assignments)

		// We could have multiple Readers but we only want to emit RevokedPartitions event once
		var revokeOnce sync.Once

		// Emit AssignedPartitions event
		cg.events <- AssignedPartitions{
			Partitions: topicPartitions,
		}

		// We don't support multiple topics yet.
		assignments, ok := gen.Assignments[cg.options.Topic]
		if !ok {
			cg.events <- fmt.Errorf("topic \"%s\" not found in assignments", cg.options.Topic)
			break
		}

		// Listen to all partitions
		for _, assignment := range assignments {
			partition, offset := assignment.ID, assignment.Offset

			logger := newKafkaLogger(cg.logger)
			errorLogger := newKafkaErrorLogger(cg.logger)
			gen.Start(func(ctx context.Context) {
				reader := kafka.NewReader(kafka.ReaderConfig{
					Brokers:         cg.options.Brokers,
					Topic:           cg.options.Topic,
					Partition:       partition,
					MinBytes:        cg.options.Reader.MinBytes,
					MaxBytes:        cg.options.Reader.MaxBytes,
					MaxWait:         cg.options.Reader.MaxWait,
					ReadLagInterval: cg.options.Reader.ReadLagInterval,
					Logger:          logger,
					ErrorLogger:     errorLogger,
					CommitInterval:  cg.options.Reader.CommitInterval,
					ReadBackoffMin:  cg.options.Reader.ReadBackoffMin,
					ReadBackoffMax:  cg.options.Reader.ReadBackoffMax,
				})
				defer reader.Close()

				// seek to the last committed offset for this partition.
				reader.SetOffset(offset)
				for {
					msg, err := reader.FetchMessage(ctx)

					switch err {
					case kafka.ErrGroupClosed:
						return
					case kafka.ErrGenerationEnded:
						// emit RevokedPartitions event
						revokeOnce.Do(func() {
							cg.events <- RevokedPartitions{
								Partitions: topicPartitions,
							}
						})

						return
					case io.EOF:
						// Reader has been closed
						return
					case nil:
						// message received.
						cg.events <- msg
					default:
						cg.events <- err
					}
				}
			})
		}
	}
}

func (cg *ConsumerGroup) Poll(ctx context.Context) (msg interface{}, err error) {
	err = cg.processor(func(c *cmd) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case msg = <-cg.events:
			var name string
			switch msg.(type) {
			case AssignedPartitions:
				name = "AssignedPartitions"
			case RevokedPartitions:
				name = "RevokedPartitions"
			case Message:
				name = "FetchMessage"
			default:
				name = "FetchError"
			}
			logCmd(cg.options.logMode, c, name, msg)
			return nil
		}
	})
	return
}

func (cg *ConsumerGroup) CommitMessages(ctx context.Context, messages ...Message) error {
	return cg.processor(func(c *cmd) error {
		logCmd(cg.options.logMode, c, "CommitMessages", nil, messages)

		cg.genMu.RLock()
		if cg.currentGen == nil {
			cg.genMu.RUnlock()
			return fmt.Errorf("generation haven't been created yet")
		}

		partitions := make(map[int]int64)
		for _, message := range messages {
			messageOffset := message.Offset + 1
			currentOffset, ok := partitions[message.Partition]
			if ok && currentOffset >= messageOffset {
				continue
			}
			partitions[message.Partition] = messageOffset
		}

		offsets := make(map[string]map[int]int64)
		offsets[cg.options.Topic] = partitions

		err := cg.currentGen.CommitOffsets(offsets)
		cg.genMu.RUnlock()

		return err
	})
}

func (cg *ConsumerGroup) Close() error {
	return cg.processor(func(c *cmd) error {
		logCmd(cg.options.logMode, c, "ConsumerClose", nil)

		err := cg.group.Close()
		cg.readerWg.Wait()
		close(cg.events)
		return err
	})
}
