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

	"github.com/jeffersonbrasilino/gomes/container"
	"github.com/jeffersonbrasilino/gomes/message"
	"github.com/jeffersonbrasilino/gomes/message/channel/adapter"
	"github.com/jeffersonbrasilino/gomes/message/endpoint"
	"github.com/segmentio/kafka-go"
)

// consumerChannelAdapterBuilder provides a builder pattern for creating
// Kafka inbound channel adapters with connection and topic configuration.
type consumerChannelAdapterBuilder struct {
	*adapter.InboundChannelAdapterBuilder[*kafka.Message]
	connectionReferenceName string
	consumerName            string
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
	}
	return builder
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
) (endpoint.InboundChannelAdapter, error) {
	con, err := container.Get(c.connectionReferenceName)

	if err != nil {
		return nil, fmt.Errorf(
			"[kafka-inbound-channel] connection %s does not exist",
			c.connectionReferenceName,
		)
	}

	consumer := con.(*connection).Consumer(c.ReferenceName(), fmt.Sprintf("%s:%s", c.connectionReferenceName, c.consumerName))
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
// This method runs in a separate goroutine and continuously polls for messages.
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

func (a *inboundChannelAdapter) CommitMessage(msg *message.Message) error {
	if segmentioMessage, ok := msg.GetRawMessage().(*kafka.Message); ok {
		return a.consumer.CommitMessages(a.ctx, *segmentioMessage)
	}
	return fmt.Errorf("[kafka-inbound-channel] failed to commit message")
}
