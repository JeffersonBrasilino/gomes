// Package rabbitmq provides RabbitMQ outbound channel adapter functionality.
package rabbitmq

import (
	"context"
	"fmt"

	"github.com/jeffersonbrasilino/gomes/container"
	"github.com/jeffersonbrasilino/gomes/message"
	"github.com/jeffersonbrasilino/gomes/message/adapter"
	"github.com/jeffersonbrasilino/gomes/message/endpoint"
	"github.com/jeffersonbrasilino/gomes/otel"
	amqp "github.com/rabbitmq/amqp091-go"
)

// Producer channel type constants.
const (
	ProducerQueue producerChannelType = iota
	ProducerExchange
)

// Exchange type constants following AMQP specification.
const (
	ExchangeFanout exchangeType = iota
	ExchangeDirect
	ExchangeTopic
	ExchangeHeaders
)

type exchangeType int8

type producerChannelType int8

// publisherChannelAdapterBuilder provides a builder pattern for creating
// RabbitMQ outbound channel adapters with connection, queue, or exchange
// configuration.
type publisherChannelAdapterBuilder struct {
	*adapter.OutboundChannelAdapterBuilder[*amqp.Publishing]
	connectionReferenceName string
	exchangeRoutingKeys     string
	channelType             producerChannelType
	exchangeType            exchangeType
	durable                 bool
	deleteUnused            bool
	exclusive               bool
	noWait                  bool
	args                    amqp.Table
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
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
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

// Type returns the AMQP exchange type string representation.
//
// Returns:
//   - string: AMQP exchange type constant
func (t exchangeType) Type() string {
	switch t {
	case ExchangeFanout:
		return amqp.ExchangeFanout
	case ExchangeHeaders:
		return amqp.ExchangeHeaders
	case ExchangeTopic:
		return amqp.ExchangeTopic
	default:
		return amqp.ExchangeDirect
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

// WithDurable sets the durability flag for the queue or exchange declaration.
// When true, the queue/exchange will survive broker restart.
//
// Parameters:
//   - data: durability flag (true = durable, false = transient)
//
// Returns:
//   - *publisherChannelAdapterBuilder: builder for method chaining
func (b *publisherChannelAdapterBuilder) WithDurable(data bool) *publisherChannelAdapterBuilder {
	b.durable = data
	return b
}

// WithDeleteUnused sets whether the queue or exchange should be deleted when
// it is no longer in use (no consumers/bindings).
//
// Parameters:
//   - data: delete when unused flag (true = auto-delete, false = persistent)
//
// Returns:
//   - *publisherChannelAdapterBuilder: builder for method chaining
func (b *publisherChannelAdapterBuilder) WithDeleteUnused(data bool) *publisherChannelAdapterBuilder {
	b.deleteUnused = data
	return b
}

// WithExclusive sets the exclusivity flag for the queue or exchange.
// When true, the resource is exclusive to the connection that declares it.
//
// Parameters:
//   - data: exclusive flag (true = exclusive, false = non-exclusive)
//
// Returns:
//   - *publisherChannelAdapterBuilder: builder for method chaining
func (b *publisherChannelAdapterBuilder) WithExclusive(data bool) *publisherChannelAdapterBuilder {
	b.exclusive = data
	return b
}

// WithNoWait sets the no-wait flag for the queue or exchange declaration.
// When true, the method returns immediately without waiting for the server
// to confirm the operation.
//
// Parameters:
//   - data: no-wait flag (true = no-wait, false = wait for confirmation)
//
// Returns:
//   - *publisherChannelAdapterBuilder: builder for method chaining
func (b *publisherChannelAdapterBuilder) WithNoWait(data bool) *publisherChannelAdapterBuilder {
	b.noWait = data
	return b
}

// WithArguments sets optional arguments table for the queue or exchange
// declaration. Arguments can contain vendor-specific extensions and
// modifications (e.g., message TTL, max length).
//
// Parameters:
//   - args: AMQP table containing optional arguments
//
// Returns:
//   - *publisherChannelAdapterBuilder: builder for method chaining
func (b *publisherChannelAdapterBuilder) WithArguments(args amqp.Table) *publisherChannelAdapterBuilder {
	b.args = args
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

	producer, err := con.(*connection).GetConnection().Channel()
	if err != nil {
		return nil, fmt.Errorf(
			"[RabbitMQ-outbound-channel] failed to create producer channel: %w",
			err,
		)
	}

	if b.channelType == ProducerExchange {
		err = producer.ExchangeDeclare(
			b.ChannelName(),
			b.exchangeType.Type(),
			b.durable,
			b.deleteUnused,
			b.exclusive,
			b.noWait,
			b.args,
		)
	} else {
		_, err = producer.QueueDeclare(
			b.ChannelName(),
			b.durable,
			b.deleteUnused,
			b.exclusive,
			b.noWait,
			b.args,
		)
	}

	if err != nil {
		return nil, fmt.Errorf(
			"[RabbitMQ-outbound-channel] failed to declare channel: %w",
			err,
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
