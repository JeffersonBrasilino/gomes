package handler

import (
	"context"
	"log/slog"
	"time"

	"github.com/jeffersonbrasilino/gomes/message"
)

type retryHandler struct {
	handler      message.MessageHandler
	attemptsTime []int
}

func NewRetryHandler(attemptsTime []int, handler message.MessageHandler) *retryHandler {
	return &retryHandler{handler: handler, attemptsTime: attemptsTime}
}

func (h *retryHandler) Handle(
	ctx context.Context,
	msg *message.Message,
) (*message.Message, error) {
	resultMessage, err := h.handler.Handle(ctx, msg)
	if err != nil {
		for k, attempt := range h.attemptsTime {
			slog.Info("[retry-handler] retrying process message after error",
				"messageId", msg.GetHeaders().MessageId,
				"attempt", k+1,
			)
			time.Sleep(time.Millisecond * time.Duration(attempt))
			resultMessage, err = h.handler.Handle(ctx, msg)
			if err == nil {
				return resultMessage, nil
			}
		}
	}
	return resultMessage, err
}
