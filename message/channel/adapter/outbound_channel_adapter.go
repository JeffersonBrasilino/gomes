// Package adapter provides outbound channel adapters for message system integration.
//
// This package implements the Outbound Channel Adapter pattern from Enterprise Integration
// Patterns, allowing applications to send messages to external systems through various
// transport mechanisms. It provides a builder pattern for configuring adapters with
// message translators, interceptors, and channel configurations.
//
// The main components include:
//   - OutboundChannelMessageTranslator: Interface for translating internal messages to
//     external format
//   - OutboundChannelAdapterBuilder: Builder for configuring outbound channel adapters
//   - OutboundChannelAdapter: Concrete implementation of outbound message handling
package adapter

import (
	"context"

	"github.com/jeffersonbrasilino/gomes/message"
	"github.com/jeffersonbrasilino/gomes/message/channel"
)

// OutboundChannelMessageTranslator defines the contract for translating internal messages
// to external system format.
//
// T represents the target message type for the external system.
type OutboundChannelMessageTranslator[T any] interface {
	// FromMessage converts an internal message to the target external format.
	//
	// Parameters:
	//   - msg: The internal message to be translated
	//
	// Returns:
	//   - T: The translated message in external format
	FromMessage(msg *message.Message) (T, error)
}

// OutboundChannelAdapterBuilder provides a fluent interface for configuring
// outbound channel adapters with various options like message translators,
// interceptors, and channel configurations.
//
// TMessageType represents the target message type for external systems.
type OutboundChannelAdapterBuilder[TMessageType any] struct {
	referenceName     string
	channelName       string
	replyChannelName  string
	messageTranslator OutboundChannelMessageTranslator[TMessageType]
	beforeProcessors  []message.MessageHandler
	afterProcessors   []message.MessageHandler
}

// OutboundChannelAdapter handles the sending of messages to external systems
// through configured publisher channels.
type OutboundChannelAdapter struct {
	outboundAdapter message.PublisherChannel
}

// NewOutboundChannelAdapterBuilder creates a new builder instance for configuring
// outbound channel adapters.
//
// Parameters:
//   - referenceName: Unique identifier for the adapter instance
//   - channelName: Name of the channel to publish messages to
//   - messageTranslator: Translator for converting internal messages to external format
//
// Returns:
//   - *OutboundChannelAdapterBuilder[T]: Configured builder instance
func NewOutboundChannelAdapterBuilder[T any](
	referenceName string,
	channelName string,
	messageTranslator OutboundChannelMessageTranslator[T],
) *OutboundChannelAdapterBuilder[T] {
	return &OutboundChannelAdapterBuilder[T]{
		referenceName:     referenceName,
		channelName:       channelName,
		messageTranslator: messageTranslator,
		beforeProcessors:  []message.MessageHandler{},
		afterProcessors:   []message.MessageHandler{},
	}
}

// NewOutboundChannelAdapter creates a new outbound channel adapter instance.
//
// Parameters:
//   - adapter: The publisher channel implementation for sending messages
//
// Returns:
//   - *OutboundChannelAdapter: Configured outbound channel adapter
func NewOutboundChannelAdapter(
	adapter message.PublisherChannel,
) *OutboundChannelAdapter {
	return &OutboundChannelAdapter{
		outboundAdapter: adapter,
	}
}

// WithReferenceName sets the reference name for the adapter builder.
//
// Parameters:
//   - value: The reference name to set
//
// Returns:
//   - *OutboundChannelAdapterBuilder[TMessageType]: Builder instance for method chaining
func (b *OutboundChannelAdapterBuilder[TMessageType]) WithReferenceName(
	value string,
) *OutboundChannelAdapterBuilder[TMessageType] {
	b.referenceName = value
	return b
}

// WithChannelName sets the channel name for the adapter builder.
//
// Parameters:
//   - value: The channel name to set
func (b *OutboundChannelAdapterBuilder[TMessageType]) WithChannelName(
	value string,
) *OutboundChannelAdapterBuilder[TMessageType] {
	b.channelName = value
	return b
}

// WithMessageTranslator sets the message translator for the adapter builder.
//
// Parameters:
//   - transator: The message translator to use for converting messages
func (b *OutboundChannelAdapterBuilder[TMessageType]) WithMessageTranslator(
	transator OutboundChannelMessageTranslator[TMessageType],
) *OutboundChannelAdapterBuilder[TMessageType] {
	b.messageTranslator = transator
	return b
}

// WithReplyChannelName sets the reply channel name for the adapter builder.
//
// Parameters:
//   - value: The reply channel name to set
func (b *OutboundChannelAdapterBuilder[TMessageType]) WithReplyChannelName(
	value string,
) *OutboundChannelAdapterBuilder[TMessageType] {
	b.replyChannelName = value
	return b
}

// WithBeforeInterceptors sets the before processing interceptors for the adapter builder.
//
// Parameters:
//   - processors: Variable number of message handlers to execute before processing
func (b *OutboundChannelAdapterBuilder[TMessageType]) WithBeforeInterceptors(
	processors ...message.MessageHandler,
) *OutboundChannelAdapterBuilder[TMessageType] {
	b.beforeProcessors = processors
	return b
}

// WithAfterInterceptors sets the after processing interceptors for the adapter builder.
//
// Parameters:
//   - processors: Variable number of message handlers to execute after processing
func (b *OutboundChannelAdapterBuilder[TMessageType]) WithAfterInterceptors(
	processors ...message.MessageHandler,
) *OutboundChannelAdapterBuilder[TMessageType] {
	b.afterProcessors = processors
	return b
}

// ReferenceName returns the current reference name of the builder.
//
// Returns:
//   - string: The reference name
func (b *OutboundChannelAdapterBuilder[TMessageType]) ReferenceName() string {
	return b.referenceName
}

// ChannelName returns the current channel name of the builder.
//
// Returns:
//   - string: The channel name
func (b *OutboundChannelAdapterBuilder[TMessageType]) ChannelName() string {
	return b.channelName
}

// ReplyChannelName returns the current reply channel name of the builder.
//
// Parameters:
//   - value: Unused parameter (consider removing in future refactoring)
//
// Returns:
//   - string: The reply channel name
func (b *OutboundChannelAdapterBuilder[TMessageType]) ReplyChannelName(
	value string,
) string {
	return b.replyChannelName
}

// MessageTranslator returns the current message translator of the builder.
//
// Returns:
//   - OutboundChannelMessageTranslator[TMessageType]: The message translator
func (
	b *OutboundChannelAdapterBuilder[TMessageType],
) MessageTranslator() OutboundChannelMessageTranslator[TMessageType] {
	return b.messageTranslator
}

// BuildOutboundAdapter creates a configured point-to-point channel with the outbound
// adapter.
//
// Parameters:
//   - outboundAdapter: The publisher channel implementation for sending messages
//
// Returns:
//   - *channel.PointToPointChannel: Configured point-to-point channel
//   - error: Any error that occurred during channel creation
func (b *OutboundChannelAdapterBuilder[TMessageType]) BuildOutboundAdapter(
	outboundAdapter message.PublisherChannel,
) (*channel.PointToPointChannel, error) {

	outboundHandler := NewOutboundChannelAdapter(outboundAdapter)

	chn := channel.NewPointToPointChannel(b.referenceName)
	chn.Subscribe(func(msg *message.Message) {
		outboundHandler.Handle(msg.GetContext(), msg)
	})

	return chn, nil
}

// Handle processes an outbound message by sending it through the configured publisher
// channel. If the message has a reply channel configured, it will publish the result
// to that channel.
//
// Parameters:
//   - ctx: Context for the operation
//   - msg: The message to be sent
//
// Returns:
//   - *message.Message: The processed message (same as input if successful)
//   - error: Any error that occurred during message processing
func (o *OutboundChannelAdapter) Handle(
	ctx context.Context,
	msg *message.Message,
) (*message.Message, error) {
	err := o.outboundAdapter.Send(ctx, msg)
	if msg.GetHeaders().ReplyChannel != nil {
		o.publishOnInternalChannel(ctx, msg, err)
	}

	if err != nil {
		return nil, err
	}

	return msg, nil
}

// publishOnInternalChannel publishes a result message to the configured reply channel.
// This method is used internally to send processing results back to the requesting
// system.
//
// Parameters:
//   - ctx: Context for the operation
//   - msg: The original message that was processed
//   - response: The response or error result from processing
func (o *OutboundChannelAdapter) publishOnInternalChannel(
	ctx context.Context,
	msg *message.Message,
	response any,
) {
	payloadMessage := msg.GetPayload()
	if response != nil {
		payloadMessage = response
	}
	resultMessage := message.NewMessageBuilderFromMessage(msg).
		WithMessageType(message.Document).
		WithPayload(payloadMessage).
		Build()
	msg.GetHeaders().ReplyChannel.Send(ctx, resultMessage)
}
