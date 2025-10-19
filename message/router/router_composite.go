// Package router provides message routing components for the message system.
//
// This package implements various routing patterns from Enterprise Integration
// Patterns, enabling flexible message routing and processing through different
// channels and handlers. It provides composite routing, recipient list routing,
// and message filtering capabilities.
//
// The Router implementation supports:
// - Composite routing with multiple handlers
// - Sequential message processing
// - Error handling and propagation
// - Flexible handler composition
package router

import (
	"context"

	"github.com/jeffersonbrasilino/gomes/message"
)

// router implements a composite router that processes messages through multiple
// handlers in sequence, allowing for complex routing and processing pipelines.
type router struct {
	handlers []message.MessageHandler
}

// NewRouter creates a new composite router instance.
//
// Returns:
//   - *router: configured composite router
func NewRouter() *router {
	return &router{
		handlers: []message.MessageHandler{},
	}
}

// AddHandler adds a message handler to the router's processing pipeline.
//
// Parameters:
//   - handler: the message handler to be added to the pipeline
//
// Returns:
//   - *router: router instance for method chaining
func (r *router) AddHandler(handler message.MessageHandler) *router {
	r.handlers = append(r.handlers, handler)
	return r
}

// Handle processes a message through all registered handlers in sequence.
// Processing stops if any handler returns an error or nil message.
//
// Parameters:
//   - ctx: context for timeout/cancellation control
//   - msg: the message to be processed through the pipeline
//
// Returns:
//   - *message.Message: the processed message if successful
//   - error: error if any handler fails or returns an error
func (r *router) Handle(
	ctx context.Context,
	msg *message.Message,
) (*message.Message, error) {
	var resultMessage = msg
	var resultError error
	for _, handler := range r.handlers {
		if resultMessage == nil {
			break
		}
		resultMessage, resultError = handler.Handle(ctx, resultMessage)
		if resultError != nil {
			break
		}
	}

	return resultMessage, resultError
}
