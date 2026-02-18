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
)

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
}

// OutboundChannelAdapter handles the sending of messages to external systems
// through configured publisher channels.
type OutboundChannelAdapter struct {
	outboundAdapter  message.PublisherChannel
	replyChannelName string
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
	}
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
) (*OutboundChannelAdapter, error) {

	outboundHandler := NewOutboundChannelAdapter(outboundAdapter, b.replyChannelName)
	return outboundHandler, nil
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
	replyChannelName string,
) *OutboundChannelAdapter {
	return &OutboundChannelAdapter{
		outboundAdapter:  adapter,
		replyChannelName: replyChannelName,
	}
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
func (o *OutboundChannelAdapter) Send(
	ctx context.Context,
	msg *message.Message,
) error {

	if o.replyChannelName != "" {
		msg.GetHeader().Set(message.HeaderReplyTo, o.replyChannelName)
	}
	err := o.outboundAdapter.Send(ctx, msg)
	if msg.GetInternalReplyChannel() != nil {
		go o.publishOnInternalChannel(ctx, msg, err)
	}

	if err != nil {
		return err
	}

	return nil
}

// Name returns the name of outbound channel adapter.
//
// Returns:
//   - string: the topic name
func (o *OutboundChannelAdapter) Name() string {
	return o.outboundAdapter.Name()
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
		WithReplyTo("").
		WithChannelName("").
		Build()
	msg.GetInternalReplyChannel().Send(ctx, resultMessage)
}

// Close closes the outbound channel adapter, releasing associated resources.
//
// Returns:
//   - error: Error if closing the channel fails
func (o *OutboundChannelAdapter) Close() error {

	closableChannel, ok := o.outboundAdapter.(ClosableChannel)
	if !ok {
		return nil
	}
	return closableChannel.Close()
}
