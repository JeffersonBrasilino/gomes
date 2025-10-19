package handler_test

import (
	"context"
	"errors"
	"testing"

	"github.com/jeffersonbrasilino/gomes/message"
	"github.com/jeffersonbrasilino/gomes/message/handler"
)

// Mock implementations for testing
type mockChannelMessageAcknowledgment struct {
	commitError error
	committed   bool
}

func (m *mockChannelMessageAcknowledgment) CommitMessage(msg *message.Message) error {
	m.committed = true
	return m.commitError
}

type mockAcknowledgeMessageHandler struct {
	handleError error
	result      *message.Message
}

func (m *mockAcknowledgeMessageHandler) Handle(ctx context.Context, msg *message.Message) (*message.Message, error) {
	return m.result, m.handleError
}

func TestNewAcknowledgeHandler(t *testing.T) {
	t.Run("should create acknowledge handler successfully", func(t *testing.T) {
		t.Parallel()

		mockChannel := &mockChannelMessageAcknowledgment{}
		mockHandler := &mockAcknowledgeMessageHandler{}

		ackHandler := handler.NewAcknowledgeHandler(mockChannel, mockHandler)

		if ackHandler == nil {
			t.Error("NewAcknowledgeHandler should not return nil")
		}
	})

	t.Run("should create acknowledge handler with nil channel", func(t *testing.T) {
		t.Parallel()

		mockHandler := &mockAcknowledgeMessageHandler{}

		ackHandler := handler.NewAcknowledgeHandler(nil, mockHandler)

		if ackHandler == nil {
			t.Error("NewAcknowledgeHandler should not return nil even with nil channel")
		}
	})

	t.Run("should create acknowledge handler with nil handler", func(t *testing.T) {
		t.Parallel()

		mockChannel := &mockChannelMessageAcknowledgment{}

		ackHandler := handler.NewAcknowledgeHandler(mockChannel, nil)

		if ackHandler == nil {
			t.Error("NewAcknowledgeHandler should not return nil even with nil handler")
		}
	})
}

func TestAcknowledgeHandler_Handle(t *testing.T) {
	t.Run("should handle message successfully with successful commit", func(t *testing.T) {
		t.Parallel()

		mockChannel := &mockChannelMessageAcknowledgment{}
		mockHandler := &mockAcknowledgeMessageHandler{
			result: &message.Message{},
		}

		ackHandler := handler.NewAcknowledgeHandler(mockChannel, mockHandler)

		ctx := context.Background()
		msg := message.NewMessageBuilder().Build()

		result, err := ackHandler.Handle(ctx, msg)

		if err != nil {
			t.Errorf("Handle should not return error, got: %v", err)
		}

		if result == nil {
			t.Error("Handle should return result message")
		}

		if !mockChannel.committed {
			t.Error("Channel should have been committed")
		}
	})

	t.Run("should handle message successfully with commit error", func(t *testing.T) {
		t.Parallel()

		mockChannel := &mockChannelMessageAcknowledgment{
			commitError: errors.New("commit failed"),
		}
		mockHandler := &mockAcknowledgeMessageHandler{
			result: &message.Message{},
		}

		ackHandler := handler.NewAcknowledgeHandler(mockChannel, mockHandler)

		ctx := context.Background()
		msg := message.NewMessageBuilder().Build()

		result, err := ackHandler.Handle(ctx, msg)

		if err != nil {
			t.Errorf("Handle should not return error from commit failure, got: %v", err)
		}

		if result == nil {
			t.Error("Handle should return result message")
		}

		if !mockChannel.committed {
			t.Error("Channel should have been attempted to commit")
		}
	})

	t.Run("should handle message with handler error and successful commit", func(t *testing.T) {
		t.Parallel()

		mockChannel := &mockChannelMessageAcknowledgment{}
		handlerError := errors.New("handler failed")
		mockHandler := &mockAcknowledgeMessageHandler{
			handleError: handlerError,
		}

		ackHandler := handler.NewAcknowledgeHandler(mockChannel, mockHandler)

		ctx := context.Background()
		msg := message.NewMessageBuilder().Build()

		result, err := ackHandler.Handle(ctx, msg)

		if err == nil {
			t.Error("Handle should return handler error")
		}

		if err != handlerError {
			t.Errorf("Handle should return handler error, got: %v", err)
		}

		if result != nil {
			t.Error("Handle should return nil result when handler fails")
		}

		if !mockChannel.committed {
			t.Error("Channel should have been committed even when handler fails")
		}
	})

	t.Run("should handle message with handler error and commit error", func(t *testing.T) {
		t.Parallel()

		mockChannel := &mockChannelMessageAcknowledgment{
			commitError: errors.New("commit failed"),
		}
		handlerError := errors.New("handler failed")
		mockHandler := &mockAcknowledgeMessageHandler{
			handleError: handlerError,
		}

		ackHandler := handler.NewAcknowledgeHandler(mockChannel, mockHandler)

		ctx := context.Background()
		msg := message.NewMessageBuilder().Build()

		result, err := ackHandler.Handle(ctx, msg)

		if err == nil {
			t.Error("Handle should return handler error")
		}

		if err != handlerError {
			t.Errorf("Handle should return handler error, got: %v", err)
		}

		if result != nil {
			t.Error("Handle should return nil result when handler fails")
		}

		if !mockChannel.committed {
			t.Error("Channel should have been attempted to commit")
		}
	})

	t.Run("should handle message with context cancellation", func(t *testing.T) {
		t.Parallel()

		mockChannel := &mockChannelMessageAcknowledgment{}
		mockHandler := &mockAcknowledgeMessageHandler{
			handleError: context.Canceled,
		}

		ackHandler := handler.NewAcknowledgeHandler(mockChannel, mockHandler)

		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		msg := &message.Message{}

		result, err := ackHandler.Handle(ctx, msg)

		if err == nil {
			t.Error("Handle should return context canceled error")
		}

		if err != context.Canceled {
			t.Errorf("Handle should return context.Canceled error, got: %v", err)
		}

		if result != nil {
			t.Error("Handle should return nil result when context is canceled")
		}

		if !mockChannel.committed {
			t.Error("Channel should have been committed even when context is canceled")
		}
	})

	t.Run("should handle message with nil message", func(t *testing.T) {
		t.Parallel()

		mockChannel := &mockChannelMessageAcknowledgment{}
		mockHandler := &mockAcknowledgeMessageHandler{
			result: &message.Message{},
		}

		ackHandler := handler.NewAcknowledgeHandler(mockChannel, mockHandler)

		ctx := context.Background()

		result, err := ackHandler.Handle(ctx, nil)

		if err != nil {
			t.Errorf("Handle should not return error with nil message, got: %v", err)
		}

		if result == nil {
			t.Error("Handle should return result message")
		}

		if !mockChannel.committed {
			t.Error("Channel should have been committed")
		}
	})
}
