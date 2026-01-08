// Package rabbitmq provides RabbitMQ message translation functionality.
package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"maps"

	"github.com/jeffersonbrasilino/gomes/message"
	"github.com/jeffersonbrasilino/gomes/otel"
	amqp "github.com/rabbitmq/amqp091-go"
)

// MessageTranslator provides message translation capabilities between internal
// message formats and RabbitMQ-specific AMQP formats.
type MessageTranslator struct{}

// NewMessageTranslator creates a new message translator instance.
//
// Returns:
//   - *MessageTranslator: new message translator instance
func NewMessageTranslator() *MessageTranslator {
	return &MessageTranslator{}
}

// FromMessage translates an internal message to RabbitMQ AMQP Publishing
// format. It serializes the message payload to JSON, converts headers to
// AMQP Table format, and injects OpenTelemetry trace context for distributed
// tracing.
//
// Parameters:
//   - msg: the internal message to translate
//
// Returns:
//   - *amqp.Publishing: AMQP publishing message with headers and JSON body
//   - error: error if header conversion or JSON marshaling fails
func (m *MessageTranslator) FromMessage(
	msg *message.Message,
) (*amqp.Publishing, error) {
	headersMap, err := msg.GetHeaders().ToMap()
	if err != nil {
		return nil, fmt.Errorf(
			"[rabbitMQ-message-translator] header converter error: %v",
			err.Error(),
		)
	}

	contextPropagator := otel.GetTraceContextPropagatorByContext(
		msg.GetContext(),
	)
	if contextPropagator != nil {
		maps.Copy(headersMap, contextPropagator)
	}

	pld, err := json.Marshal(msg.GetPayload())
	if err != nil {
		return nil, fmt.Errorf(
			"[rabbitMQ-message-translator] converter error: %v",
			err.Error(),
		)
	}

	headers := amqp.Table{}
	for k, v := range headersMap {
		headers[k] = v
	}

	return &amqp.Publishing{
		ContentType: "application/json",
		Headers:     headers,
		Body:        pld,
	}, nil
}

// ToMessage translates a RabbitMQ AMQP Delivery message to internal message
// format. It extracts headers, reconstructs OpenTelemetry trace context if
// present, and builds the internal message with the raw AMQP delivery.
//
// Parameters:
//   - msg: the AMQP delivery message to translate
//
// Returns:
//   - *message.Message: internal message with extracted headers and payload
//   - error: error if header conversion or message building fails
func (m *MessageTranslator) ToMessage(
	msg amqp.Delivery,
) (*message.Message, error) {
	headers := map[string]string{}
	for k, h := range msg.Headers {
		if strVal, ok := h.(string); ok {
			headers[k] = strVal
		}
	}

	messageBuilder, err := message.NewMessageBuilderFromHeaders(headers)
	if err != nil {
		return nil, fmt.Errorf(
			"[rabbitMQ-message-translator] header converter error: %v",
			err.Error(),
		)
	}

	traceParenValue, exists := headers["Traceparent"]
	if exists && traceParenValue != "" {
		ctx := otel.GetTraceContextPropagatorByTraceParent(
			context.Background(),
			traceParenValue,
		)
		messageBuilder.WithContext(ctx)
	}

	messageBuilder.WithPayload(msg.Body)
	messageBuilder.WithRawMessage(msg)
	buildedMessage := messageBuilder.Build()

	return buildedMessage, nil
}
