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
	exclusive               bool
	noLocal                 bool
	noWait                  bool
	args                    amqp091.Table
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
	otelTrace         otel.OtelTrace
	noLocal           bool
	exclusive         bool
	noWait            bool
	args              amqp091.Table
	stopTrigger       chan bool
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
		true,  // durable
		false, // no-local
		false, // no-wait
		nil,   // arguments
	}
	return builder
}

// WithNoLocal sets whether the queue or exchange should be deleted when
// it is no longer in use (no consumers/bindings).
//
// Parameters:
//   - data: delete when unused flag (true = auto-delete, false = persistent)
//
// Returns:
//   - *consumerChannelAdapterBuilder: builder for method chaining
func (c *consumerChannelAdapterBuilder) WithNoLocal(data bool) *consumerChannelAdapterBuilder {
	c.noLocal = data
	return c
}

// WithExclusive sets the exclusivity flag for the queue or exchange.
// When true, the resource is exclusive to the connection that declares it.
//
// Parameters:
//   - data: exclusive flag (true = exclusive, false = non-exclusive)
//
// Returns:
//   - *consumerChannelAdapterBuilder: builder for method chaining
func (c *consumerChannelAdapterBuilder) WithExclusive(data bool) *consumerChannelAdapterBuilder {
	c.exclusive = data
	return c
}

// WithNoWait sets the no-wait flag for the queue or exchange declaration.
// When true, the method returns immediately without waiting for the server
// to confirm the operation.
//
// Parameters:
//   - data: no-wait flag (true = no-wait, false = wait for confirmation)
//
// Returns:
//   - *consumerChannelAdapterBuilder: builder for method chaining
func (c *consumerChannelAdapterBuilder) WithNoWait(data bool) *consumerChannelAdapterBuilder {
	c.noWait = data
	return c
}

// WithArguments sets optional arguments table for the queue or exchange
// declaration. Arguments can contain vendor-specific extensions and
// modifications (e.g., message TTL, max length).
//
// Parameters:
//   - args: AMQP table containing optional arguments
//
// Returns:
//   - *consumerChannelAdapterBuilder: builder for method chaining
func (c *consumerChannelAdapterBuilder) WithArguments(args amqp091.Table) *consumerChannelAdapterBuilder {
	c.args = args
	return c
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

	consumer, err := con.(*connection).GetConnection().Channel()
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
		c.noLocal,
		c.exclusive,
		c.noWait,
		c.args,
	)
	return c.InboundChannelAdapterBuilder.BuildInboundAdapter(adapter), nil
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
	noLocal bool,
	exclusive bool,
	noWait bool,
	args amqp091.Table,
) *inboundChannelAdapter {
	adp := &inboundChannelAdapter{
		consumer:          consumer,
		queue:             queue,
		messageTranslator: messageTranslator,
		messageChannel:    make(chan *message.Message),
		errorChannel:      make(chan error),
		otelTrace:         otel.InitTrace("rabbitMQ-inbound-channel-adapter"),
		stopTrigger:       make(chan bool),
	}
	go adp.subscribeOnQueue()
	return adp
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
	close(a.stopTrigger)
	a.consumer.Close()
	return nil
}

// subscribeOnQueue subscribes to the RabbitMQ queue and processes incoming
// messages continuously. This method runs in a separate goroutine and handles
// message translation and error propagation.
func (a *inboundChannelAdapter) subscribeOnQueue() {
	defer func() {
		close(a.messageChannel)
		close(a.errorChannel)
	}()

	rabbitmqMessages, err := a.consumer.Consume(
		a.queue,
		"",    // consumer tag (server generates if empty)
		false, // auto-ack (we handle ack manually)
		a.exclusive,
		a.noLocal,
		a.noWait,
		a.args,
	)

	if err != nil {
        a.errorChannel <- fmt.Errorf("failed to start consuming: %w", err)
        return
    }

	for msg := range rabbitmqMessages {
		message, translateErr := a.messageTranslator.ToMessage(msg)
		if translateErr != nil {
			select {
			case a.errorChannel <- translateErr:
			case <-a.stopTrigger:
				return
			}
			continue
		}

		select {
		case <-a.stopTrigger:
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
