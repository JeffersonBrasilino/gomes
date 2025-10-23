package adapter

import "github.com/jeffersonbrasilino/gomes/message"

// ChannelConnection defines the contract for managing channel connections
// with connect and disconnect capabilities.
type ChannelConnection interface {
	ReferenceName() string
	Connect() error
	Disconnect() error
}

type ClosableChannel interface{
	Close() error
}

// InboundChannelMessageTranslator defines the contract for translating external messages
// to the internal format.
//
// T represents the external message type that needs to be translated.
type InboundChannelMessageTranslator[T any] interface {
	// ToMessage converts an external message to the internal message format.
	//
	// Parameters:
	//   - msg: The external message to be translated
	//
	// Returns:
	//   - *message.Message: The translated message in internal format
	ToMessage(msg T) (*message.Message, error)
}

// OutboundChannelMessageTranslator defines the contract for translating internal messages
// to external system format.
//
// T represents the target message type for the external system.
type OutboundChannelMessageTranslator[T any] interface {
	// FromMessage converts an internal message to the target external format.
	//
	// Parameters:
	//   - msg: The internal message to be translated
	//
	// Returns:
	//   - T: The translated message in external format
	FromMessage(msg *message.Message) (T, error)
}