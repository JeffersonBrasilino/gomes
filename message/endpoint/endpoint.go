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
	SendReplyUsingReplyTo() bool
}

type OutboundChannelAdapter interface {
	Send(ctx context.Context, message *message.Message) error
	Close() error
}
