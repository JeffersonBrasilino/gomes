package bus

import (
	"context"

	"github.com/jeffersonbrasilino/gomes/message"
)

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
