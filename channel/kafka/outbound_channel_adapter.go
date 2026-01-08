// Package kafka provides Kafka integration for the message system.
//
// This package implements Kafka-specific channel adapters and connections for
// publishing and consuming messages through Apache Kafka. It provides outbound
// and inbound channel adapters with message translation capabilities.
//
// The OutboundChannelAdapter implementation supports:
// - Kafka producer integration for message publishing
// - Message translation between internal and Kafka formats
// - Context-aware message sending with timeout support
// - Error handling and connection management
package kafka

import (
	"context"
	"fmt"

	"github.com/jeffersonbrasilino/gomes/container"
	"github.com/jeffersonbrasilino/gomes/message"
	"github.com/jeffersonbrasilino/gomes/message/channel/adapter"
	"github.com/jeffersonbrasilino/gomes/message/endpoint"
	"github.com/jeffersonbrasilino/gomes/otel"
	"github.com/segmentio/kafka-go"
)

// publisherChannelAdapterBuilder provides a builder pattern for creating
// Kafka outbound channel adapters with connection and topic configuration.
type publisherChannelAdapterBuilder struct {
	*adapter.OutboundChannelAdapterBuilder[*kafka.Message]
	connectionReferenceName string
}

// outboundChannelAdapter implements the PublisherChannel interface for Kafka,
// providing message publishing capabilities through a Kafka producer.
type outboundChannelAdapter struct {
	producer          *kafka.Writer
	topicName         string
	messageTranslator adapter.OutboundChannelMessageTranslator[*kafka.Message]
	otelTrace         otel.OtelTrace
}

// NewPublisherChannelAdapterBuilder creates a new Kafka publisher channel
// adapter builder instance.
//
// Parameters:
//   - connectionReferenceName: reference name for the Kafka connection
//   - topicName: the Kafka topic to publish messages to
//
// Returns:
//   - *publisherChannelAdapterBuilder: configured builder instance
func NewPublisherChannelAdapterBuilder(
	connectionReferenceName string,
	topicName string,
) *publisherChannelAdapterBuilder {
	builder := &publisherChannelAdapterBuilder{
		adapter.NewOutboundChannelAdapterBuilder(
			topicName,
			topicName,
			NewMessageTranslator(),
		),
		connectionReferenceName,
	}
	return builder
}

// NewOutboundChannelAdapter creates a new Kafka outbound channel adapter instance.
//
// Parameters:
//   - producer: the Kafka sync producer for sending messages
//   - topicName: the Kafka topic name
//   - messageTranslator: translator for converting internal messages to Kafka format
//
// Returns:
//   - *outboundChannelAdapter: configured outbound channel adapter
func NewOutboundChannelAdapter(
	producer *kafka.Writer,
	topicName string,
	messageTranslator adapter.OutboundChannelMessageTranslator[*kafka.Message],
) *outboundChannelAdapter {

	return &outboundChannelAdapter{
		producer:          producer,
		topicName:         topicName,
		messageTranslator: messageTranslator,
		otelTrace:         otel.InitTrace("kafka-outbound-channel-adapter"),
	}
}

// Build constructs a Kafka outbound channel adapter from the dependency container.
//
// Parameters:
//   - container: dependency container containing required components
//
// Returns:
//   - endpoint.OutboundChannelAdapter: configured publisher channel
//   - error: error if construction fails
func (b *publisherChannelAdapterBuilder) Build(
	container container.Container[any, any],
) (endpoint.OutboundChannelAdapter, error) {
	con, err := container.Get(b.connectionReferenceName)

	if err != nil {
		return nil, fmt.Errorf(
			"[kafka-outbound-channel] connection %s does not exist",
			b.connectionReferenceName,
		)
	}

	producer := con.(*connection).Producer()
	adapter := NewOutboundChannelAdapter(producer, b.ChannelName(), b.MessageTranslator())

	return b.OutboundChannelAdapterBuilder.BuildOutboundAdapter(adapter)
}

// Name returns the topic name of the Kafka outbound channel adapter.
//
// Returns:
//   - string: the topic name
func (a *outboundChannelAdapter) Name() string {
	return a.topicName
}

// Send publishes a message to the Kafka topic with context support.
//
// Parameters:
//   - ctx: context for timeout/cancellation control
//   - msg: the message to be published
//
// Returns:
//   - error: error if sending fails or context is cancelled
func (a *outboundChannelAdapter) Send(ctx context.Context, msg *message.Message) error {

	_, span := a.otelTrace.Start(
		ctx,
		"",
		otel.WithMessagingSystemType(otel.MessageSystemTypeKafka),
		otel.WithSpanOperation(otel.SpanOperationSend),
		otel.WithSpanKind(otel.SpanKindProducer),
		otel.WithMessage(msg),
	)
	defer span.End()

	select {
	case <-ctx.Done():
		err := fmt.Errorf("[KAFKA OUTBOUND CHANNEL] Context cancelled after processing before sending result. ")
		span.Error(err, err.Error())
		return err
	default:
	}

	msgToSend, errP := a.messageTranslator.FromMessage(msg)

	if errP != nil {
		span.Error(errP, errP.Error())
		return errP
	}

	err := a.producer.WriteMessages(ctx, *msgToSend)

	select {
	case <-ctx.Done():
		err := fmt.Errorf("[KAFKA OUTBOUND CHANNEL] Context cancelled after processing after sending result. ")
		span.Error(err, err.Error())
		return err
	default:
	}

	if err != nil {
		span.Error(err, err.Error())
	} else {
		span.Success("message sent to kafka topic successfully")
	}

	return err
}

func (a *outboundChannelAdapter) Close() error {
	return a.producer.Close()
}
