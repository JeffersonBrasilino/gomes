// Package bus provides message bus implementations for the message system.
//
// This package implements various message bus types that provide high-level
// abstractions for sending and receiving messages. It supports command/query
// separation (CQRS) patterns and event-driven messaging with different bus
// types for different use cases.
//
// The EventBus implementation supports:
// - Event publishing for notifications and broadcasts
// - Raw message publishing with custom headers
// - Automatic correlation ID generation
// - Asynchronous event distribution
package bus

import (
	"context"

	"github.com/google/uuid"
	"github.com/jeffersonbrasilino/gomes/message"
	"github.com/jeffersonbrasilino/gomes/message/endpoint"
	"github.com/jeffersonbrasilino/gomes/message/handler"
)

// EventBus provides event publishing capabilities for broadcasting events
// throughout the system.
type EventBus struct {
	dispatcher endpoint.Dispatcher
}

// NewEventBus creates a new event bus instance with the specified dispatcher.
//
// Parameters:
//   - dispatcher: the message dispatcher to use for publishing events
//
// Returns:
//   - *EventBus: new event bus instance
func NewEventBus(dispatcher endpoint.Dispatcher) *EventBus {

	eventBus := &EventBus{
		dispatcher: dispatcher,
	}
	return eventBus
}

// Publish publishes an event action asynchronously to the message system.
//
// Parameters:
//   - ctx: context for timeout/cancellation control
//   - action: the action to be published as an event
//
// Returns:
//   - error: error if publishing fails
func (c *EventBus) Publish(ctx context.Context, action handler.Action) error {
	builder := c.buildMessage()
	msg := builder.WithPayload(action).
		WithRoute(action.Name()).
		Build()
	return c.dispatcher.PublishMessage(ctx, msg)
}

// PublishRaw publishes a raw event message with custom payload and headers.
//
// Parameters:
//   - ctx: context for timeout/cancellation control
//   - route: the route for the event
//   - payload: the event payload
//   - headers: custom headers for the event
//
// Returns:
//   - error: error if publishing fails
func (c *EventBus) PublishRaw(
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

// buildMessage creates a message builder configured for event messages with
// automatic correlation ID generation.
//
// Returns:
//   - *message.MessageBuilder: configured message builder for events
func (c *EventBus) buildMessage() *message.MessageBuilder {
	builder := message.NewMessageBuilder().
		WithMessageType(message.Event).
		WithCorrelationId(uuid.New().String())
	return builder
}
