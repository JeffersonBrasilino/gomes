// Package handler provides message handlers for processing and intercepting messages
// in the system's message pipeline. It includes various handler implementations for
// acknowledgment, retry logic, dead letter handling, and other message processing
// concerns following the Enterprise Integration Patterns.
package handler

import (
	"context"
	"fmt"

	"github.com/jeffersonbrasilino/gomes/container"
	"github.com/jeffersonbrasilino/gomes/message"
)

// ErrorResult represents an error response to be sent as a message payload.
type ErrorResult struct {
	// Result contains the error message string.
	Result string `json:"error"`
}

// SendReplyToHandler handles sending reply messages to the channel specified in
// the original message's reply-to header, supporting asynchronous request-response
// patterns.
type SendReplyToHandler struct {
	gomesContainer container.Container[any, any]
	handler        message.MessageHandler
}

// NewSendReplyToHandler creates a new send reply-to handler that wraps an existing
// message handler and sends responses to the configured reply channel.
//
// Parameters:
//   - handler: The underlying message handler to wrap
//   - container: Dependency container for retrieving reply channel instances
//
// Returns:
//   - *SendReplyToHandler: Configured send reply-to handler instance
func NewSendReplyToHandler(
	handler message.MessageHandler,
	container container.Container[any, any],
) *SendReplyToHandler {
	return &SendReplyToHandler{
		gomesContainer: container,
		handler:        handler,
	}
}

// Handle processes a message through the wrapped handler and sends the result to
// the reply channel specified in the message's reply-to header. Errors during
// processing are serialized and sent as ErrorResult payloads.
//
// Parameters:
//   - ctx: Context for timeout/cancellation control
//   - msg: The message to process
//
// Returns:
//   - *message.Message: The resulting message from processing
//   - error: Error if the reply-to channel is not specified or retrieval fails
func (s *SendReplyToHandler) Handle(
	ctx context.Context,
	msg *message.Message,
) (*message.Message, error) {

	replyMessage, err := s.handler.Handle(ctx, msg)
	replyToChannelName := msg.GetHeader().Get(message.HeaderReplyTo)

	if replyToChannelName == "" {
		return nil, fmt.Errorf(
			"[send-reply-to-handler] cannot send message: channel not specified",
		)
	}

	replyChannel, errch := s.gomesContainer.Get(replyToChannelName)
	if errch != nil {
		return nil, fmt.Errorf("[send-reply-to-handler] %v", errch.Error())
	}

	channel, ok := replyChannel.(message.PublisherChannel)
	if !ok {
		return nil, fmt.Errorf(
			"[send-reply-to-handler] reply channel is not a publisher channel",
		)
	}

	if err != nil {
		rplMessage := message.NewMessageBuilder().
			WithChannelName(msg.GetInternalReplyChannel().Name()).
			WithMessageType(message.Document).
			WithCorrelationId(
				msg.GetHeader().Get(message.HeaderCorrelationId),
			).
			WithChannelName(replyToChannelName).
			WithPayload(&ErrorResult{err.Error()}).
			Build()

		channel.Send(ctx, rplMessage)

		return nil, err
	}

	rplMessage := replyMessage
	if payload, ok := msg.GetPayload().(error); ok {
		rplMessage = message.NewMessageBuilderFromMessage(
			replyMessage,
		).
			WithPayload(&ErrorResult{payload.Error()}).
			Build()
	}

	channel.Send(ctx, rplMessage)

	if errorMessage, ok := replyMessage.GetPayload().(error); ok {
		return nil, errorMessage
	}

	return replyMessage, nil
}
