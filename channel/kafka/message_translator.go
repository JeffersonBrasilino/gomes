// Package kafka provides Kafka integration for the message system.
//
// This package implements Kafka-specific channel adapters and connections for
// publishing and consuming messages through Apache Kafka. It provides outbound
// and inbound channel adapters with message translation capabilities.
//
// The MessageTranslator implementation supports:
// - Message translation between internal and Kafka formats
// - JSON serialization and deserialization
// - Header mapping and conversion
// - Error handling for translation failures
package kafka

import (
	"encoding/json"
	"fmt"

	"github.com/jeffersonbrasilino/gomes/message"
	"github.com/segmentio/kafka-go"
)

// MessageTranslator provides message translation capabilities between internal
// message formats and Kafka-specific formats.
type MessageTranslator struct{}

// NewMessageTranslator creates a new message translator instance.
//
// Returns:
//   - *MessageTranslator: new message translator instance
func NewMessageTranslator() *MessageTranslator {
	return &MessageTranslator{}
}

// FromMessage converts an internal message to a Kafka producer message format.
// It serializes the message headers and payload to JSON and creates appropriate
// Kafka record headers.
//
// Parameters:
//   - msg: the internal message to be converted
//
// Returns:
//   - *kafka.Message: the Kafka producer message
func (m *MessageTranslator) FromMessage(msg *message.Message) (*kafka.Message, error) {
	headersMap, err := msg.GetHeaders().ToMap()
	if err != nil {
		return nil, fmt.Errorf("[kafka-message-translator] header converter error: %v", err.Error())
	}

	kafkaHeaders := []kafka.Header{}
	for k, v := range headersMap {
		kafkaHeaders = append(kafkaHeaders, kafka.Header{
			Key:   k,
			Value: []byte(v),
		})
	}

	payload, err := json.Marshal(msg.GetPayload())
	if err != nil {
		return nil, fmt.Errorf("[kafka-message-translator] payload converter error: %v", err.Error())
	}

	return &kafka.Message{
		Topic:   msg.GetHeaders().ChannelName,
		Key:     []byte(msg.GetHeaders().CorrelationId),
		Value:   payload,
		Headers: kafkaHeaders,
	}, nil
}

// ToMessage converts a Kafka consumer message to an internal message format using map-dispatcher pattern.
// This method is currently a placeholder and should be implemented based on
// specific requirements for message consumption.
//
// Parameters:
//   - data: the Kafka consumer message to be converted
//
// Returns:
//   - *message.Message: the internal message (placeholder implementation)
func (m *MessageTranslator) ToMessage(data *kafka.Message) (*message.Message, error) {
	headers := map[string]string{}
	for _, h := range data.Headers {
		headers[h.Key] = string(h.Value)
	}
	
	messageBuilder, err := message.NewMessageBuilderFromHeaders(headers)
	if err != nil {
		return nil, fmt.Errorf("[rabbitMQ-message-translator] header converter error: %v", err.Error())
	}

	messageBuilder.WithPayload(data.Value)
	messageBuilder.WithRawMessage(data)
	msg := messageBuilder.Build()
	return msg, nil
}
