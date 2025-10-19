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
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// MessageType represents the type of a message in the system.
type MessageType int8

// MessageType constants define the different types of messages supported by the
// system.
const (
	Command  MessageType = iota // Command messages for actions
	Query                       // Query messages for data retrieval
	Event                       // Event messages for notifications
	Document                    // Document messages for data transfer
)

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

// customHeaders represents a map of custom header key-value pairs.
type CustomHeaders map[string]string

// messageHeaders contains all the metadata associated with a message including
// routing information, timestamps, and custom headers.
type messageHeaders struct {
	Origin           string
	Route            string
	MessageType      MessageType
	Timestamp        time.Time
	ReplyChannel     PublisherChannel
	CustomHeaders    CustomHeaders
	CorrelationId    string
	ChannelName      string
	MessageId        string
	ReplyChannelName string
	Version          string
}

// Message represents a message in the system with payload, headers, and context.
type Message struct {
	payload    any
	headers    *messageHeaders
	context    context.Context
	rawMessage any
}

// NewMessageHeaders creates a new message headers instance with the specified
// parameters and automatically generated message ID and timestamp.
//
// Parameters:
//   - route: the routing information for the message
//   - messageType: the type of the message
//   - replyChannel: the channel for reply messages
//   - correlationId: the correlation identifier for message tracking
//   - channelName: the name of the target channel
//   - replyChannelName: the name of the reply channel
//
// Returns:
//   - *messageHeaders: new message headers instance
func NewMessageHeaders(
	origin string,
	messageId string,
	route string,
	messageType MessageType,
	replyChannel PublisherChannel,
	correlationId string,
	channelName string,
	replyChannelName string,
	timestamp time.Time,
	version string,
) *messageHeaders {
	if messageId == "" {
		messageId = uuid.New().String()
	}
	if timestamp.IsZero() {
		timestamp = time.Now()
	}
	if origin == "" {
		origin = "gomes"
	}
	if version == "" {
		version = "1.0"
	}
	return &messageHeaders{
		Origin:           origin,
		MessageId:        messageId,
		Route:            route,
		MessageType:      messageType,
		Timestamp:        timestamp,
		ReplyChannel:     replyChannel,
		CustomHeaders:    make(CustomHeaders),
		CorrelationId:    correlationId,
		ChannelName:      channelName,
		ReplyChannelName: replyChannelName,
		Version:          version,
	}
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
	payload any,
	headers *messageHeaders,
	context context.Context,
) *Message {
	return &Message{
		payload: payload,
		headers: headers,
		context: context,
	}
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

// SetCustomHeaders sets the custom headers for the message.
//
// Parameters:
//   - data: the custom headers to be set
func (m *messageHeaders) SetCustomHeaders(data CustomHeaders) {
	m.CustomHeaders = data
}

// ToMap converts the message headers to a map[string]string representation.
//
// Returns:
//   - map[string]string: a map containing all header fields as strings
//   - error: error if marshaling custom headers fails
func (m *messageHeaders) ToMap() (map[string]string, error) {
	chs, err := json.Marshal(m.CustomHeaders)
	if err != nil {
		return nil, err
	}

	var customHeaders string
	if len(chs) > 2 {
		customHeaders = string(chs)
	}

	return map[string]string{
		"origin":        m.Origin,
		"route":         m.Route,
		"type":          m.MessageType.String(),
		"timestamp":     m.Timestamp.Format("2006-01-02 15:04:05"),
		"replyChannel":  m.ReplyChannelName,
		"customHeaders": customHeaders,
		"correlationId": m.CorrelationId,
		"channelName":   m.ChannelName,
		"messageId":     m.MessageId,
		"version":       m.Version,
	}, nil
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
//   - *messageHeaders: the message headers
func (m *Message) GetHeaders() *messageHeaders {
	return m.headers
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
	return m.headers.MessageType == Command || m.headers.MessageType == Query
}

// SetRawMessage sets the raw message.
//
// Parameters:
//   - rawMessage: the raw message
func (m *Message) SetRawMessage(rawMessage any) {
	m.rawMessage = rawMessage
}

// GetRawMessage returns the raw message.
//
// Returns:
//   - any: the raw message
func (m *Message) GetRawMessage() any {
	return m.rawMessage
}
