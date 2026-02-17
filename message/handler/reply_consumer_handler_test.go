package handler_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/jeffersonbrasilino/gomes/container"
	"github.com/jeffersonbrasilino/gomes/message"
	"github.com/jeffersonbrasilino/gomes/message/handler"
)

// mockConsumerChannel implements message.ConsumerChannel for tests.
type mockConsumerChannel struct {
	msgReceived chan *message.Message
	shouldError bool
}

func (d *mockConsumerChannel) Send(_ context.Context, msg *message.Message) error {
	if d.shouldError {
		return errors.New("erro de canal")
	}
	d.msgReceived <- msg
	return nil
}

func (d *mockConsumerChannel) Receive(ctx context.Context) (*message.Message, error) {
	msg := <-d.msgReceived
	if msg.GetPayload() == "error" {
		return nil, fmt.Errorf("error processing")
	}
	return msg, nil
}

func (d *mockConsumerChannel) Name() string {
	return "canal1"
}

func (d *mockConsumerChannel) Close() error {
	close(d.msgReceived)
	return nil
}

type invalidConsumerChannel struct{}

func (d *invalidConsumerChannel) Send(_ context.Context, msg *message.Message) error {
	return nil
}
func (d *invalidConsumerChannel) Name() string {
	return "invalid channel"
}
func TestReplyConsumerHandler_Handle(t *testing.T) {
	t.Run("should process reply message successfully", func(t *testing.T) {
		t.Parallel()
		chn := make(chan *message.Message, 50)
		ch := &mockConsumerChannel{msgReceived: chn}
		requestMessage := message.NewMessageBuilder().
			WithPayload("payload").
			WithMessageType(message.Command).
			WithInternalReplyChannel(ch).
			Build()
		responseMessage := message.NewMessageBuilder().WithPayload("ok").Build()
		ch.Send(context.Background(), responseMessage)
		container := container.NewGenericContainer[any, any]()
		h := handler.NewReplyConsumerHandler(container)
		res, err := h.Handle(context.Background(), requestMessage)

		if err != nil {
			t.Errorf("Expected success, got error: %v", err)
		}
		if res != responseMessage {
			t.Error("Expected returned message to match input")
		}
	})
	t.Run("should return error if reply channel is nil", func(t *testing.T) {
		t.Parallel()
		requestMessage := message.NewMessageBuilder().
			WithPayload("payload").
			WithMessageType(message.Command).
			WithInternalReplyChannel(nil).
			Build()
		container := container.NewGenericContainer[any, any]()
		h := handler.NewReplyConsumerHandler(container)
		res, err := h.Handle(context.Background(), requestMessage)
		if err.Error() != "reply channel not found" {
			t.Errorf("Expected error for nil channel, got: %v", err)
		}
		if res != nil {
			t.Error("Expected nil message for nil channel")
		}
	})
	t.Run("should return error when handler processing", func(t *testing.T) {
		t.Parallel()
		chn := make(chan *message.Message, 50)
		ch := &mockConsumerChannel{msgReceived: chn}
		requestMessage := message.NewMessageBuilder().
			WithPayload("payload").
			WithMessageType(message.Command).
			WithInternalReplyChannel(ch).
			Build()

		responseMessage := message.NewMessageBuilder().WithPayload("error").Build()
		ch.Send(context.Background(), responseMessage)

		container := container.NewGenericContainer[any, any]()
		h := handler.NewReplyConsumerHandler(container)
		res, err := h.Handle(context.Background(), requestMessage)

		if err.Error() != "error processing" {
			t.Errorf("Expected error for nil channel, got: %v", err)
		}
		if res != nil {
			t.Error("Expected nil message for nil channel")
		}

	})
	t.Run("should return error when error message type", func(t *testing.T) {
		t.Parallel()
		chn := make(chan *message.Message, 50)
		ch := &mockConsumerChannel{msgReceived: chn}
		requestMessage := message.NewMessageBuilder().
			WithPayload("payload").
			WithMessageType(message.Command).
			WithInternalReplyChannel(ch).
			Build()

		responseMessage := message.NewMessageBuilder().WithPayload(fmt.Errorf("error")).Build()
		ch.Send(context.Background(), responseMessage)

		container := container.NewGenericContainer[any, any]()
		h := handler.NewReplyConsumerHandler(container)
		res, err := h.Handle(context.Background(), requestMessage)

		if err.Error() != "error" {
			t.Errorf("Expected error for nil channel, got: %v", err)
		}
		if res != nil {
			t.Error("Expected nil message for nil channel")
		}
	})

	t.Run("should return error when invalid consumer channel", func(t *testing.T) {
		t.Parallel()

		requestMessage := message.NewMessageBuilder().
			WithPayload("payload").
			WithMessageType(message.Command).
			WithInternalReplyChannel(&invalidConsumerChannel{}).
			Build()

		container := container.NewGenericContainer[any, any]()
		h := handler.NewReplyConsumerHandler(container)
		res, err := h.Handle(context.Background(), requestMessage)

		if err.Error() != "reply channel is not a consumer channel" {
			t.Errorf("Expected error for nil channel, got: %v", err)
		}
		if res != nil {
			t.Error("Expected nil message for nil channel")
		}
	})
}
