// Package message provides message building and construction utilities for the
// message system.
//
// This package implements the Message Builder pattern, enabling fluent construction
// of messages with various configurations including payload, headers, routing
// information, and context. It provides a type-safe and expressive way to create
// messages for the messaging system.
//
// The MessageBuilder implementation supports:
// - Fluent builder pattern for message construction
// - Comprehensive header configuration
// - Context-aware message creation
// - Reply channel and routing setup
// - Custom header management
package message

import (
	"context"
	"time"
)

// MessageBuilder provides a fluent interface for constructing messages with
// various configurations including payload, headers, routing, and context.
type MessageBuilder struct {
	payload          any
	origin           string
	messageId        string
	route            string
	messageType      MessageType
	replyChannel     PublisherChannel
	customHeaders    CustomHeaders
	correlationId    string
	channelName      string
	replyChannelName string
	timestamp        time.Time
	context          context.Context
	version          string
	rawMessage       any
}

// NewMessageBuilder creates a new message builder instance.
//
// Returns:
//   - *MessageBuilder: new message builder instance
func NewMessageBuilder() *MessageBuilder {
	return &MessageBuilder{}
}

// NewMessageBuilderFromMessage creates a new message builder instance from an
// existing message, copying all its properties.
//
// Parameters:
//   - msg: the source message to copy properties from
//
// Returns:
//   - *MessageBuilder: new message builder with copied properties
func NewMessageBuilderFromMessage(msg *Message) *MessageBuilder {
	return &MessageBuilder{
		origin:        msg.GetHeaders().Origin,
		messageId:     msg.GetHeaders().MessageId,
		payload:       msg.GetPayload(),
		route:         msg.GetHeaders().Route,
		messageType:   msg.GetHeaders().MessageType,
		replyChannel:  msg.GetHeaders().ReplyChannel,
		customHeaders: msg.GetHeaders().CustomHeaders,
		correlationId: msg.GetHeaders().CorrelationId,
		channelName:   msg.GetHeaders().ChannelName,
		timestamp:     msg.GetHeaders().Timestamp,
		context:       msg.GetContext(),
		version:       msg.GetHeaders().Version,
		rawMessage:    msg.GetRawMessage(),
	}
}

// WithPayload sets the message payload.
//
// Parameters:
//   - payload: the data to be carried by the message
//
// Returns:
//   - *MessageBuilder: builder instance for method chaining
func (b *MessageBuilder) WithPayload(payload any) *MessageBuilder {
	b.payload = payload
	return b
}

// WithMessageType sets the message type.
//
// Parameters:
//   - typeMessage: the type of the message
//
// Returns:
//   - *MessageBuilder: builder instance for method chaining
func (b *MessageBuilder) WithMessageType(typeMessage MessageType) *MessageBuilder {
	b.messageType = typeMessage
	return b
}

// WithRoute sets the message route for routing purposes.
//
// Parameters:
//   - route: the route identifier for the message
//
// Returns:
//   - *MessageBuilder: builder instance for method chaining
func (b *MessageBuilder) WithRoute(route string) *MessageBuilder {
	b.route = route
	return b
}

// WithReplyChannel sets the reply channel for request-response patterns.
//
// Parameters:
//   - value: the publisher channel for reply messages
//
// Returns:
//   - *MessageBuilder: builder instance for method chaining
func (b *MessageBuilder) WithReplyChannel(value PublisherChannel) *MessageBuilder {
	b.replyChannel = value
	return b
}

// WithCustomHeader sets custom headers for the message.
//
// Parameters:
//   - value: the custom headers to be included in the message
//
// Returns:
//   - *MessageBuilder: builder instance for method chaining
func (b *MessageBuilder) WithCustomHeader(value CustomHeaders) *MessageBuilder {
	b.customHeaders = value
	return b
}

// WithCorrelationId sets the correlation ID for message tracking.
//
// Parameters:
//   - value: the correlation identifier
//
// Returns:
//   - *MessageBuilder: builder instance for method chaining
func (b *MessageBuilder) WithCorrelationId(value string) *MessageBuilder {
	b.correlationId = value
	return b
}

// WithChannelName sets the channel name for message routing.
//
// Parameters:
//   - value: the name of the target channel
//
// Returns:
//   - *MessageBuilder: builder instance for method chaining
func (b *MessageBuilder) WithChannelName(value string) *MessageBuilder {
	b.channelName = value
	return b
}

// WithReplyChannelName sets the reply channel name for request-response patterns.
//
// Parameters:
//   - value: the name of the reply channel
//
// Returns:
//   - *MessageBuilder: builder instance for method chaining
func (b *MessageBuilder) WithReplyChannelName(value string) *MessageBuilder {
	b.replyChannelName = value
	return b
}

// WithContext sets the context for the message.
//
// Parameters:
//   - value: the context for timeout/cancellation control
//
// Returns:
//   - *MessageBuilder: builder instance for method chaining
func (b *MessageBuilder) WithContext(value context.Context) *MessageBuilder {
	b.context = value
	return b
}

// WithMessageId sets the message ID for the message being built.
//
// Parameters:
//   - value: the unique identifier for the message
//
// Returns:
//   - *MessageBuilder: builder instance for method chaining
func (b *MessageBuilder) WithMessageId(value string) *MessageBuilder {
	b.messageId = value
	return b
}

// WithTimestamp sets the timestamp for the message being built.
//
// Parameters:
//   - value: the timestamp to be set for the message
//
// Returns:
//   - *MessageBuilder: builder instance for method chaining
func (b *MessageBuilder) WithTimestamp(value time.Time) *MessageBuilder {
	b.timestamp = value
	return b
}

// WithOrigin sets the origin for the message being built.
//
// Parameters:
//   - value: the origin to be set for the message
//
// Returns:
//   - *MessageBuilder: builder instance for method chaining
func (b *MessageBuilder) WithOrigin(value string) *MessageBuilder {
	b.origin = value
	return b
}

// WithVersion sets the version for the message being built.
//
// Parameters:
//   - value: the version to be set for the message
//
// Returns:
//   - *MessageBuilder: builder instance for method chaining
func (b *MessageBuilder) WithVersion(value string) *MessageBuilder {
	b.version = value
	return b
}

// WithRawMessage sets raw message for the message being built.
//
// Parameters:
//   - value: the raw message
//
// Returns:
//   - *MessageBuilder: builder instance for method chaining
func (b *MessageBuilder) WithRawMessage(value any) *MessageBuilder {
	b.rawMessage = value
	return b
}

// Build constructs a new message instance with all configured properties.
//
// Returns:
//   - *Message: the constructed message instance
func (b *MessageBuilder) Build() *Message {
	headers := b.buildHeaders()
	msg := NewMessage(b.payload, headers, b.context)
	if b.rawMessage != nil {
		msg.SetRawMessage(b.rawMessage)
	}
	return msg
}

// buildHeaders creates the message headers from the builder's configuration.
//
// Returns:
//   - *messageHeaders: the constructed message headers
func (b *MessageBuilder) buildHeaders() *messageHeaders {
	headers := NewMessageHeaders(
		b.origin,
		b.messageId,
		b.route,
		b.messageType,
		b.replyChannel,
		b.correlationId,
		b.channelName,
		b.replyChannelName,
		b.timestamp,
		b.version,
	)
	if b.customHeaders != nil {
		headers.SetCustomHeaders(b.customHeaders)
	}
	return headers
}
