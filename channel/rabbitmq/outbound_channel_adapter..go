package rabbitmq

import (
	"context"
	"fmt"

	"github.com/jeffersonbrasilino/gomes/container"
	"github.com/jeffersonbrasilino/gomes/message"
	"github.com/jeffersonbrasilino/gomes/message/channel/adapter"
	"github.com/jeffersonbrasilino/gomes/message/endpoint"
	amqp "github.com/rabbitmq/amqp091-go"
)

// publisherChannelAdapterBuilder provides a builder pattern for creating
// RabbitMQ outbound channel adapters with connection and topic configuration.
type publisherChannelAdapterBuilder struct {
	*adapter.OutboundChannelAdapterBuilder[*amqp.Publishing]
	connectionReferenceName string
	exchangeRoutingKeys     string
	channelType             producerChannelType
	exchangeType            exchangeType
}

type outboundChannelAdapter struct {
	producer            *amqp.Channel
	channelName         string
	messageTranslator   adapter.OutboundChannelMessageTranslator[*amqp.Publishing]
	exchangeRoutingKeys string
	channelType         producerChannelType
}

// NewPublisherChannelAdapterBuilder creates a new RabbitMQ publishing channel
// The default creation pattern is a channel adapter that will publish the message to the default
// RabbitMQ exchange and use the channel name as the destination queue name (work-queues).
//
// If the output channel is a specific exchange, use the exchange configuration methods.
//
// Parameters:
// - connectionReferenceName: reference name for the RabbitMQ connection
// - queueName: the RabbitMQ queue to publish messages to
//
// Returns:
// - *publisherChannelAdapterBuilder: configured builder instance
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
	}
}

// WithExchangeRoutingKeys set the exchange routing for the output channel;
// Only valid when channelType = producerExchange
//
// Parameters:
//   - value: string exchange name
//
// Returns:
//   - publisherChannelAdapterBuilder: builder for publisher channel
func (b *publisherChannelAdapterBuilder) WithExchangeRoutingKeys(
	routingKeys string,
) *publisherChannelAdapterBuilder {
	b.exchangeRoutingKeys = routingKeys
	return b
}

// WithExchangeType defines the exchange type the output channel;
// Valid only when channelType = producerExchange
//
// Parameters:
//	 - value: exchangeType exchange type
//
// Returns:
// 	 - publisherChannelAdapterBuilder: builder for the publisher channel
func (b *publisherChannelAdapterBuilder) WithExchangeType(
	value exchangeType,
) *publisherChannelAdapterBuilder {
	b.exchangeType = value
	return b
}

// WithChannelType set type of rabbitMQ publisher channel;
//
// Possible values: ProducerExchange or ProducerQueue
// When the type is ProducerQueue, Rabbit will publish the message to the default exchange
// and assume the channel name is the queue name.
//
// When the type is ProducerExchange, the Adapter will publish the message to the
// exchange specified in the channel name.
//
// Parameters:
//   - value: producerChannelType value type of rabbitMQ channel
//
// Returns:
//   - publisherChannelAdapterBuilder: builder for publisher channel
func (b *publisherChannelAdapterBuilder) WithChannelType(
	value producerChannelType,
) *publisherChannelAdapterBuilder {
	b.channelType = value
	return b
}

// Build constructs a RabbitMQ outbound channel adapter from the dependency container.
//
// Parameters:
//   - container: dependency container containing required components
//
// Returns:
//   - message.PublisherChannel: configured publisher channel
//   - error: error if construction fails
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

	producer, err := con.(*connection).Producer(b.ChannelName(), b.channelType, b.exchangeType)

	if err != nil {
		return nil, fmt.Errorf("[RabbitMQ-outbound-channel] %s", err.Error())
	}

	adapter := NewOutboundChannelAdapter(
		producer,
		b.ChannelName(),
		b.MessageTranslator(),
		b.exchangeRoutingKeys,
		b.channelType,
	)

	//container.Set(fmt.Sprintf("outbound-channel-adapter:%s", b.ChannelName()), adapter)

	return b.OutboundChannelAdapterBuilder.BuildOutboundAdapter(adapter)
}

// Name returns the queue name of the RabbitMQ outbound channel adapter.
//
// Returns:
//   - string: the queue name
func (a *outboundChannelAdapter) Name() string {
	return a.channelName
}

// Send publishes a message to the RabbitMQ topic with context support.
//
// Parameters:
//   - ctx: context for timeout/cancellation control
//   - msg: the message to be published
//
// Returns:
//   - error: error if sending fails or context is cancelled
func (a *outboundChannelAdapter) Send(ctx context.Context, msg *message.Message) error {
	select {
	case <-ctx.Done():
		return fmt.Errorf(
			"[RabbitMQ OUTBOUND CHANNEL] Context cancelled after processing before sending result",
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
		false,
		false,
		*msgToSend,
	)

	select {
	case <-ctx.Done():
		return fmt.Errorf(
			"[RabbitMQ OUTBOUND CHANNEL] Context cancelled after processing after sending result",
		)
	default:
	}
	return err
}

func (a *outboundChannelAdapter) Close() error {
	a.producer.Close()
	return nil
}
