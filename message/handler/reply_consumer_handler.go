// Package handler provides message handling components for the message system.
//
// This package implements various message handlers that process and route messages
// through the system. It provides specialized handlers for different message
// processing scenarios including reply handling, context management, and error
// handling patterns.
//
// The ReplyConsumerHandler implementation supports:
// - Reply message processing and routing
// - Consumer channel integration
// - Error handling and validation
// - Context-aware message processing
package handler

import (
	"context"
	"fmt"

	"github.com/jeffersonbrasilino/gomes/message"
)

// replyConsumerHandler processes reply messages by receiving them from consumer
// channels and handling the response appropriately.
type replyConsumerHandler struct{}

// NewReplyConsumerHandler creates a new reply consumer handler instance.
//
// Returns:
//   - *replyConsumerHandler: configured reply consumer handler
func NewReplyConsumerHandler() *replyConsumerHandler {
	return &replyConsumerHandler{}
}

// Handle processes reply messages by receiving them from the configured reply
// channel and handling the response or error appropriately.
//
// Parameters:
//   - ctx: context for timeout/cancellation control
//   - msg: the message containing reply channel information
//
// Returns:
//   - *message.Message: the reply message if successful
//   - error: error if processing fails or reply channel is invalid
func (s *replyConsumerHandler) Handle(
	ctx context.Context,
	msg *message.Message,
) (*message.Message, error) {

	channel := msg.GetHeaders().ReplyChannel
	if channel == nil {
		return nil, fmt.Errorf("reply channel not found")
	}

	replyChannel, ok := channel.(message.ConsumerChannel)
	if !ok {
		return nil, fmt.Errorf("reply channel is not a consumer channel")
	}

	replyMessage, err := replyChannel.Receive(ctx)

	if err != nil {
		return nil, err
	}

	if errorMessage, ok := replyMessage.GetPayload().(error); ok {
		return nil, errorMessage
	}

	return replyMessage, nil
}
