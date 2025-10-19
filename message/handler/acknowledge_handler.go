package handler

import (
	"context"
	"log/slog"

	"github.com/jeffersonbrasilino/gomes/message"
)

type ChannelMessageAcknowledgment interface {
	CommitMessage(msg *message.Message) error
}

type acknowledgeHandler struct {
	channelAdapter ChannelMessageAcknowledgment
	handler        message.MessageHandler
}

func NewAcknowledgeHandler(
	channel ChannelMessageAcknowledgment,
	handler message.MessageHandler,
) *acknowledgeHandler {
	return &acknowledgeHandler{channelAdapter: channel, handler: handler}
}

func (h *acknowledgeHandler) Handle(
	ctx context.Context,
	msg *message.Message,
) (*message.Message, error) {
	resultMessage, err := h.handler.Handle(ctx, msg)
	errC := h.channelAdapter.CommitMessage(msg)
	if errC != nil {
		slog.Info("[acknowledgeHandler-handler] failed to acknowledge message:",
			"messageId", msg.GetHeaders().MessageId,
			"reason", errC.Error(),
		)
	}
	return resultMessage, err
}
