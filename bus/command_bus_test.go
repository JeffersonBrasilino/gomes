package bus_test

import (
	"context"
	"errors"
	"testing"

	"github.com/hex-api-go/pkg/core/infrastructure/gomes/bus"
	"github.com/hex-api-go/pkg/core/infrastructure/gomes/message"
)

type mockDispatcher struct {
	lastMsg    *message.Message
	returnErr  error
	returnAny  any
	publishErr error
}

func (m *mockDispatcher) SendMessage(ctx context.Context, msg *message.Message) (any, error) {
	m.lastMsg = msg
	return m.returnAny, m.returnErr
}

func (m *mockDispatcher) PublishMessage(ctx context.Context, msg *message.Message) error {
	m.lastMsg = msg
	return m.publishErr
}

// Mock action for handler.Action
type mockAction struct {
	name string
}

func (a mockAction) Name() string { return a.name }

func TestCommandBus_Send(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		t.Parallel()
		dispatcher := &mockDispatcher{returnAny: "ok", returnErr: nil}
		cb := bus.NewCommandBus(dispatcher)
		ctx := context.Background()
		action := mockAction{name: "TestAction"}

		result, err := cb.Send(ctx, action)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if result != "ok" {
			t.Errorf("expected result 'ok', got %v", result)
		}
		if dispatcher.lastMsg == nil {
			t.Errorf("expected message to be sent")
		}
	})

	t.Run("error", func(t *testing.T) {
		t.Parallel()
		dispatcher := &mockDispatcher{returnErr: errors.New("fail")}
		cb := bus.NewCommandBus(dispatcher)
		ctx := context.Background()
		action := mockAction{name: "TestAction"}

		_, err := cb.Send(ctx, action)
		if err == nil {
			t.Errorf("expected error, got nil")
		}
	})
}

func TestCommandBus_SendRaw(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		t.Parallel()
		dispatcher := &mockDispatcher{returnAny: "raw", returnErr: nil}
		cb := bus.NewCommandBus(dispatcher)
		ctx := context.Background()
		payload := []byte("data")
		headers := map[string]string{"x": "y"}

		result, err := cb.SendRaw(ctx, "route", payload, headers)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if result != "raw" {
			t.Errorf("expected result 'raw', got %v", result)
		}
		if dispatcher.lastMsg == nil {
			t.Errorf("expected message to be sent")
		}
	})
	t.Run("error", func(t *testing.T) {
		t.Parallel()
		dispatcher := &mockDispatcher{returnErr: errors.New("fail")}
		cb := bus.NewCommandBus(dispatcher)
		ctx := context.Background()
		payload := []byte("data")
		headers := map[string]string{"x": "y"}

		_, err := cb.SendRaw(ctx, "route", payload, headers)
		if err == nil {
			t.Errorf("expected error, got nil")
		}
	})
}

func TestCommandBus_SendAsync(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		t.Parallel()
		dispatcher := &mockDispatcher{publishErr: nil}
		cb := bus.NewCommandBus(dispatcher)
		ctx := context.Background()
		action := mockAction{name: "AsyncAction"}

		err := cb.SendAsync(ctx, action)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if dispatcher.lastMsg == nil {
			t.Errorf("expected message to be published")
		}
	})

	t.Run("error", func(t *testing.T) {
		t.Parallel()
		dispatcher := &mockDispatcher{publishErr: errors.New("fail")}
		cb := bus.NewCommandBus(dispatcher)
		ctx := context.Background()
		action := mockAction{name: "AsyncAction"}

		err := cb.SendAsync(ctx, action)
		if err == nil {
			t.Errorf("expected error, got nil")
		}
	})

}

func TestCommandBus_SendRawAsync(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		t.Parallel()
		dispatcher := &mockDispatcher{publishErr: nil}
		cb := bus.NewCommandBus(dispatcher)
		ctx := context.Background()
		payload := "data"
		headers := map[string]string{"x": "y"}

		err := cb.SendRawAsync(ctx, "route", payload, headers)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if dispatcher.lastMsg == nil {
			t.Errorf("expected message to be published")
		}
	})

	t.Run("error", func(t *testing.T) {
		t.Parallel()
		dispatcher := &mockDispatcher{publishErr: errors.New("fail")}
		cb := bus.NewCommandBus(dispatcher)
		ctx := context.Background()
		payload := "data"
		headers := map[string]string{"x": "y"}

		err := cb.SendRawAsync(ctx, "route", payload, headers)
		if err == nil {
			t.Errorf("expected error, got nil")
		}
	})
}
