// Package kafka provides Kafka integration for the message system.
//
// This package implements Kafka-specific channel adapters and connections for
// publishing and consuming messages through Apache Kafka. It provides outbound
// and inbound channel adapters with message translation capabilities.
//
// The InboundChannelAdapter implementation supports:
// - Kafka consumer integration for message consumption
// - Message translation between Kafka and internal formats
// - Asynchronous message processing with context support
// - Graceful shutdown and resource cleanup
package kafka

import (
	"context"
	"fmt"
	"time"

	"github.com/jeffersonbrasilino/gomes/container"
	"github.com/jeffersonbrasilino/gomes/message"
	"github.com/jeffersonbrasilino/gomes/message/adapter"
	"github.com/jeffersonbrasilino/gomes/otel"
	"github.com/segmentio/kafka-go"
)

// consumerChannelAdapterBuilder provides a builder pattern for creating
// Kafka inbound channel adapters with connection and topic configuration.
type consumerChannelAdapterBuilder struct {
	*adapter.InboundChannelAdapterBuilder[*kafka.Message]
	connectionReferenceName string
	consumerName            string
	kafkaConsumerConfig     *kafka.ReaderConfig
}

// inboundChannelAdapter implements the InboundChannelAdapter interface for Kafka,
// providing message consumption capabilities through a Kafka consumer.
type inboundChannelAdapter struct {
	consumer          *kafka.Reader
	topic             string
	messageTranslator adapter.InboundChannelMessageTranslator[*kafka.Message]
	messageChannel    chan *message.Message
	errorChannel      chan error
	ctx               context.Context
	cancelCtx         context.CancelFunc
	otelTrace         otel.OtelTrace
}

// NewConsumerChannelAdapterBuilder creates a new Kafka consumer channel
// adapter builder instance.
//
// Parameters:
//   - connectionReferenceName: reference name for the Kafka connection
//   - topicName: the Kafka topic to consume messages from
//   - consumerName: the consumer group name
//
// Returns:
//   - *consumerChannelAdapterBuilder: configured builder instance
func NewConsumerChannelAdapterBuilder(
	connectionReferenceName string,
	topicName string,
	consumerName string,
) *consumerChannelAdapterBuilder {
	builder := &consumerChannelAdapterBuilder{
		adapter.NewInboundChannelAdapterBuilder(
			consumerName,
			topicName,
			NewMessageTranslator(),
		),
		connectionReferenceName,
		consumerName,
		&kafka.ReaderConfig{},
	}
	return builder
}

// WithGroupTopics sets the group topics for the Kafka consumer.
// This allows the consumer to subscribe to multiple topics at once.
//
// Parameters:
//   - groupTopics: list of topics to subscribe to
//
// Returns:
//   - *consumerChannelAdapterBuilder: builder instance for chaining
func (b *consumerChannelAdapterBuilder) WithGroupTopics(
	groupTopics []string,
) *consumerChannelAdapterBuilder {
	b.kafkaConsumerConfig.GroupTopics = groupTopics
	return b
}

// WithPartition sets the partition for the Kafka consumer.
// When specified, the consumer will only consume from the given partition.
//
// Parameters:
//   - partition: partition number to consume from
//
// Returns:
//   - *consumerChannelAdapterBuilder: builder instance for chaining
func (b *consumerChannelAdapterBuilder) WithPartition(
	partition int,
) *consumerChannelAdapterBuilder {
	b.kafkaConsumerConfig.Partition = partition
	return b
}

// WithQueueCapacity sets the queue capacity for the Kafka consumer.
// This controls the buffer size for fetch requests.
//
// Parameters:
//   - queueCapacity: queue capacity in bytes
//
// Returns:
//   - *consumerChannelAdapterBuilder: builder instance for chaining
func (b *consumerChannelAdapterBuilder) WithQueueCapacity(
	queueCapacity int,
) *consumerChannelAdapterBuilder {
	b.kafkaConsumerConfig.QueueCapacity = queueCapacity
	return b
}

// WithMinBytes sets the minimum bytes for the Kafka consumer.
// The consumer will wait for at least this many bytes from the broker.
//
// Parameters:
//   - minBytes: minimum number of bytes to fetch
//
// Returns:
//   - *consumerChannelAdapterBuilder: builder instance for chaining
func (b *consumerChannelAdapterBuilder) WithMinBytes(
	minBytes int,
) *consumerChannelAdapterBuilder {
	b.kafkaConsumerConfig.MinBytes = minBytes
	return b
}

// WithMaxBytes sets the maximum bytes for the Kafka consumer.
// The consumer will not fetch more than this many bytes in a single response.
//
// Parameters:
//   - maxBytes: maximum number of bytes to fetch
//
// Returns:
//   - *consumerChannelAdapterBuilder: builder instance for chaining
func (b *consumerChannelAdapterBuilder) WithMaxBytes(
	maxBytes int,
) *consumerChannelAdapterBuilder {
	b.kafkaConsumerConfig.MaxBytes = maxBytes
	return b
}

// WithMaxWait sets the maximum wait time for the Kafka consumer.
// The broker will not wait longer than this for the minimum bytes to be
// available.
//
// Parameters:
//   - maxWait: maximum duration to wait for data
//
// Returns:
//   - *consumerChannelAdapterBuilder: builder instance for chaining
func (b *consumerChannelAdapterBuilder) WithMaxWait(
	maxWait time.Duration,
) *consumerChannelAdapterBuilder {
	b.kafkaConsumerConfig.MaxWait = maxWait
	return b
}

// WithReadBatchTimeout sets the read batch timeout for the Kafka consumer.
// This timeout controls how long to wait when reading a batch of messages.
//
// Parameters:
//   - readBatchTimeout: timeout duration for reading batches
//
// Returns:
//   - *consumerChannelAdapterBuilder: builder instance for chaining
func (b *consumerChannelAdapterBuilder) WithReadBatchTimeout(
	readBatchTimeout time.Duration,
) *consumerChannelAdapterBuilder {
	b.kafkaConsumerConfig.ReadBatchTimeout = readBatchTimeout
	return b
}

// WithReadLagInterval sets the read lag interval for the Kafka consumer.
// This controls how often lag statistics are updated.
//
// Parameters:
//   - readLagInterval: interval for updating lag statistics
//
// Returns:
//   - *consumerChannelAdapterBuilder: builder instance for chaining
func (b *consumerChannelAdapterBuilder) WithReadLagInterval(
	readLagInterval time.Duration,
) *consumerChannelAdapterBuilder {
	b.kafkaConsumerConfig.ReadLagInterval = readLagInterval
	return b
}

// WithHeartbeatInterval sets the heartbeat interval for the Kafka consumer.
// This controls how often heartbeats are sent to maintain group membership.
//
// Parameters:
//   - heartbeatInterval: interval between heartbeats
//
// Returns:
//   - *consumerChannelAdapterBuilder: builder instance for chaining
func (b *consumerChannelAdapterBuilder) WithHeartbeatInterval(
	heartbeatInterval time.Duration,
) *consumerChannelAdapterBuilder {
	b.kafkaConsumerConfig.HeartbeatInterval = heartbeatInterval
	return b
}

// WithCommitInterval sets the commit interval for the Kafka consumer.
// This controls how often offsets are committed to Kafka.
//
// Parameters:
//   - commitInterval: interval between offset commits
//
// Returns:
//   - *consumerChannelAdapterBuilder: builder instance for chaining
func (b *consumerChannelAdapterBuilder) WithCommitInterval(
	commitInterval time.Duration,
) *consumerChannelAdapterBuilder {
	b.kafkaConsumerConfig.CommitInterval = commitInterval
	return b
}

// WithPartitionWatchInterval sets the partition watch interval for the Kafka
// consumer. This controls how often partition assignments are refreshed.
//
// Parameters:
//   - partitionWatchInterval: interval for checking partition changes
//
// Returns:
//   - *consumerChannelAdapterBuilder: builder instance for chaining
func (b *consumerChannelAdapterBuilder) WithPartitionWatchInterval(
	partitionWatchInterval time.Duration,
) *consumerChannelAdapterBuilder {
	b.kafkaConsumerConfig.PartitionWatchInterval = partitionWatchInterval
	return b
}

// WithWatchPartitionChanges configures whether to watch for partition changes.
// When enabled, the consumer will rebalance when partitions are added or removed.
//
// Parameters:
//   - watchPartitionChanges: whether to monitor partition changes
//
// Returns:
//   - *consumerChannelAdapterBuilder: builder instance for chaining
func (b *consumerChannelAdapterBuilder) WithWatchPartitionChanges(
	watchPartitionChanges bool,
) *consumerChannelAdapterBuilder {
	b.kafkaConsumerConfig.WatchPartitionChanges = watchPartitionChanges
	return b
}

// WithSessionTimeout sets the session timeout for the Kafka consumer.
// If no heartbeats are received within this time, the consumer is removed from
// the group.
//
// Parameters:
//   - sessionTimeout: timeout for session maintenance
//
// Returns:
//   - *consumerChannelAdapterBuilder: builder instance for chaining
func (b *consumerChannelAdapterBuilder) WithSessionTimeout(
	sessionTimeout time.Duration,
) *consumerChannelAdapterBuilder {
	b.kafkaConsumerConfig.SessionTimeout = sessionTimeout
	return b
}

// WithRebalanceTimeout sets the rebalance timeout for the Kafka consumer.
// This is the maximum time allowed for a rebalance operation to complete.
//
// Parameters:
//   - rebalanceTimeout: timeout for rebalance operations
//
// Returns:
//   - *consumerChannelAdapterBuilder: builder instance for chaining
func (b *consumerChannelAdapterBuilder) WithRebalanceTimeout(
	rebalanceTimeout time.Duration,
) *consumerChannelAdapterBuilder {
	b.kafkaConsumerConfig.RebalanceTimeout = rebalanceTimeout
	return b
}

// WithJoinGroupBackoff sets the join group backoff for the Kafka consumer.
// This controls the initial backoff time for retrying group joins.
//
// Parameters:
//   - joinGroupBackoff: backoff duration for group join retries
//
// Returns:
//   - *consumerChannelAdapterBuilder: builder instance for chaining
func (b *consumerChannelAdapterBuilder) WithJoinGroupBackoff(
	joinGroupBackoff time.Duration,
) *consumerChannelAdapterBuilder {
	b.kafkaConsumerConfig.JoinGroupBackoff = joinGroupBackoff
	return b
}

// WithRetentionTime sets the retention time for the Kafka consumer.
// Offsets will be discarded after this duration of inactivity.
//
// Parameters:
//   - retentionTime: duration for offset retention
//
// Returns:
//   - *consumerChannelAdapterBuilder: builder instance for chaining
func (b *consumerChannelAdapterBuilder) WithRetentionTime(
	retentionTime time.Duration,
) *consumerChannelAdapterBuilder {
	b.kafkaConsumerConfig.RetentionTime = retentionTime
	return b
}

// WithStartOffset sets the start offset for the Kafka consumer.
// This determines where the consumer begins reading from the topic.
//
// Parameters:
//   - startOffset: offset to start consuming from
//
// Returns:
//   - *consumerChannelAdapterBuilder: builder instance for chaining
func (b *consumerChannelAdapterBuilder) WithStartOffset(
	startOffset int64,
) *consumerChannelAdapterBuilder {
	b.kafkaConsumerConfig.StartOffset = startOffset
	return b
}

// WithReadBackoffMin sets the minimum read backoff for the Kafka consumer.
// This is the initial backoff time when read operations fail.
//
// Parameters:
//   - readBackoffMin: minimum backoff duration
//
// Returns:
//   - *consumerChannelAdapterBuilder: builder instance for chaining
func (b *consumerChannelAdapterBuilder) WithReadBackoffMin(
	readBackoffMin time.Duration,
) *consumerChannelAdapterBuilder {
	b.kafkaConsumerConfig.ReadBackoffMin = readBackoffMin
	return b
}

// WithReadBackoffMax sets the maximum read backoff for the Kafka consumer.
// This is the maximum backoff time when read operations fail.
//
// Parameters:
//   - readBackoffMax: maximum backoff duration
//
// Returns:
//   - *consumerChannelAdapterBuilder: builder instance for chaining
func (b *consumerChannelAdapterBuilder) WithReadBackoffMax(
	readBackoffMax time.Duration,
) *consumerChannelAdapterBuilder {
	b.kafkaConsumerConfig.ReadBackoffMax = readBackoffMax
	return b
}

// WithIsolationLevel sets the isolation level for the Kafka consumer.
// This determines which uncommitted messages are visible to the consumer.
//
// Parameters:
//   - isolationLevel: isolation level (0=ReadUncommitted, 1=ReadCommitted)
//
// Returns:
//   - *consumerChannelAdapterBuilder: builder instance for chaining
func (b *consumerChannelAdapterBuilder) WithIsolationLevel(
	isolationLevel int8,
) *consumerChannelAdapterBuilder {
	b.kafkaConsumerConfig.IsolationLevel = kafka.IsolationLevel(isolationLevel)
	return b
}

// WithMaxAttempts sets the maximum attempts for the Kafka consumer.
// Failed requests will be retried up to this many times.
//
// Parameters:
//   - maxAttempts: maximum number of retry attempts
//
// Returns:
//   - *consumerChannelAdapterBuilder: builder instance for chaining
func (b *consumerChannelAdapterBuilder) WithMaxAttempts(
	maxAttempts int,
) *consumerChannelAdapterBuilder {
	b.kafkaConsumerConfig.MaxAttempts = maxAttempts
	return b
}

// WithOffsetOutOfRangeError configures whether to return an error when the
// offset is out of range. If false, the consumer resets to the first or last
// available offset.
//
// Parameters:
//   - offsetOutOfRangeError: whether to error on out-of-range offsets
//
// Returns:
//   - *consumerChannelAdapterBuilder: builder instance for chaining
func (b *consumerChannelAdapterBuilder) WithOffsetOutOfRangeError(
	offsetOutOfRangeError bool,
) *consumerChannelAdapterBuilder {
	b.kafkaConsumerConfig.OffsetOutOfRangeError = offsetOutOfRangeError
	return b
}

// NewInboundChannelAdapter creates a new Kafka inbound channel adapter instance.
//
// Parameters:
//   - consumer: the Kafka consumer for receiving messages
//   - topic: the Kafka topic name
//   - messageTranslator: translator for converting Kafka messages to internal format
//
// Returns:
//   - *inboundChannelAdapter: configured inbound channel adapter
func NewInboundChannelAdapter(
	consumer *kafka.Reader,
	topic string,
	messageTranslator adapter.InboundChannelMessageTranslator[*kafka.Message],
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
		otelTrace:         otel.InitTrace("kafka-inbound-channel-adapter"),
	}
	go adp.subscribeOnTopic()
	return adp
}

// Build constructs a Kafka inbound channel adapter from the dependency container.
//
// Parameters:
//   - container: dependency container containing required components
//
// Returns:
//   - message.InboundChannelAdapter: configured inbound channel adapter
//   - error: error if construction fails
func (c *consumerChannelAdapterBuilder) Build(
	container container.Container[any, any],
) (*adapter.InboundChannelAdapter, error) {
	con, err := container.Get(c.connectionReferenceName)

	if err != nil {
		return nil, fmt.Errorf(
			"[kafka-inbound-channel] connection %s does not exist",
			c.connectionReferenceName,
		)
	}

	conn, ok := con.(*connection)
	if !ok {
		return nil, fmt.Errorf(
			"[kafka-outbound-channel] connection %s is not a valid Kafka connection",
			c.connectionReferenceName,
		)
	}
	c.kafkaConsumerConfig.Brokers = conn.getHost()
	c.kafkaConsumerConfig.Topic = c.ReferenceName()
	c.kafkaConsumerConfig.GroupID = fmt.Sprintf("%s:%s", c.connectionReferenceName, c.consumerName)
	c.kafkaConsumerConfig.Dialer = conn.getDialer()

	consumer := kafka.NewReader(*c.kafkaConsumerConfig)
	adapter := NewInboundChannelAdapter(consumer, c.ReferenceName(), c.MessageTranslator())
	return c.InboundChannelAdapterBuilder.BuildInboundAdapter(adapter), nil
}

// Name returns the topic name of the Kafka inbound channel adapter.
//
// Returns:
//   - string: the topic name
func (a *inboundChannelAdapter) Name() string {
	return a.topic
}

// Receive receives a message from the Kafka topic.
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

// Close gracefully closes the Kafka inbound channel adapter and stops
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

// subscribeOnTopic subscribes to the Kafka topic and processes incoming messages.
// This method runs in a separate goroutine and continuously polls for messages,
// translating them to the internal message format and sending them to the
// message channel.
func (a *inboundChannelAdapter) subscribeOnTopic() {
	for {
		select {
		case <-a.ctx.Done():
			return
		default:
		}
		msg, err := a.consumer.FetchMessage(a.ctx)

		if err != nil {
			if err == context.Canceled {
				return
			}
			a.errorChannel <- err
		}

		message, translateErr := a.messageTranslator.ToMessage(&msg)

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

// CommitMessage commits the Kafka message offset to the broker, marking it as
// consumed.
//
// Parameters:
//   - msg: the internal message whose offset should be committed
//
// Returns:
//   - error: error if the message is not a Kafka message or commit fails
func (a *inboundChannelAdapter) CommitMessage(msg *message.Message) error {
	if segmentioMessage, ok := msg.GetRawMessage().(*kafka.Message); ok {
		return a.consumer.CommitMessages(a.ctx, *segmentioMessage)
	}
	return fmt.Errorf("[kafka-inbound-channel] failed to commit message")
}
