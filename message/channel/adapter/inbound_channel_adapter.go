// Package adapter provides inbound channel adapters for message system integration.
//
// This package implements the Inbound Channel Adapter pattern from Enterprise Integration
// Patterns, facilitating the translation, processing, and routing of messages received
// from external sources to the internal domain format. It supports patterns such as
// Dead Letter Channel, interceptors, and extensibility for different protocols and
// sources.
//
// The builders and adapters defined here allow configuration of message processing
// pipelines, including error handling, dead letter processing, and integration with
// the system's messaging core.
package adapter

import (
	"context"
	"fmt"

	"github.com/hex-api-go/pkg/core/infrastructure/gomes/message"
	"github.com/hex-api-go/pkg/core/infrastructure/gomes/message/handler"
)

// InboundChannelMessageTranslator defines the contract for translating external messages
// to the internal format.
//
// T represents the external message type that needs to be translated.
type InboundChannelMessageTranslator[T any] interface {
	// ToMessage converts an external message to the internal message format.
	//
	// Parameters:
	//   - msg: The external message to be translated
	//
	// Returns:
	//   - *message.Message: The translated message in internal format
	ToMessage(msg T) (*message.Message, error)
}

// InboundChannelAdapterBuilder provides a fluent interface for configuring
// inbound channel adapters with various options like message translators,
// dead letter channels, and interceptors.
//
// TMessageType represents the external message type that will be received.
type InboundChannelAdapterBuilder[TMessageType any] struct {
	channelName           string
	messageTranslator     InboundChannelMessageTranslator[TMessageType]
	referenceName         string
	deadLetterChannelName string
	beforeProcessors      []message.MessageHandler
	afterProcessors       []message.MessageHandler
	retryTimeAttempts     []int
}

// InboundChannelAdapter handles the reception, processing, and forwarding of messages
// from external sources to the system core.
type InboundChannelAdapter struct {
	inboundAdapter        message.ConsumerChannel
	referenceName         string
	deadLetterChannelName string
	beforeProcessors      []message.MessageHandler
	afterProcessors       []message.MessageHandler
	retryTimeAttempts     []int
}

// NewInboundChannelAdapterBuilder creates a new builder instance for configuring
// inbound channel adapters.
//
// Parameters:
//   - referenceName: Unique identifier for the adapter instance
//   - channelName: Name of the logical channel
//   - messageTranslator: Translator for converting external messages to internal format
//
// Returns:
//   - *InboundChannelAdapterBuilder[T]: Configured builder instance
func NewInboundChannelAdapterBuilder[T any](
	referenceName string,
	channelName string,
	messageTranslator InboundChannelMessageTranslator[T],
) *InboundChannelAdapterBuilder[T] {
	return &InboundChannelAdapterBuilder[T]{
		channelName:       channelName,
		messageTranslator: messageTranslator,
		referenceName:     referenceName,
		beforeProcessors:  []message.MessageHandler{},
		afterProcessors:   []message.MessageHandler{},
	}
}

// NewInboundChannelAdapter creates a new inbound channel adapter instance.
//
// Parameters:
//   - adapter: The consumer channel implementation for receiving messages
//   - referenceName: Unique identifier for the adapter instance
//   - deadLetterChannelName: Name of the dead letter channel for failed messages
//   - beforeProcessors: List of pre-processing message handlers
//   - afterProcessors: List of post-processing message handlers
//   - retryTimeAttempts: time and number of retry attempts
//
// Returns:
//   - *InboundChannelAdapter: Configured inbound channel adapter
func NewInboundChannelAdapter(
	adapter message.ConsumerChannel,
	referenceName string,
	deadLetterChannelName string,
	beforeProcessors []message.MessageHandler,
	afterProcessors []message.MessageHandler,
	retryTimeAttempts []int,
) *InboundChannelAdapter {
	return &InboundChannelAdapter{
		inboundAdapter:        adapter,
		referenceName:         referenceName,
		deadLetterChannelName: deadLetterChannelName,
		beforeProcessors:      beforeProcessors,
		afterProcessors:       afterProcessors,
		retryTimeAttempts:     retryTimeAttempts,
	}
}

// WithDeadLetterChannelName sets the dead letter channel name for the adapter builder.
//
// Parameters:
//   - value: The dead letter channel name to set
func (b *InboundChannelAdapterBuilder[TMessageType]) WithDeadLetterChannelName(
	value string,
) {
	b.deadLetterChannelName = value
}

// WithBeforeInterceptors sets the before processing interceptors for the adapter builder.
//
// Parameters:
//   - processors: Variable number of message handlers to execute before processing
func (b *InboundChannelAdapterBuilder[TMessageType]) WithBeforeInterceptors(
	processors ...message.MessageHandler,
) {
	b.beforeProcessors = processors
}

// WithAfterInterceptors sets the after processing interceptors for the adapter builder.
//
// Parameters:
//   - processors: Variable number of message handlers to execute after processing
func (b *InboundChannelAdapterBuilder[TMessageType]) WithAfterInterceptors(
	processors ...message.MessageHandler,
) {
	b.afterProcessors = processors
}

// ReferenceName returns the current reference(ChannelName) name of the builder.
//
// Returns:
//   - string: The reference name
func (b *InboundChannelAdapterBuilder[TMessageType]) ReferenceName() string {
	return b.channelName
}

// WithRetryTimes Sets the time and number of retry attempts.
//
// Parameters:
//   - hitTimesMillisecond: retry time attempt
func (b *InboundChannelAdapterBuilder[TMessageType]) WithRetryTimes(
	hitTimesMillisecond ...int,
) {
	b.retryTimeAttempts = hitTimesMillisecond
}

func (
	b *InboundChannelAdapterBuilder[TMessageType],
) MessageTranslator() InboundChannelMessageTranslator[TMessageType] {
	return b.messageTranslator
}

// BuildInboundAdapter creates a configured inbound channel adapter from the builder
// settings.
//
// Parameters:
//   - inboundAdapter: The consumer channel implementation for receiving messages
//
// Returns:
//   - *InboundChannelAdapter: Configured inbound channel adapter
func (b *InboundChannelAdapterBuilder[TMessageType]) BuildInboundAdapter(
	inboundAdapter message.ConsumerChannel,
) *InboundChannelAdapter {
	return NewInboundChannelAdapter(
		inboundAdapter,
		b.referenceName,
		b.deadLetterChannelName,
		b.beforeProcessors,
		b.afterProcessors,
		b.retryTimeAttempts,
	)
}

// ReferenceName returns the reference name of the adapter.
//
// Returns:
//   - string: The reference name
func (i *InboundChannelAdapter) ReferenceName() string {
	return i.referenceName
}

// DeadLetterChannelName returns the configured dead letter channel name.
//
// Returns:
//   - string: The dead letter channel name
func (i *InboundChannelAdapter) DeadLetterChannelName() string {
	return i.deadLetterChannelName
}

// BeforeProcessors returns the configured pre-processing handlers.
//
// Returns:
//   - []message.MessageHandler: List of pre-processing handlers
func (i *InboundChannelAdapter) BeforeProcessors() []message.MessageHandler {
	return i.beforeProcessors
}

// AfterProcessors returns the configured post-processing handlers.
//
// Returns:
//   - []message.MessageHandler: List of post-processing handlers
func (i *InboundChannelAdapter) AfterProcessors() []message.MessageHandler {
	return i.afterProcessors
}

// RetryAttempts returns the configured retry time and attempts.
//
// Returns:
//   - []int: List of time attempts
func (i *InboundChannelAdapter) RetryAttempts() []int {
	return i.retryTimeAttempts
}

// ReceiveMessage receives a message from the channel, respecting context cancellation.
//
// Parameters:
//   - ctx: Context for timeout/cancellation control
//
// Returns:
//   - *message.Message: The received message from the channel
//   - error: Error if the context is cancelled or reception fails
func (i *InboundChannelAdapter) ReceiveMessage(ctx context.Context) (*message.Message, error) {
	select {
	case <-ctx.Done():
		return nil, fmt.Errorf(
			"[inbound-channel] Context cancelled after processing, before sending result",
		)
	default:
	}
	return i.inboundAdapter.Receive(ctx)
}

// Close closes the inbound channel adapter, releasing associated resources.
//
// Returns:
//   - error: Error if closing the channel fails
func (i *InboundChannelAdapter) Close() error {
	return i.inboundAdapter.Close()
}

func (i *InboundChannelAdapter) CommitMessage(msg *message.Message) error {
	ackChannel, ok := i.inboundAdapter.(handler.ChannelMessageAcknowledgment)
	if !ok {
		return nil
	}

	return ackChannel.CommitMessage(msg)
}
