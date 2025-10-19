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
	"time"

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
	messageBuilder := message.NewMessageBuilder()
	headersMap := map[string]func(value string) error{
		"origin": func(value string) error {
			messageBuilder.WithOrigin(value)
			return nil
		},
		"route": func(value string) error {
			messageBuilder.WithRoute(value)
			return nil
		},
		"type": func(value string) error {
			tp := m.chooseMessageType(value)
			messageBuilder.WithMessageType(tp)
			return nil
		},
		"timestamp": func(value string) error {
			dt, err := time.Parse("2006-01-02 15:04:05", value)
			if err != nil {
				return err
			}
			messageBuilder.WithTimestamp(dt)
			return nil
		},
		"replyChannel": func(value string) error {
			messageBuilder.WithReplyChannelName(value)
			return nil
		},
		"customHeaders": func(value string) error {
			ch, err := m.makeCustomHeaders(value)
			if err != nil {
				return err
			}
			messageBuilder.WithCustomHeader(ch)
			return nil
		},
		"correlationId": func(value string) error {
			messageBuilder.WithCorrelationId(value)
			return nil
		},
		"channelName": func(value string) error {
			messageBuilder.WithChannelName(value)
			return nil
		},
		"messageId": func(value string) error {
			messageBuilder.WithMessageId(value)
			return nil
		},
		"version": func(value string) error {
			messageBuilder.WithVersion(value)
			return nil
		},
	}

	for _, h := range data.Headers {
		key := string(h.Key)
		if headersMap[key] == nil {
			continue
		}

		if string(h.Value) == "" {
			continue
		}

		err := headersMap[key](string(h.Value))
		if err != nil {
			return nil, fmt.Errorf("[kafka-message-translator] header converter error: %v - %v", key, err.Error())
		}
	}

	messageBuilder.WithPayload(data.Value)
	messageBuilder.WithRawMessage(data)
	msg := messageBuilder.Build()
	return msg, nil
}

func (m *MessageTranslator) chooseMessageType(value string) message.MessageType {
	switch value {
	case "Command":
		return message.Command
	case "Query":
		return message.Query
	case "Event":
		return message.Event
	}
	return message.Document
}

func (m *MessageTranslator) makeCustomHeaders(value string) (message.CustomHeaders, error) {
	var customHeaders message.CustomHeaders
	errCh := json.Unmarshal([]byte(value), &customHeaders)
	if errCh != nil {
		return nil, errCh
	}
	return customHeaders, nil
}
