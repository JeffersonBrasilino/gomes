// Package handler provides message handlers for processing and intercepting messages
// in the system's message pipeline. It includes various handler implementations for
// acknowledgment, retry logic, dead letter handling, and other message processing
// concerns following the Enterprise Integration Patterns.
package handler

import (
	"context"
	"log/slog"

	"github.com/jeffersonbrasilino/gomes/message"
)

// ChannelMessageAcknowledgment defines the interface for acknowledging successful
// message processing on a communication channel.
type ChannelMessageAcknowledgment interface {
	// CommitMessage marks a message as successfully processed and commits it to the
	// underlying channel, preventing message redelivery.
	//
	// Parameters:
	//   - msg: The message to acknowledge
	//
	// Returns:
	//   - error: Error if the commitment fails
	CommitMessage(msg *message.Message) error
}

// acknowledgeHandler wraps a message handler with automatic message acknowledgment
// support, ensuring messages are committed after successful processing.
type acknowledgeHandler struct {
	channelAdapter ChannelMessageAcknowledgment
	handler        message.MessageHandler
}

// NewAcknowledgeHandler creates a new acknowledge handler that wraps an existing
// message handler with automatic message commitment after processing.
//
// Parameters:
//   - channel: The channel message acknowledgment implementation
//   - handler: The underlying message handler to wrap
//
// Returns:
//   - *acknowledgeHandler: Configured acknowledge handler instance
func NewAcknowledgeHandler(
	channel ChannelMessageAcknowledgment,
	handler message.MessageHandler,
) *acknowledgeHandler {
	return &acknowledgeHandler{channelAdapter: channel, handler: handler}
}

// Handle processes a message through the wrapped handler and automatically
// acknowledges it after processing, regardless of success or failure.
//
// Parameters:
//   - ctx: Context for timeout/cancellation control
//   - msg: The message to process
//
// Returns:
//   - *message.Message: The resulting message from processing
//   - error: Error from the handler if processing fails
func (h *acknowledgeHandler) Handle(
	ctx context.Context,
	msg *message.Message,
) (*message.Message, error) {
	resultMessage, err := h.handler.Handle(ctx, msg)
	errC := h.channelAdapter.CommitMessage(msg)
	if errC != nil {
		slog.Info("[acknowledgeHandler-handler] failed to acknowledge message:",
			"messageId", msg.GetHeader().Get(message.HeaderMessageId),
			"reason", errC.Error(),
		)
	}
	return resultMessage, err
}
