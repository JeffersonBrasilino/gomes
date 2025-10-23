package rabbitmq

import (
	"encoding/json"
	"fmt"

	"github.com/jeffersonbrasilino/gomes/message"
	amqp "github.com/rabbitmq/amqp091-go"
)

// MessageTranslator provides message translation capabilities between internal
// message formats and RabbitMQ-specific formats.
type MessageTranslator struct{}

// NewMessageTranslator creates a new message translator instance.
//
// Returns:
//   - *MessageTranslator: new message translator instance
func NewMessageTranslator() *MessageTranslator {
	return &MessageTranslator{}
}

func (m *MessageTranslator) FromMessage(msg *message.Message) (*amqp.Publishing, error) {
	headersMap, err := msg.GetHeaders().ToMap()
	if err != nil {
		return nil, fmt.Errorf(
			"[rabbitMQ-message-translator] header converter error: %v", err.Error(),
		)
	}

	pld, err :=json.Marshal(msg.GetPayload())
	if err != nil{
		return nil, fmt.Errorf(
			"[rabbitMQ-message-translator] converter error: %v", err.Error(),
		)
	}

	headers :=amqp.Table{}
	for k, v := range headersMap {
		headers[k] = v
	}

	return &amqp.Publishing{
		ContentType: "application/json",
		Headers: headers,
		Body: pld,
	}, nil
}

func (m *MessageTranslator) ToMessage(msg any) (*message.Message, error) {
	return nil, nil
}
