// Package handler provides message handling components for the message system.
//
// This package implements various message handlers that process and route messages
// through the system. It provides specialized handlers for different message
// processing scenarios including context management, reply handling, and error
// handling patterns.
//
// The ContextHandler implementation supports:
// - Context-aware message processing
// - Timeout and cancellation handling
// - Context error propagation
// - Graceful error handling for different context states
package handler

import (
	"context"

	"github.com/hex-api-go/pkg/core/infrastructure/gomes/message"
)

// contextHandler wraps another message handler and provides context-aware
// processing with timeout and cancellation support.
type contextHandler struct {
	handler message.MessageHandler
}

// NewContextHandler creates a new context handler instance that wraps the
// provided message handler.
//
// Parameters:
//   - handler: the message handler to be wrapped with context support
//
// Returns:
//   - *contextHandler: configured context handler
func NewContextHandler(handler message.MessageHandler) *contextHandler {
	return &contextHandler{handler: handler}
}

// Handle processes a message with context awareness, checking for timeout
// and cancellation before delegating to the wrapped handler.
//
// Parameters:
//   - ctx: context for timeout/cancellation control
//   - msg: the message to be processed
//
// Returns:
//   - *message.Message: the processed message if successful
//   - error: error if context is cancelled or processing fails
func (h *contextHandler) Handle(
	ctx context.Context,
	msg *message.Message,
) (*message.Message, error) {
	select {
	case <-ctx.Done():
		switch ctx.Err() {
		case context.DeadlineExceeded:
			return nil, context.DeadlineExceeded
		case context.Canceled:
			return nil, context.Canceled
		default:
			return nil, ctx.Err()
		}
	default:
	}
	return h.handler.Handle(ctx, msg)
}
