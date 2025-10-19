// Package bus provides message bus implementations for the message system.
//
// This package implements various message bus types that provide high-level
// abstractions for sending and receiving messages. It supports command/query
// separation (CQRS) patterns and event-driven messaging with different bus
// types for different use cases.
//
// The QueryBus implementation supports:
// - Synchronous query execution with response handling
// - Raw query execution with custom payload and headers
// - Asynchronous query execution for fire-and-forget scenarios
// - Automatic correlation ID generation
package bus

import (
	"context"

	"github.com/google/uuid"
	"github.com/jeffersonbrasilino/gomes/message"
	"github.com/jeffersonbrasilino/gomes/message/endpoint"
	"github.com/jeffersonbrasilino/gomes/message/handler"
)

// QueryBus provides query execution capabilities for data retrieval operations.
type QueryBus struct {
	dispatcher endpoint.Dispatcher
}

// NewQueryBus creates a new query bus instance with the specified dispatcher.
//
// Parameters:
//   - dispatcher: the message dispatcher to use for query execution
//
// Returns:
//   - *QueryBus: new query bus instance
func NewQueryBus(dispatcher endpoint.Dispatcher) *QueryBus {

	queryBus := &QueryBus{
		dispatcher: dispatcher,
	}
	return queryBus
}

// Send executes a query action synchronously and returns the result.
//
// Parameters:
//   - ctx: context for timeout/cancellation control
//   - action: the query action to be executed
//
// Returns:
//   - any: the query result
//   - error: error if query execution fails
func (c *QueryBus) Send(
	ctx context.Context,
	action handler.Action,
) (any, error) {
	builder := c.buildMessage()
	msg := builder.WithPayload(action).
		WithRoute(action.Name()).
		Build()

	return c.dispatcher.SendMessage(ctx, msg)
}

// SendRaw executes a raw query with custom payload and headers synchronously.
//
// Parameters:
//   - ctx: context for timeout/cancellation control
//   - route: the route for the query
//   - payload: the query payload
//   - headers: custom headers for the query
//
// Returns:
//   - any: the query result
//   - error: error if query execution fails
func (c *QueryBus) SendRaw(
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

// SendAsync executes a query action asynchronously without waiting for a response.
//
// Parameters:
//   - ctx: context for timeout/cancellation control
//   - action: the query action to be executed
//
// Returns:
//   - error: error if query execution fails
func (c *QueryBus) SendAsync(
	ctx context.Context,
	action handler.Action,
) error {
	builder := c.buildMessage()
	msg := builder.WithPayload(action).
		WithRoute(action.Name()).
		Build()
	return c.dispatcher.PublishMessage(ctx, msg)
}

// SendRawAsync executes a raw query asynchronously with custom payload and headers.
//
// Parameters:
//   - ctx: context for timeout/cancellation control
//   - route: the route for the query
//   - payload: the query payload
//   - headers: custom headers for the query
//
// Returns:
//   - error: error if query execution fails
func (c *QueryBus) SendRawAsync(
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

// buildMessage creates a message builder configured for query messages with
// automatic correlation ID generation.
//
// Returns:
//   - *message.MessageBuilder: configured message builder for queries
func (c *QueryBus) buildMessage() *message.MessageBuilder {
	builder := message.NewMessageBuilder().
		WithMessageType(message.Query).
		WithCorrelationId(uuid.New().String())
	return builder
}
