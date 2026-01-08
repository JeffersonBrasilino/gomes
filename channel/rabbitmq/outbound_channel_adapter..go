// Package rabbitmq provides RabbitMQ outbound channel adapter functionality.
package rabbitmq

import (
	"context"
	"fmt"

	"github.com/jeffersonbrasilino/gomes/container"
	"github.com/jeffersonbrasilino/gomes/message"
	"github.com/jeffersonbrasilino/gomes/message/channel/adapter"
	"github.com/jeffersonbrasilino/gomes/message/endpoint"
	"github.com/jeffersonbrasilino/gomes/otel"
	amqp "github.com/rabbitmq/amqp091-go"
)

// publisherChannelAdapterBuilder provides a builder pattern for creating
// RabbitMQ outbound channel adapters with connection, queue, or exchange
// configuration.
type publisherChannelAdapterBuilder struct {
	*adapter.OutboundChannelAdapterBuilder[*amqp.Publishing]
	connectionReferenceName string
	exchangeRoutingKeys     string
	channelType             producerChannelType
	exchangeType            exchangeType
}

// outboundChannelAdapter implements the OutboundChannelAdapter interface for
// RabbitMQ, providing message publishing capabilities through RabbitMQ queues
// or exchanges with OpenTelemetry tracing support.
type outboundChannelAdapter struct {
	producer            *amqp.Channel
	channelName         string
	messageTranslator   adapter.OutboundChannelMessageTranslator[*amqp.Publishing]
	exchangeRoutingKeys string
	channelType         producerChannelType
	otelTrace           otel.OtelTrace
}

// NewPublisherChannelAdapterBuilder creates a new RabbitMQ publishing channel
// adapter builder. The default configuration publishes messages to the default
// RabbitMQ exchange using the channel name as the destination queue name
// (work-queues pattern).
//
// For publishing to specific exchanges, use the WithChannelType and
// WithExchangeType configuration methods.
//
// Parameters:
//   - connectionReferenceName: reference name for the RabbitMQ connection
//   - channelName: the RabbitMQ queue or exchange name to publish to
//
// Returns:
//   - *publisherChannelAdapterBuilder: configured builder instance with default
//     queue configuration
func NewPublisherChannelAdapterBuilder(
	connectionReferenceName string,
	channelName string,
) *publisherChannelAdapterBuilder {
	builder := &publisherChannelAdapterBuilder{
		adapter.NewOutboundChannelAdapterBuilder(
			channelName,
			channelName,
			NewMessageTranslator(),
		),
		connectionReferenceName,
		"",
		ProducerQueue,
		ExchangeDirect,
	}
	return builder
}

// NewOutboundChannelAdapter creates a new RabbitMQ outbound channel adapter
// instance with OpenTelemetry tracing support.
//
// Parameters:
//   - producer: the RabbitMQ channel for publishing messages
//   - channelName: the queue or exchange name
//   - messageTranslator: translator for converting internal messages to AMQP
//   - exchangeRoutingKeys: routing keys for exchange-based publishing
//   - channelType: type of producer channel (queue or exchange)
//
// Returns:
//   - *outboundChannelAdapter: configured outbound channel adapter
func NewOutboundChannelAdapter(
	producer *amqp.Channel,
	channelName string,
	messageTranslator adapter.OutboundChannelMessageTranslator[*amqp.Publishing],
	exchangeRoutingKeys string,
	channelType producerChannelType,
) *outboundChannelAdapter {
	return &outboundChannelAdapter{
		producer:            producer,
		channelName:         channelName,
		messageTranslator:   messageTranslator,
		exchangeRoutingKeys: exchangeRoutingKeys,
		channelType:         channelType,
		otelTrace:           otel.InitTrace("rabbitmq-outbound-channel-adapter"),
	}
}

// WithExchangeRoutingKeys sets the routing keys for exchange-based message
// routing. This is only applicable when channelType is set to ProducerExchange.
//
// Parameters:
//   - routingKeys: routing key pattern for message routing (e.g., "user.created")
//
// Returns:
//   - *publisherChannelAdapterBuilder: builder for method chaining
func (b *publisherChannelAdapterBuilder) WithExchangeRoutingKeys(
	routingKeys string,
) *publisherChannelAdapterBuilder {
	b.exchangeRoutingKeys = routingKeys
	return b
}

// WithExchangeType defines the exchange type for the output channel. This is
// only applicable when channelType is set to ProducerExchange.
//
// Parameters:
//   - value: exchange type (ExchangeDirect, ExchangeFanout, ExchangeTopic,
//     ExchangeHeaders)
//
// Returns:
//   - *publisherChannelAdapterBuilder: builder for method chaining
func (b *publisherChannelAdapterBuilder) WithExchangeType(
	value exchangeType,
) *publisherChannelAdapterBuilder {
	b.exchangeType = value
	return b
}

// WithChannelType sets the type of RabbitMQ publisher channel.
//
// When set to ProducerQueue (default), messages are published to the default
// exchange with the channel name as the queue name (work-queues pattern).
//
// When set to ProducerExchange, messages are published to the exchange
// specified in the channel name with routing keys.
//
// Parameters:
//   - value: producer channel type (ProducerQueue or ProducerExchange)
//
// Returns:
//   - *publisherChannelAdapterBuilder: builder for method chaining
func (b *publisherChannelAdapterBuilder) WithChannelType(
	value producerChannelType,
) *publisherChannelAdapterBuilder {
	b.channelType = value
	return b
}

// Build constructs a RabbitMQ outbound channel adapter from the dependency
// container by retrieving the connection and creating a producer channel.
//
// Parameters:
//   - container: dependency container containing RabbitMQ connection
//
// Returns:
//   - endpoint.OutboundChannelAdapter: configured outbound channel adapter
//   - error: error if connection retrieval or producer creation fails
func (b *publisherChannelAdapterBuilder) Build(
	container container.Container[any, any],
) (endpoint.OutboundChannelAdapter, error) {
	con, err := container.Get(b.connectionReferenceName)

	if err != nil {
		return nil, fmt.Errorf(
			"[RabbitMQ-outbound-channel] connection %s does not exist",
			b.connectionReferenceName,
		)
	}

	producer, err := con.(*connection).Producer(
		b.ChannelName(),
		b.channelType,
		b.exchangeType,
	)

	if err != nil {
		return nil, fmt.Errorf(
			"[RabbitMQ-outbound-channel] %s",
			err.Error(),
		)
	}

	adapter := NewOutboundChannelAdapter(
		producer,
		b.ChannelName(),
		b.MessageTranslator(),
		b.exchangeRoutingKeys,
		b.channelType,
	)

	return b.OutboundChannelAdapterBuilder.BuildOutboundAdapter(adapter)
}

// Name returns the queue or exchange name of the RabbitMQ outbound channel
// adapter.
//
// Returns:
//   - string: the channel name
func (a *outboundChannelAdapter) Name() string {
	return a.channelName
}

// Send publishes a message to RabbitMQ queue or exchange with OpenTelemetry
// tracing support. The method automatically determines whether to publish to
// a queue (default exchange) or a specific exchange based on the channel type.
//
// Parameters:
//   - ctx: context for timeout and cancellation control
//   - msg: the message to be published with headers and payload
//
// Returns:
//   - error: error if context is cancelled, message translation fails, or
//     publishing fails
func (a *outboundChannelAdapter) Send(
	ctx context.Context,
	msg *message.Message,
) error {
	_, span := a.otelTrace.Start(
		ctx,
		"",
		otel.WithMessagingSystemType(otel.MessageSystemTypeRabbitMQ),
		otel.WithSpanOperation(otel.SpanOperationSend),
		otel.WithSpanKind(otel.SpanKindProducer),
		otel.WithMessage(msg),
	)
	defer span.End()

	select {
	case <-ctx.Done():
		return fmt.Errorf(
			"[RabbitMQ OUTBOUND CHANNEL] Context cancelled before sending",
		)
	default:
	}

	msgToSend, errP := a.messageTranslator.FromMessage(msg)
	if errP != nil {
		return errP
	}

	channelName := ""
	routingKey := a.channelName
	if a.channelType == ProducerExchange {
		channelName = a.channelName
		routingKey = a.exchangeRoutingKeys
	}

	err := a.producer.PublishWithContext(
		ctx,
		channelName,
		routingKey,
		false, // mandatory
		false, // immediate
		*msgToSend,
	)

	select {
	case <-ctx.Done():
		return fmt.Errorf(
			"[RabbitMQ OUTBOUND CHANNEL] Context cancelled after sending",
		)
	default:
	}
	return err
}

// Close gracefully closes the RabbitMQ producer channel and releases
// associated resources.
//
// Returns:
//   - error: error if closing fails (typically nil)
func (a *outboundChannelAdapter) Close() error {
	a.producer.Close()
	return nil
}
