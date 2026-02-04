// Package rabbitmq provides RabbitMQ inbound channel adapter functionality.
package rabbitmq

import (
	"context"
	"fmt"

	"github.com/jeffersonbrasilino/gomes/container"
	"github.com/jeffersonbrasilino/gomes/message"
	"github.com/jeffersonbrasilino/gomes/message/adapter"
	"github.com/jeffersonbrasilino/gomes/otel"
	"github.com/rabbitmq/amqp091-go"
)

// consumerChannelAdapterBuilder provides a builder pattern for creating
// RabbitMQ inbound channel adapters with connection and queue configuration.
type consumerChannelAdapterBuilder struct {
	*adapter.InboundChannelAdapterBuilder[amqp091.Delivery]
	connectionReferenceName string
	consumerName            string
}

// inboundChannelAdapter implements the InboundChannelAdapter interface for
// RabbitMQ, providing message consumption capabilities through a RabbitMQ
// consumer with automatic message translation and error handling.
type inboundChannelAdapter struct {
	consumer          *amqp091.Channel
	queue             string
	messageTranslator adapter.InboundChannelMessageTranslator[amqp091.Delivery]
	messageChannel    chan *message.Message
	errorChannel      chan error
	ctx               context.Context
	cancelCtx         context.CancelFunc
	otelTrace         otel.OtelTrace
}

// NewConsumerChannelAdapterBuilder creates a new RabbitMQ consumer channel
// adapter builder instance.
//
// Parameters:
//   - connectionReferenceName: reference name for the RabbitMQ connection
//   - queueName: the RabbitMQ queue to consume messages from
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

// NewInboundChannelAdapter creates a new RabbitMQ inbound channel adapter
// instance with OpenTelemetry tracing support. It automatically starts a
// goroutine to subscribe to the queue and process incoming messages.
//
// Parameters:
//   - consumer: the RabbitMQ channel for receiving messages
//   - queue: the RabbitMQ queue name to consume from
//   - messageTranslator: translator for converting RabbitMQ messages to
//     internal format
//
// Returns:
//   - *inboundChannelAdapter: configured inbound channel adapter with active
//     subscription
func NewInboundChannelAdapter(
	consumer *amqp091.Channel,
	queue string,
	messageTranslator adapter.InboundChannelMessageTranslator[amqp091.Delivery],
) *inboundChannelAdapter {
	ctx, cancel := context.WithCancel(context.Background())
	adp := &inboundChannelAdapter{
		consumer:          consumer,
		queue:             queue,
		messageTranslator: messageTranslator,
		messageChannel:    make(chan *message.Message),
		errorChannel:      make(chan error),
		ctx:               ctx,
		cancelCtx:         cancel,
		otelTrace:         otel.InitTrace("rabbitMQ-inbound-channel-adapter"),
	}
	go adp.subscribeOnQueue()
	return adp
}

// Build constructs a RabbitMQ inbound channel adapter from the dependency
// container by retrieving the connection and creating a consumer channel.
//
// Parameters:
//   - container: dependency container containing RabbitMQ connection
//
// Returns:
//   - endpoint.InboundChannelAdapter: configured inbound channel adapter with
//     retry and dead letter capabilities
//   - error: error if connection retrieval or consumer creation fails
func (c *consumerChannelAdapterBuilder) Build(
	container container.Container[any, any],
) (*adapter.InboundChannelAdapter, error) {
	con, err := container.Get(c.connectionReferenceName)

	if err != nil {
		return nil, fmt.Errorf(
			"[RabbitMQ-inbound-channel] connection %s does not exist",
			c.connectionReferenceName,
		)
	}

	consumer, err := con.(*connection).Consumer(c.ReferenceName())
	if err != nil {
		return nil, fmt.Errorf(
			"[RabbitMQ-inbound-channel] consumer %s could not be created: %s",
			c.connectionReferenceName,
			err.Error(),
		)
	}
	adapter := NewInboundChannelAdapter(
		consumer,
		c.ReferenceName(),
		c.MessageTranslator(),
	)
	return c.InboundChannelAdapterBuilder.BuildInboundAdapter(adapter), nil
}

// Name returns the queue name of the RabbitMQ inbound channel adapter.
//
// Returns:
//   - string: the queue name
func (a *inboundChannelAdapter) Name() string {
	return a.queue
}

// Receive receives a message from the RabbitMQ queue using a non-blocking
// select pattern that respects context cancellation.
//
// Parameters:
//   - ctx: context for cancellation and timeout control
//
// Returns:
//   - *message.Message: the received message with translated payload and
//     headers
//   - error: error if receiving fails, context is cancelled, or channel is
//     closed
func (a *inboundChannelAdapter) Receive(
	ctx context.Context,
) (*message.Message, error) {
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

// subscribeOnQueue subscribes to the RabbitMQ queue and processes incoming
// messages continuously. This method runs in a separate goroutine and handles
// message translation and error propagation.
func (a *inboundChannelAdapter) subscribeOnQueue() {
	rabbitmqMessages, err := a.consumer.Consume(
		a.queue,
		"",    // consumer tag (server generates if empty)
		false, // auto-ack (we handle ack manually)
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // args
	)

	for {
		select {
		case <-a.ctx.Done():
			return
		default:
		}

		msg := <-rabbitmqMessages

		if err != nil {
			if err == context.Canceled {
				return
			}
			a.errorChannel <- err
		}

		message, translateErr := a.messageTranslator.ToMessage(msg)
		if translateErr != nil {
			a.errorChannel <- translateErr
		}

		select {
		case <-a.ctx.Done():
			return
		case a.messageChannel <- message:
		}
	}
}

// CommitMessage acknowledges a message to RabbitMQ, confirming successful
// processing. This removes the message from the queue.
//
// Parameters:
//   - msg: the message to acknowledge
//
// Returns:
//   - error: error if acknowledgment fails or message type is invalid
func (a *inboundChannelAdapter) CommitMessage(
	msg *message.Message,
) error {
	if externalMessage, ok := msg.GetRawMessage().(amqp091.Delivery); ok {
		return externalMessage.Ack(false)
	}
	return fmt.Errorf(
		"[rabbitmq-inbound-channel] failed to commit message",
	)
}
