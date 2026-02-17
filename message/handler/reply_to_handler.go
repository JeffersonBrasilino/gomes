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
	"github.com/jeffersonbrasilino/gomes/otel"
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
	otelTrace      otel.OtelTrace
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
		otelTrace:      otel.InitTrace("send-reply-to-handler"),
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

	ctx, span := s.otelTrace.Start(
		ctx,
		"Send message to reply channel",
		otel.WithMessagingSystemType(otel.MessageSystemTypeInternal),
		otel.WithSpanOperation(otel.SpanOperationProcess),
		otel.WithSpanKind(otel.SpanKindInternal),
		otel.WithMessage(msg),
	)
	defer span.End()

	replyToChannelName := msg.GetHeader().Get(message.HeaderReplyTo)

	if replyToChannelName == "" {
		err := fmt.Errorf(
			"[send-reply-to-handler] cannot send message: channel not specified",
		)
		span.Error(err, "[send-reply-to-handler] cannot send message: channel not specified")
		return nil, err
	}

	replyChannel, errch := s.gomesContainer.Get(replyToChannelName)
	if errch != nil {
		span.Error(errch, "[send-reply-to-handler] failed to retrieve reply channel from container")
		return nil, fmt.Errorf("[send-reply-to-handler] %v", errch.Error())
	}

	channel, ok := replyChannel.(message.PublisherChannel)
	if !ok {
		err := fmt.Errorf(
			"[send-reply-to-handler] reply channel is not a publisher channel",
		)
		span.Error(err, "[send-reply-to-handler] reply channel is not a publisher channel")
		return nil, err
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
		span.Success("[send-reply-to-handler] sent error message to reply channel")

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
	span.Success("[send-reply-to-handler] sent reply message to reply channel")

	if errorMessage, ok := replyMessage.GetPayload().(error); ok {
		return nil, errorMessage
	}

	return replyMessage, nil
}
