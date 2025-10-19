// Package bus provides message bus implementations for the message system.
//
// This package implements various message bus types that provide high-level
// abstractions for sending and receiving messages. It supports command/query
// separation (CQRS) patterns and event-driven messaging with different bus
// types for different use cases.
//
// The CommandBus implementation supports:
// - Synchronous command execution with response handling
// - Raw command execution with custom payload and headers
// - Asynchronous command execution for fire-and-forget scenarios
// - Automatic correlation ID generation
package bus

import (
	"context"

	"github.com/google/uuid"
	"github.com/jeffersonbrasilino/gomes/message"
	"github.com/jeffersonbrasilino/gomes/message/endpoint"
	"github.com/jeffersonbrasilino/gomes/message/handler"
)

// CommandBus provides command execution capabilities for action processing.
type CommandBus struct {
	dispatcher endpoint.Dispatcher
}

// NewCommandBus creates a new command bus instance with the specified dispatcher.
//
// Parameters:
//   - dispatcher: the message dispatcher to use for command execution
//
// Returns:
//   - *CommandBus: new command bus instance
func NewCommandBus(dispatcher endpoint.Dispatcher) *CommandBus {
	commandBus := &CommandBus{
		dispatcher: dispatcher,
	}
	return commandBus
}

// Send executes a command action synchronously and returns the result.
//
// Parameters:
//   - ctx: context for timeout/cancellation control
//   - action: the command action to be executed
//
// Returns:
//   - any: the command result
//   - error: error if command execution fails
func (c *CommandBus) Send(
	ctx context.Context,
	action handler.Action,
) (any, error) {

	builder := c.buildMessage()
	msg := builder.WithPayload(action).
		WithRoute(action.Name()).
		Build()
	return c.dispatcher.SendMessage(ctx, msg)
}

// SendRaw executes a raw command with custom payload and headers synchronously.
//
// Parameters:
//   - ctx: context for timeout/cancellation control
//   - route: the route for the command
//   - payload: the command payload
//   - headers: custom headers for the command
//
// Returns:
//   - any: the command result
//   - error: error if command execution fails
func (c *CommandBus) SendRaw(
	ctx context.Context,
	route string,
	payload []byte,
	headers map[string]string,
) (any, error) {
	builder := c.buildMessage()
	msg := builder.WithPayload(payload).
		WithRoute(route).
		WithCustomHeader(headers).
		Build()
	return c.dispatcher.SendMessage(ctx, msg)
}

// SendAsync executes a command action asynchronously without waiting for a response.
//
// Parameters:
//   - ctx: context for timeout/cancellation control
//   - action: the command action to be executed
//
// Returns:
//   - error: error if command execution fails
func (c *CommandBus) SendAsync(
	ctx context.Context,
	action handler.Action,
) error {
	builder := c.buildMessage()
	msg := builder.WithPayload(action).
		WithRoute(action.Name()).
		Build()
	return c.dispatcher.PublishMessage(ctx, msg)
}

// SendRawAsync executes a raw command asynchronously with custom payload and headers.
//
// Parameters:
//   - ctx: context for timeout/cancellation control
//   - route: the route for the command
//   - payload: the command payload
//   - headers: custom headers for the command
//
// Returns:
//   - error: error if command execution fails
func (c *CommandBus) SendRawAsync(
	ctx context.Context,
	route string,
	payload any,
	headers map[string]string,
) error {
	builder := c.buildMessage()
	msg := builder.WithPayload(payload).
		WithRoute(route).
		WithCustomHeader(headers).
		Build()
	return c.dispatcher.PublishMessage(ctx, msg)
}

// buildMessage creates a message builder configured for command messages with
// automatic correlation ID generation.
//
// Returns:
//   - *message.MessageBuilder: configured message builder for commands
func (c *CommandBus) buildMessage() *message.MessageBuilder {
	builder := message.NewMessageBuilder().
		WithMessageType(message.Command).
		WithCorrelationId(uuid.New().String())
	return builder
}
