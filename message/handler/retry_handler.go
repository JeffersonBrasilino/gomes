// Package handler provides message handlers for processing and intercepting messages
// in the system's message pipeline. It includes various handler implementations for
// acknowledgment, retry logic, dead letter handling, and other message processing
// concerns following the Enterprise Integration Patterns.
package handler

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/jeffersonbrasilino/gomes/message"
)

// retryHandler implements retry logic for failed message processing attempts,
// allowing configurable delays between retry attempts.
type retryHandler struct {
	handler      message.MessageHandler
	attemptsTime []int
}

// NewRetryHandler creates a new retry handler that wraps an existing message
// handler with automatic retry capability on processing failures.
//
// Parameters:
//   - attemptsTime: Array of retry delay intervals in milliseconds
//   - handler: The underlying message handler to wrap
//
// Returns:
//   - *retryHandler: Configured retry handler instance
func NewRetryHandler(
	attemptsTime []int,
	handler message.MessageHandler,
) *retryHandler {
	return &retryHandler{handler: handler, attemptsTime: attemptsTime}
}

// Handle processes a message through the wrapped handler with automatic retry on
// failure. If processing fails, it retries with configured delay intervals until
// success or all retries are exhausted.
//
// Parameters:
//   - ctx: Context for timeout/cancellation control
//   - msg: The message to process
//
// Returns:
//   - *message.Message: The resulting message from processing
//   - error: Error if all retry attempts fail
func (h *retryHandler) Handle(
	ctx context.Context,
	msg *message.Message,
) (*message.Message, error) {
	resultMessage, err := h.handler.Handle(ctx, msg)
	if err == nil {
		return resultMessage, nil
	}

	for k, attempt := range h.attemptsTime {
		select {
		case <-ctx.Done():
			return msg, ctx.Err()
		default:
		}

		slog.Info(
			"[retry-handler] retrying process message after error",
			"message.id",
			msg.GetHeader().Get(message.HeaderMessageId),
			"attempt", k+1,
			"start.in",
			fmt.Sprintf("%v milliseconds", attempt),
		)
		time.Sleep(time.Millisecond * time.Duration(attempt))
		resultMessage, err = h.handler.Handle(ctx, msg)
		if err == nil {
			return resultMessage, nil
		}
	}
	return resultMessage, err
}
