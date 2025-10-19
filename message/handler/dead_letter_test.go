package handler_test

import (
	"context"
	"errors"
	"testing"

	"github.com/jeffersonbrasilino/gomes/message"
	"github.com/jeffersonbrasilino/gomes/message/handler"
)

type mockPublisherChannel struct {
	sentMsg *message.Message
	ctx     context.Context
}

func (m *mockPublisherChannel) Send(ctx context.Context, msg *message.Message) error {
	m.sentMsg = msg
	m.ctx = ctx
	return nil
}
func (m *mockPublisherChannel) Name() string {
	return "mock"
}

type mockDeadMessageHandler struct {
	shouldFail bool
	failErr    error
}

func (m *mockDeadMessageHandler) Handle(ctx context.Context, msg *message.Message) (*message.Message, error) {
	if m.shouldFail {
		return nil, m.failErr
	}
	return msg, nil
}

func TestDeadLetter_Handle(t *testing.T) {

	msg := message.NewMessageBuilder().
		WithPayload("payload").
		Build()
	ctx := context.Background()

	t.Run("should process message successfully", func(t *testing.T) {
		t.Parallel()
		channel := &mockPublisherChannel{}
		handlerMock := &mockDeadMessageHandler{shouldFail: false}
		dl := handler.NewDeadLetter(channel, handlerMock)
		retMsg, err := dl.Handle(ctx, msg)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if retMsg != msg {
			t.Errorf("expected returned message to be input message")
		}
		if channel.sentMsg != nil {
			t.Errorf("expected no message sent to dead letter channel")
		}
	})

	t.Run("should send to dead letter channel on handler error", func(t *testing.T) {
		t.Parallel()
		dlErr := errors.New("handler failed")
		channel := &mockPublisherChannel{}
		handlerMock := &mockDeadMessageHandler{shouldFail: true, failErr: dlErr}
		dl := handler.NewDeadLetter(channel, handlerMock)
		_, err := dl.Handle(ctx, msg)
		if err == nil {
			t.Errorf("expected error, got nil")
		}
	})

	t.Run("should error when convert message payload", func(t *testing.T) {
		t.Parallel()
		dlErr := errors.New("handler failed")
		msgE := message.NewMessageBuilder().
			WithPayload([]byte(`{"name": "Alice`)).
			Build()
		channel := &mockPublisherChannel{}
		handlerMock := &mockDeadMessageHandler{shouldFail: true, failErr: dlErr}
		dl := handler.NewDeadLetter(channel, handlerMock)
		retMsg, err := dl.Handle(ctx, msgE)
		if err == nil {
			t.Errorf("expected error, got nil")
		}
		if retMsg != nil {
			t.Errorf("expected returned message to be nil")
		}
		if err.Error() != dlErr.Error() {
			t.Errorf("expected error to be %v, got %v", dlErr.Error(), err.Error())
		}
	})
}
