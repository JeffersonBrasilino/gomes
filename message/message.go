// Package message provides core message types and interfaces for the message system.
//
// This package defines the fundamental message structures, types, and interfaces
// that form the foundation of the messaging system. It implements message types,
// headers, channels, and handlers that enable flexible message processing and
// routing throughout the system.
//
// The Message implementation supports:
// - Multiple message types (Command, Query, Event, Document)
// - Comprehensive header management
// - Context-aware message processing
// - JSON serialization and deserialization
// - Channel interfaces for different messaging patterns
package message

import (
	"context"
	"fmt"
	"maps"
	"slices"
	"time"

	"github.com/google/uuid"
)

// MessageType constants define the different types of messages supported by the
// system.
const (
	Command             MessageType = iota // Command messages for actions
	Query                                  // Query messages for data retrieval
	Event                                  // Event messages for notifications
	Document                               // Document messages for data transfer
	HeaderOrigin        = "origin"
	HeaderRoute         = "route"
	HeaderMessageType   = "messageType"
	HeaderTimestamp     = "timestamp"
	HeaderCorrelationId = "correlationId"
	HeaderChannelName   = "channelName"
	HeaderMessageId     = "messageId"
	HeaderReplyTo       = "replyTo"
	HeaderVersion       = "version"
)

var restrictedHeaders = []string{
	HeaderMessageId,
	HeaderMessageType,
	HeaderTimestamp,
	HeaderOrigin,
}

// MessageType represents the type of a message in the system.
type MessageType int8

// Header represents a map of header key-value pairs for message metadata.
type Header map[string]string

// MessageHandler defines the contract for processing messages in the system.
type MessageHandler interface {
	Handle(ctx context.Context, message *Message) (*Message, error)
}

// PublisherChannel defines the contract for channels that can publish messages.
type PublisherChannel interface {
	Name() string
	Send(ctx context.Context, message *Message) error
}

// ConsumerChannel defines the contract for channels that can consume messages.
type ConsumerChannel interface {
	Name() string
	Receive(ctx context.Context) (*Message, error)
	Close() error
}

// SubscriberChannel defines the contract for channels that support subscription
// patterns.
type SubscriberChannel interface {
	Name() string
	Subscribe(callable ...func(m *Message))
	Unsubscribe() error
}

// Message represents a message in the system with payload, headers, and context.
type Message struct {
	payload              any
	header               Header
	context              context.Context
	rawMessage           any
	internalreplyChannel PublisherChannel
}

// NewHeader creates a new header with default values and custom attributes.
// It automatically sets MessageId, Timestamp, Origin, and Version if not provided.
//
// Parameters:
//   - attributes: Map of custom header key-value pairs, or nil for empty map
//
// Returns:
//   - Header: New header instance with defaults and custom attributes
func NewHeader(attributes map[string]string) Header {
	if attributes == nil {
		attributes = make(map[string]string)
	}

	attributes[HeaderMessageId] = uuid.New().String()

	if val, ok := attributes[HeaderTimestamp]; !ok || val == "" {
		attributes[HeaderTimestamp] = time.Now().Format("2006-01-02 15:04:05")
	}

	if val, ok := attributes[HeaderOrigin]; !ok || val == "" {
		attributes[HeaderOrigin] = "Gomes"
	}

	if val, ok := attributes[HeaderVersion]; !ok || val == "" {
		attributes[HeaderVersion] = "1.0"
	}

	return Header(attributes)
}

// Set sets a custom header value. Restricted headers cannot be set manually.
//
// Parameters:
//   - key: The header key to set
//   - value: The header value
//
// Returns:
//   - error: Error if the header key is restricted
func (h Header) Set(key string, value string) error {

	if slices.Contains(restrictedHeaders, key) {
		return fmt.Errorf(
			"header %s is restricted and cannot be set manually",
			key,
		)
	}

	h[key] = value
	return nil
}

// Get retrieves a header value by key.
//
// Parameters:
//   - key: The header key to retrieve
//
// Returns:
//   - string: The header value, or empty string if not found
func (h Header) Get(key string) string {
	val, ok := h[key]
	if !ok {
		return ""
	}
	return val
}

// All returns a copy of all headers as a map.
//
// Returns:
//   - map[string]string: A shallow copy of all headers
func (h Header) All() map[string]string {
	headerCopy := maps.Clone(h)
	return headerCopy
}

// String returns the string representation of a MessageType.
//
// Returns:
//   - string: the string representation of the message type
func (m MessageType) String() string {
	switch m {
	case Command:
		return "Command"
	case Query:
		return "Query"
	case Event:
		return "Event"
	}
	return "Document"
}

// NewMessage creates a new message instance with the specified payload, headers,
// and context.
//
// Parameters:
//   - payload: the data carried by the message
//   - headers: the message headers containing metadata
//   - context: the context for the message
//
// Returns:
//   - *Message: new message instance
func NewMessage(
	context context.Context,
	payload any,
	header Header,
) *Message {
	return &Message{
		payload: payload,
		header:  header,
		context: context,
	}
}

// GetPayload returns the payload of the message.
//
// Returns:
//   - any: the message payload
func (m *Message) GetPayload() any {
	return m.payload
}

// GetHeaders returns the headers of the message.
//
// Returns:
//   - Header: The message headers
func (m *Message) GetHeader() Header {
	return m.header
}

// SetContext sets the context for the message.
//
// Parameters:
//   - ctx: the context to be set
func (m *Message) SetContext(ctx context.Context) {
	m.context = ctx
}

// GetContext returns the context of the message.
//
// Returns:
//   - context.Context: the message context
func (m *Message) GetContext() context.Context {
	return m.context
}

// ReplyRequired determines if the message requires a reply based on its type.
// Commands and Queries typically require replies, while Events and Documents
// do not.
//
// Returns:
//   - bool: true if the message requires a reply, false otherwise
func (m *Message) ReplyRequired() bool {
	return m.header[HeaderMessageType] == Command.String() ||
		m.header[HeaderMessageType] == Query.String()
}

// SetRawMessage sets the raw message from the external source.
//
// Parameters:
//   - rawMessage: The raw message to store
func (m *Message) SetRawMessage(rawMessage any) {
	m.rawMessage = rawMessage
}

// GetRawMessage returns the raw message from the external source.
//
// Returns:
//   - any: The raw message
func (m *Message) GetRawMessage() any {
	return m.rawMessage
}

// SetInternalReplyChannel sets the internal reply channel for the message.
//
// Parameters:
//   - channel: The publisher channel for replies
func (m *Message) SetInternalReplyChannel(channel PublisherChannel) {
	m.internalreplyChannel = channel
}

// GetInternalReplyChannel returns the internal reply channel for the message.
//
// Returns:
//   - PublisherChannel: The publisher channel for replies
func (m *Message) GetInternalReplyChannel() PublisherChannel {
	return m.internalreplyChannel
}
