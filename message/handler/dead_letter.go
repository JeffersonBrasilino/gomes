// Package handler provides message handling components for the message system.
//
// This package implements various message handlers that process and route messages
// through the system. It provides specialized handlers for different message
// processing scenarios including dead letter handling, context management, and
// error handling patterns.
//
// The DeadLetter implementation supports:
// - Failed message handling and routing
// - Dead letter channel integration
// - Error logging and monitoring
// - Graceful error recovery patterns
package handler

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/jeffersonbrasilino/gomes/message"
)

// deadLetter implements the Dead Letter Channel pattern, routing failed messages
// to a designated dead letter channel for further processing or analysis.
type deadLetter struct {
	channel message.PublisherChannel
	handler message.MessageHandler
}
type deadLetterMessage struct {
	ReasonError string
	Payload     any
	Headers     map[string]string
}

// NewDeadLetter creates a new dead letter handler instance that routes failed
// messages to the specified dead letter channel.
//
// Parameters:
//   - channel: the publisher channel for sending failed messages
//   - handler: the message handler to attempt processing with
//
// Returns:
//   - *deadLetter: configured dead letter handler
func NewDeadLetter(
	channel message.PublisherChannel,
	handler message.MessageHandler,
) *deadLetter {
	return &deadLetter{
		channel: channel,
		handler: handler,
	}
}

// Handle processes a message by attempting to process it with the wrapped handler.
// If processing fails, the message is sent to the dead letter channel for further
// analysis or processing.
//
// Parameters:
//   - ctx: context for timeout/cancellation control
//   - msg: the message to be processed
//
// Returns:
//   - *message.Message: the original message (regardless of processing success)
//   - error: error if processing fails (message is sent to dead letter channel)
func (s *deadLetter) Handle(
	ctx context.Context,
	msg *message.Message,
) (*message.Message, error) {
	resultMessage, err := s.handler.Handle(ctx, msg)
	if err == nil {
		return resultMessage, nil
	}

	originalPayload, errP := s.convertMessagePayload(msg)
	if errP != nil {
		slog.Info("[dead-letter-handler] cannot convert original payload",
			"messageId", msg.GetHeaders().MessageId,
			"reason", errP.Error(),
			"dlqChannelName", s.channel.Name(),
		)

		return resultMessage, err
	}

	ctxDql := context.Background()
	dlqMessage := s.makeDeadLetterMessage(ctxDql, msg, &deadLetterMessage{
		ReasonError: err.Error(),
		Payload:     originalPayload,
	})

	errDql := s.channel.Send(ctxDql, dlqMessage)
	if errDql != nil {
		slog.Error("[dead-letter-handler] failed to send message to dead letter",
			"messageId", msg.GetHeaders().MessageId,
			"reason", errDql.Error(),
			"dlqChannelName", s.channel.Name(),
		)
		return resultMessage, err
	}

	slog.Info("[dead-letter-handler] Sent message to dead letter",
		"messageId", msg.GetHeaders().MessageId,
		"reason", err.Error(),
		"dlqChannelName", s.channel.Name(),
	)

	return resultMessage, err
}

func (s *deadLetter) convertMessagePayload(msg *message.Message) (any, error) {
	originalPayload, ok := msg.GetPayload().([]byte)
	if ok {
		var payloadMap any
		errU := json.Unmarshal(originalPayload, &payloadMap)
		return payloadMap, errU
	}

	return msg.GetPayload(), nil
}

func (s *deadLetter) makeDeadLetterMessage(
	ctxDql context.Context,
	msg *message.Message,
	payload *deadLetterMessage,
) *message.Message {
	headers, _ := msg.GetHeaders().ToMap()
	payload.Headers = headers
	dlqMessage := message.NewMessageBuilder()
	dlqMessage.WithContext(ctxDql)
	dlqMessage.WithChannelName(s.channel.Name())
	dlqMessage.WithMessageType(message.Document)
	dlqMessage.WithCorrelationId(msg.GetHeaders().CorrelationId)
	dlqMessage.WithPayload(payload)

	return dlqMessage.Build()
}
