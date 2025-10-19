package handler_test

import (
	"context"
	"errors"
	"testing"

	"github.com/hex-api-go/pkg/core/infrastructure/gomes/message"
	"github.com/hex-api-go/pkg/core/infrastructure/gomes/message/handler"
)

// mockMessageHandler implements message.MessageHandler for tests.
type mockMessageHandler struct {
	result *message.Message
	err    error
}

func (m *mockMessageHandler) Handle(ctx context.Context, msg *message.Message) (*message.Message, error) {
	return m.result, m.err
}

func TestContextHandler_Handle(t *testing.T) {
	msg := &message.Message{}

	t.Run("should process message successfully", func(t *testing.T) {
		t.Parallel()
		h := &mockMessageHandler{result: msg}
		ctx := context.Background()
		contextHandler := handler.NewContextHandler(h)
		res, err := contextHandler.Handle(ctx, msg)
		if err != nil {
			t.Errorf("Expected success, got error: %v", err)
		}
		if res != msg {
			t.Error("Expected returned message to match input")
		}
	})

	t.Run("should return error on context cancel", func(t *testing.T) {
		t.Parallel()
		h := &mockMessageHandler{result: msg}
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		contextHandler := handler.NewContextHandler(h)
		res, err := contextHandler.Handle(ctx, msg)
		if !errors.Is(err, context.Canceled) {
			t.Errorf("Expected context.Canceled error, got: %v", err)
		}
		if res != nil {
			t.Error("Expected nil message when context is canceled")
		}
	})

	t.Run("should return error on context deadline exceeded", func(t *testing.T) {
		t.Parallel()
		h := &mockMessageHandler{result: msg}
		ctx, cancel := context.WithTimeout(context.Background(), 0)
		defer cancel()
		contextHandler := handler.NewContextHandler(h)
		<-ctx.Done()
		res, err := contextHandler.Handle(ctx, msg)
		if !errors.Is(err, context.DeadlineExceeded) {
			t.Errorf("Expected context.DeadlineExceeded error, got: %v", err)
		}
		if res != nil {
			t.Error("Expected nil message when deadline exceeded")
		}
	})

	t.Run("should propagate handler error", func(t *testing.T) {
		t.Parallel()
		h := &mockMessageHandler{err: errors.New("handler error")}
		ctx := context.Background()
		contextHandler := handler.NewContextHandler(h)
		res, err := contextHandler.Handle(ctx, msg)
		if err == nil || err.Error() != "handler error" {
			t.Errorf("Expected handler error, got: %v", err)
		}
		if res != nil {
			t.Error("Expected nil message when handler returns error")
		}
	})
}
