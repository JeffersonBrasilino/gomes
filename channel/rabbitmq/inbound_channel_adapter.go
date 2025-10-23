package rabbitmq

import (
	"context"
	"fmt"

	"github.com/jeffersonbrasilino/gomes/container"
	"github.com/jeffersonbrasilino/gomes/message"
	"github.com/jeffersonbrasilino/gomes/message/channel/adapter"
	"github.com/jeffersonbrasilino/gomes/message/endpoint"
	"github.com/wagslane/go-rabbitmq"
)

// consumerChannelAdapterBuilder provides a builder pattern for creating
// RabbitMQ inbound channel adapters with connection and topic configuration.
type consumerChannelAdapterBuilder struct {
	*adapter.InboundChannelAdapterBuilder[any]
	connectionReferenceName string
	consumerName            string
}

// inboundChannelAdapter implements the InboundChannelAdapter interface for RabbitMQ,
// providing message consumption capabilities through a RabbitMQ consumer.
type inboundChannelAdapter struct {
	consumer          *rabbitmq.Consumer
	topic             string
	messageTranslator adapter.InboundChannelMessageTranslator[any]
	messageChannel    chan *message.Message
	errorChannel      chan error
	ctx               context.Context
	cancelCtx         context.CancelFunc
}

// NewConsumerChannelAdapterBuilder creates a new RabbitMQ consumer channel
// adapter builder instance.
//
// Parameters:
//   - connectionReferenceName: reference name for the RabbitMQ connection
//   - queueName: the RabbitMQ topic to consume messages from
//   - consumerName: the consumer group name
//
// Returns:
//   - *consumerChannelAdapterBuilder: configured builder instance
func NewConsumerChannelAdapterBuilder(
	connectionReferenceName string,
	queueName string,
	consumerName string,
) *consumerChannelAdapterBuilder {
	builder := &consumerChannelAdapterBuilder{
		adapter.NewInboundChannelAdapterBuilder(
			consumerName,
			queueName,
			NewMessageTranslator(),
		),
		connectionReferenceName,
		consumerName,
	}
	return builder
}

// NewInboundChannelAdapter creates a new RabbitMQ inbound channel adapter instance.
//
// Parameters:
//   - consumer: the RabbitMQ consumer for receiving messages
//   - topic: the RabbitMQ topic name
//   - messageTranslator: translator for converting RabbitMQ messages to internal format
//
// Returns:
//   - *inboundChannelAdapter: configured inbound channel adapter
func NewInboundChannelAdapter(
	consumer *rabbitmq.Consumer,
	topic string,
	messageTranslator adapter.InboundChannelMessageTranslator[any],
) *inboundChannelAdapter {
	ctx, cancel := context.WithCancel(context.Background())
	adp := &inboundChannelAdapter{
		consumer:          consumer,
		topic:             topic,
		messageTranslator: messageTranslator,
		messageChannel:    make(chan *message.Message),
		errorChannel:      make(chan error),
		ctx:               ctx,
		cancelCtx:         cancel,
	}
	go adp.subscribeOnQueue()
	return adp
}

// Build constructs a RabbitMQ inbound channel adapter from the dependency container.
//
// Parameters:
//   - container: dependency container containing required components
//
// Returns:
//   - message.InboundChannelAdapter: configured inbound channel adapter
//   - error: error if construction fails
func (c *consumerChannelAdapterBuilder) Build(
	container container.Container[any, any],
) (endpoint.InboundChannelAdapter, error) {
	con, err := container.Get(c.connectionReferenceName)

	if err != nil {
		return nil, fmt.Errorf(
			"[RabbitMQ-inbound-channel] connection %s does not exist",
			c.connectionReferenceName,
		)
	}

	consumer := con.(*connection).Consumer(c.ReferenceName())
	adapter := NewInboundChannelAdapter(consumer, c.ReferenceName(), c.MessageTranslator())
	return c.InboundChannelAdapterBuilder.BuildInboundAdapter(adapter), nil
}

// Name returns the topic name of the RabbitMQ inbound channel adapter.
//
// Returns:
//   - string: the topic name
func (a *inboundChannelAdapter) Name() string {
	return a.topic
}

// Receive receives a message from the RabbitMQ topic.
//
// Parameters:
//   - ctx: context
//
// Returns:
//   - *message.Message: the received message
//   - error: error if receiving fails or channel is closed
func (a *inboundChannelAdapter) Receive(ctx context.Context) (*message.Message, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-a.ctx.Done():
		return nil, a.ctx.Err()
	case msg := <-a.messageChannel:
		return msg, nil
	case err := <-a.errorChannel:
		return nil, err
	}
}

// Close gracefully closes the RabbitMQ inbound channel adapter and stops
// message consumption.
//
// Returns:
//   - error: error if closing fails (typically nil)
func (a *inboundChannelAdapter) Close() error {
	a.cancelCtx()
	a.consumer.Close()
	close(a.messageChannel)
	close(a.errorChannel)
	return nil
}

// subscribeOnTopic subscribes to the RabbitMQ topic and processes incoming messages.
// This method runs in a separate goroutine and continuously polls for messages.
func (a *inboundChannelAdapter) subscribeOnQueue() {
	for {
		select {
		case <-a.ctx.Done():
			return
		default:
		}

		err := a.consumer.Run(func(d rabbitmq.Delivery) rabbitmq.Action{
			fmt.Printf("consumed: %v", string(d.Body)) 
			return rabbitmq.Ack
		})

		if err != nil {
			if err == context.Canceled {
				return
			}
			a.errorChannel <- err
		}

		/* message, translateErr := a.messageTranslator.ToMessage(nil)
		if translateErr != nil {
			a.errorChannel <- translateErr
		} */

		select {
		case <-a.ctx.Done():
			return
		case a.messageChannel <- nil:
		}
	}
}