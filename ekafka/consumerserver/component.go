package consumerserver

import (
	"context"
	"errors"
	"fmt"

	"github.com/gotomicro/ego-component/ekafka"
	"github.com/gotomicro/ego/core/constant"
	"github.com/gotomicro/ego/core/elog"
	"github.com/gotomicro/ego/server"
	"github.com/segmentio/kafka-go"
)

// Interface check
var _ server.Server = (*Component)(nil)

// PackageName is the name of this component.
const PackageName = "component.ekafka.consumerserver"

type consumptionMode int

const (
	consumptionModeSingle consumptionMode = iota + 1
)

// Component starts an Ego server for message consuming.
type Component struct {
	config               *config
	name                 string
	ekafkaComponent      *ekafka.Component
	logger               *elog.Component
	serverCtx            context.Context
	stopServer           context.CancelFunc
	mode                 consumptionMode
	singleMessageHandler SingleMessageHandler
	consumptionErrors    chan error
}

// PackageName returns the package name.
func (cmp *Component) PackageName() string {
	return PackageName
}

// Info returns server info, used by governor and consumer balancer.
func (cmp *Component) Info() *server.ServiceInfo {
	info := server.ApplyOptions(
		server.WithKind(constant.ServiceProvider),
	)
	return &info
}

// GracefulStop stops the server.
func (cmp *Component) GracefulStop(ctx context.Context) error {
	cmp.stopServer()
	return nil
}

// Stop stops the server.
func (cmp *Component) Stop() error {
	cmp.stopServer()
	return nil
}

// Init ...
func (cmp *Component) Init() error {
	return nil
}

// Name returns the name of this instance.
func (cmp *Component) Name() string {
	return cmp.name
}

// Start will start consuming.
func (cmp *Component) Start() error {
	switch cmp.mode {
	case consumptionModeSingle:
		return cmp.startSingleMode()
	default:
		return fmt.Errorf("undefined consumption mode: %v", cmp.mode)
	}
}

// GetConsumer returns the default consumer.
func (cmp *Component) GetConsumer() *ekafka.Consumer {
	return cmp.ekafkaComponent.Consumer(cmp.config.ConsumerName)
}

// EachMessage registers a single message handler.
func (cmp *Component) EachMessage(consumptionErrors chan error, handler SingleMessageHandler) error {
	cmp.consumptionErrors = consumptionErrors
	cmp.mode = consumptionModeSingle
	cmp.singleMessageHandler = handler
	return nil
}

func (cmp *Component) startSingleMode() error {
	consumer := cmp.GetConsumer()

	if cmp.singleMessageHandler == nil {
		return errors.New("you must define a MessageHandler first")
	}

	unrecoverableError := make(chan error)
	go func() {
		for {
			message, err := consumer.ReadMessage(cmp.serverCtx)
			if err != nil {
				cmp.consumptionErrors <- err
				cmp.logger.Error("encountered an error while reading message", elog.FieldErr(err))

				if kafkaError, ok := err.(kafka.Error); ok {
					// If this error is unrecoverable, stop consuming.
					if kafkaError.Temporary() == false {
						unrecoverableError <- kafkaError
						return
					}
				}
			}

			if cmp.serverCtx.Err() != nil {
				return
			}

			err = cmp.singleMessageHandler(message)
			if err != nil {
				cmp.logger.Error("encountered an error while handling message", elog.FieldErr(err))
				cmp.consumptionErrors <- err
			}
		}
	}()

	select {
	case <-cmp.serverCtx.Done():
		rootErr := cmp.serverCtx.Err()
		cmp.logger.Error("terminating consumer because a context error", elog.FieldErr(rootErr))

		err := cmp.closeConsumer(consumer)
		if err != nil {
			return fmt.Errorf("encountered an error while closing consumer: %w", err)
		}

		if errors.Is(rootErr, context.Canceled) {
			return nil
		}

		return rootErr
	case rootErr := <-unrecoverableError:
		if rootErr == nil {
			panic("unrecoverableError should receive an error instead of nil")
		}

		cmp.logger.Fatal("stopping server because of an unrecoverable error", elog.FieldErr(rootErr))
		cmp.Stop()

		err := cmp.closeConsumer(consumer)
		if err != nil {
			return fmt.Errorf("exiting due to an unrecoverable error, but encountered an error while closing consumer: %w", err)
		}
		return rootErr
	}
}

func (cmp *Component) closeConsumer(consumer *ekafka.Consumer) error {
	if err := consumer.Close(); err != nil {
		cmp.logger.Fatal("failed to close kafka writer", elog.FieldErr(err))
		return err
	}
	cmp.logger.Info("consumer server terminated")
	return nil
}

// NewConsumerServerComponent creates a new server instance.
func NewConsumerServerComponent(name string, config *config, ekafkaComponent *ekafka.Component, logger *elog.Component) *Component {
	serverCtx, stopServer := context.WithCancel(context.Background())
	return &Component{
		name:            name,
		config:          config,
		ekafkaComponent: ekafkaComponent,
		logger:          logger,
		serverCtx:       serverCtx,
		stopServer:      stopServer,
		mode:            consumptionModeSingle,
	}
}
