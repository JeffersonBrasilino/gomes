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
	"github.com/jeffersonbrasilino/gomes/message/adapter"
	"github.com/jeffersonbrasilino/gomes/message/endpoint"
	"github.com/jeffersonbrasilino/gomes/otel"
	"github.com/segmentio/kafka-go"
)

// publisherChannelAdapterBuilder provides a builder pattern for creating
// Kafka outbound channel adapters with connection and topic configuration.
type publisherChannelAdapterBuilder struct {
	*adapter.OutboundChannelAdapterBuilder[*kafka.Message]
	connectionReferenceName string
	maxAttempts             int
	batchSize               int
	batchBytes              int64
	async                   bool
	requiredAcks            int
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
		10,
		100,
		1048576,
		true,
		0,
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

// WithMaxAttempts sets the maximum number of attempts for sending messages.
// Failed sends will be retried up to this many times.
//
// Parameters:
//   - attempts: maximum number of send attempts
//
// Returns:
//   - *publisherChannelAdapterBuilder: builder instance for chaining
func (b *publisherChannelAdapterBuilder) WithMaxAttempts(
	attempts int,
) *publisherChannelAdapterBuilder {
	b.maxAttempts = attempts
	return b
}

// WithBatchSize sets the number of messages to batch before sending.
// Larger batches improve throughput but increase latency.
//
// Parameters:
//   - size: number of messages per batch
//
// Returns:
//   - *publisherChannelAdapterBuilder: builder instance for chaining
func (b *publisherChannelAdapterBuilder) WithBatchSize(
	size int,
) *publisherChannelAdapterBuilder {
	b.batchSize = size
	return b
}

// WithBatchBytes sets the maximum size of a batch in bytes before sending.
// The producer sends when either batch size or batch bytes limit is reached.
//
// Parameters:
//   - bytes: maximum batch size in bytes
//
// Returns:
//   - *publisherChannelAdapterBuilder: builder instance for chaining
func (b *publisherChannelAdapterBuilder) WithBatchBytes(
	bytes int64,
) *publisherChannelAdapterBuilder {
	b.batchBytes = bytes
	return b
}

// WithAsync enables or disables asynchronous message sending.
// When enabled, Send returns immediately without waiting for broker acknowledgment.
//
// Parameters:
//   - async: whether to send messages asynchronously
//
// Returns:
//   - *publisherChannelAdapterBuilder: builder instance for chaining
func (b *publisherChannelAdapterBuilder) WithAsync(
	async bool,
) *publisherChannelAdapterBuilder {
	b.async = async
	return b
}

// WithRequiredAcks sets the required acknowledgments level.
// Higher levels ensure greater reliability but lower throughput.
//
// Parameters:
//   - acks: acknowledgment level (0=None, 1=Leader, -1=All)
//
// Returns:
//   - *publisherChannelAdapterBuilder: builder instance for chaining
func (b *publisherChannelAdapterBuilder) WithRequiredAcks(
	acks int,
) *publisherChannelAdapterBuilder {
	b.requiredAcks = acks
	return b
}

// Build constructs a Kafka outbound channel adapter from the dependency
// container. It retrieves the connection, creates a Kafka writer with the
// configured settings, and returns a wrapped outbound adapter.
//
// Parameters:
//   - container: dependency container containing required components
//
// Returns:
//   - endpoint.OutboundChannelAdapter: configured publisher channel
//   - error: error if connection not found or is invalid
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

	conn, ok := con.(*connection)
	if !ok {
		return nil, fmt.Errorf(
			"[kafka-outbound-channel] connection %s is not a valid Kafka connection",
			b.connectionReferenceName,
		)
	}

	producer := &kafka.Writer{
		Addr:         kafka.TCP(conn.getHost()...),
		Topic:        b.ChannelName(),
		Transport:    conn.getTransport(),
		MaxAttempts:  b.maxAttempts,
		BatchSize:    b.batchSize,
		BatchBytes:   b.batchBytes,
		Async:        b.async,
		RequiredAcks: kafka.RequiredAcks(b.requiredAcks),
	}

	adapter := NewOutboundChannelAdapter(
		producer,
		b.ChannelName(),
		b.MessageTranslator(),
	)

	return b.OutboundChannelAdapterBuilder.BuildOutboundAdapter(adapter)
}

// Name returns the topic name of the Kafka outbound channel adapter.
//
// Returns:
//   - string: the topic name
func (a *outboundChannelAdapter) Name() string {
	return a.topicName
}

// Send publishes a message to the Kafka topic with context support and
// distributed tracing. The message is translated to Kafka format and sent
// through the producer.
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
		err := fmt.Errorf(
			"[KAFKA OUTBOUND CHANNEL] Context cancelled after processing before sending result.",
		)
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
		err := fmt.Errorf(
			"[KAFKA OUTBOUND CHANNEL] Context cancelled after processing after sending result.",
		)
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

// Close closes the Kafka producer and releases associated resources.
//
// Returns:
//   - error: error if closing the producer fails
func (a *outboundChannelAdapter) Close() error {
	return a.producer.Close()
}
