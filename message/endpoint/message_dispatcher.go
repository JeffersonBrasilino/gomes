// Package endpoint provides message dispatching capabilities for the message system.
//
// This package implements the Message Dispatcher pattern, enabling applications to
// send messages to specific channels and receive responses. It provides a simplified
// interface for message routing and processing through gateways.
//
// The MessageDispatcher implementation supports:
// - Synchronous message sending with response handling
// - Asynchronous message publishing
// - Integration with gateway-based message processing
// - Context-aware operations with timeout support
package endpoint

import (
	"context"
	"fmt"

	"github.com/hex-api-go/pkg/core/infrastructure/gomes/container"
	"github.com/hex-api-go/pkg/core/infrastructure/gomes/message"
)

type Dispatcher interface {
	SendMessage(
		ctx context.Context,
		msg *message.Message,
	) (any, error)

	PublishMessage(
		ctx context.Context,
		msg *message.Message,
	) error
}

// messageDispatcherBuilder provides a builder pattern for creating MessageDispatcher
// instances with specific channel configurations.
type messageDispatcherBuilder struct {
	referenceName      string
	requestChannelName string
}

// MessageDispatcher handles message dispatching operations through configured gateways.
// It provides both synchronous and asynchronous message sending capabilities.
type MessageDispatcher struct {
	gateway *Gateway
}

// NewMessageDispatcherBuilder creates a new message dispatcher builder instance.
//
// Parameters:
//   - referenceName: unique identifier for the dispatcher
//   - requestChannelName: name of the channel to send messages to
//
// Returns:
//   - *messageDispatcherBuilder: configured builder instance
func NewMessageDispatcherBuilder(
	referenceName string,
	requestChannelName string,
) *messageDispatcherBuilder {
	return &messageDispatcherBuilder{
		referenceName:      referenceName,
		requestChannelName: requestChannelName,
	}
}

// NewMessageDispatcher creates a new message dispatcher instance.
//
// Parameters:
//   - gateway: the gateway to use for message processing
//
// Returns:
//   - *MessageDispatcher: configured message dispatcher
func NewMessageDispatcher(gateway *Gateway) *MessageDispatcher {
	return &MessageDispatcher{
		gateway: gateway,
	}
}

// Build constructs a MessageDispatcher from the dependency container.
//
// Parameters:
//   - container: dependency container containing required components
//
// Returns:
//   - *MessageDispatcher: configured message dispatcher
//   - error: error if construction fails
func (b *messageDispatcherBuilder) Build(
	container container.Container[any, any],
) (*MessageDispatcher, error) {

	gateway, err := NewGatewayBuilder(
		b.referenceName,
		b.requestChannelName,
	).
		Build(container)

	if err != nil {
		return nil, fmt.Errorf("[message-dispatcher] %s", err)
	}

	dispatcher := NewMessageDispatcher(gateway)
	return dispatcher, nil
}

// SendMessage sends a message synchronously and waits for a response.
//
// Parameters:
//   - ctx: context for timeout/cancellation control
//   - msg: the message to be sent
//
// Returns:
//   - any: the response from message processing
//   - error: error if sending or processing fails
func (m *MessageDispatcher) SendMessage(
	ctx context.Context,
	msg *message.Message,
) (any, error) {

	result, err := m.gateway.Execute(ctx, msg)

	if err != nil {
		return nil, err
	}

	return result, nil
}

// PublishMessage publishes a message asynchronously without waiting for a response.
//
// Parameters:
//   - ctx: context for timeout/cancellation control
//   - msg: the message to be published
//
// Returns:
//   - error: error if publishing fails
func (m *MessageDispatcher) PublishMessage(
	ctx context.Context,
	msg *message.Message,
) error {
	_, err := m.gateway.Execute(ctx, msg)
	if err != nil {
		return err
	}

	return nil
}
