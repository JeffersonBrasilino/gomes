package endpoint

import (
	"context"

	"github.com/jeffersonbrasilino/gomes/message"
)

// InboundChannelAdapter defines the contract for inbound channel adapters that
// receive messages from external sources.
type InboundChannelAdapter interface {
	ReferenceName() string
	DeadLetterChannelName() string
	AfterProcessors() []message.MessageHandler
	BeforeProcessors() []message.MessageHandler
	ReceiveMessage(ctx context.Context) (*message.Message, error)
	RetryAttempts() []int
	Close() error
}

type Dispatcher interface {
	SendMessage(
		ctx context.Context,
		msg *message.Message,
	) (any, error)

	PublishMessage(
		ctx context.Context,
		msg *message.Message,
	) error
}

type OutboundChannelAdapter interface {
	Send(ctx context.Context, message *message.Message) error
	Close() error
}
