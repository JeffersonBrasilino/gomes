package handler_test

import (
	"context"
	"errors"
	"testing"

	"github.com/hex-api-go/pkg/core/infrastructure/gomes/message"
	"github.com/hex-api-go/pkg/core/infrastructure/gomes/message/handler"
)

type mockRetryMessageHandler struct {
	shouldFail        bool
	failErr           error
	attempts          int
	attemptSuccessNro int
}

func (m *mockRetryMessageHandler) Handle(ctx context.Context, msg *message.Message) (*message.Message, error) {
	if m.attemptSuccessNro == m.attempts && m.attempts != 0 {
		return msg, nil
	}
	m.attempts++
	if m.shouldFail {
		return nil, m.failErr
	}
	return msg, nil
}

func TestRetryHandler_Handle(t *testing.T) {

	msg := message.NewMessageBuilder().
		WithPayload("payload").
		Build()
	ctx := context.Background()

	t.Run("should process message successfully", func(t *testing.T) {
		t.Parallel()
		handlerMock := &mockRetryMessageHandler{shouldFail: false}
		dl := handler.NewRetryHandler([]int{500}, handlerMock)
		retMsg, err := dl.Handle(ctx, msg)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if retMsg != msg {
			t.Errorf("expected returned message to be input message")
		}
	})

	t.Run("should return error when retry handler failed", func(t *testing.T) {
		t.Parallel()
		dlErr := errors.New("handler failed")
		handlerMock := &mockRetryMessageHandler{shouldFail: true, failErr: dlErr}
		dl := handler.NewRetryHandler([]int{500}, handlerMock)
		_, err := dl.Handle(ctx, msg)
		if err == nil {
			t.Errorf("expected error, got nil")
		}
	})

	t.Run("should return success when retry handler success", func(t *testing.T) {
		t.Parallel()
		dlErr := errors.New("handler failed")
		handlerMock := &mockRetryMessageHandler{shouldFail: true, failErr: dlErr, attemptSuccessNro: 1}
		dl := handler.NewRetryHandler([]int{250, 250}, handlerMock)
		msgR, err := dl.Handle(ctx, msg)
		if err != nil {
			t.Errorf("expected error nil, got %v", err.Error())
		}
		if msgR != msg {
			t.Errorf("response is not equal to %v", msg)
		}
	})
}
