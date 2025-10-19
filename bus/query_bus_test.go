package bus_test

import (
	"context"
	"errors"
	"testing"

	"github.com/hex-api-go/pkg/core/infrastructure/gomes/bus"
	"github.com/hex-api-go/pkg/core/infrastructure/gomes/message"
)

type mockQDispatcher struct {
	lastMsg    *message.Message
	returnErr  error
	returnAny  any
	publishErr error
}

func (m *mockQDispatcher) SendMessage(ctx context.Context, msg *message.Message) (any, error) {
	m.lastMsg = msg
	return m.returnAny, m.returnErr
}

func (m *mockQDispatcher) PublishMessage(ctx context.Context, msg *message.Message) error {
	m.lastMsg = msg
	return m.publishErr
}

// Mock action for handler.Action
type mockqAction struct {
	name string
}

func (a mockqAction) Name() string { return a.name }
func TestQueryBus_Send(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		t.Parallel()
		dispatcher := &mockQDispatcher{returnAny: "ok", returnErr: nil}
		qb := bus.NewQueryBus(dispatcher)
		ctx := context.Background()
		action := mockqAction{name: "TestQuery"}

		result, err := qb.Send(ctx, action)
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
		dispatcher := &mockQDispatcher{returnErr: errors.New("fail")}
		qb := bus.NewQueryBus(dispatcher)
		ctx := context.Background()
		action := mockqAction{name: "TestQuery"}

		_, err := qb.Send(ctx, action)
		if err == nil {
			t.Errorf("expected error, got nil")
		}
	})
}

func TestQueryBus_SendRaw(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		t.Parallel()
		dispatcher := &mockQDispatcher{returnAny: "raw", returnErr: nil}
		qb := bus.NewQueryBus(dispatcher)
		ctx := context.Background()
		payload := []byte("data")
		headers := map[string]string{"x": "y"}

		result, err := qb.SendRaw(ctx, "route", payload, headers)
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
		dispatcher := &mockQDispatcher{returnErr: errors.New("fail")}
		qb := bus.NewQueryBus(dispatcher)
		ctx := context.Background()
		payload := []byte("data")
		headers := map[string]string{"x": "y"}

		_, err := qb.SendRaw(ctx, "route", payload, headers)
		if err == nil {
			t.Errorf("expected error, got nil")
		}
	})
}

func TestQueryBus_SendAsync(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		t.Parallel()
		dispatcher := &mockQDispatcher{publishErr: nil}
		qb := bus.NewQueryBus(dispatcher)
		ctx := context.Background()
		action := mockqAction{name: "AsyncQuery"}

		err := qb.SendAsync(ctx, action)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if dispatcher.lastMsg == nil {
			t.Errorf("expected message to be published")
		}
	})

	t.Run("error", func(t *testing.T) {
		t.Parallel()
		dispatcher := &mockQDispatcher{publishErr: errors.New("fail")}
		qb := bus.NewQueryBus(dispatcher)
		ctx := context.Background()
		action := mockqAction{name: "AsyncQuery"}

		err := qb.SendAsync(ctx, action)
		if err == nil {
			t.Errorf("expected error, got nil")
		}
	})
}

func TestQueryBus_SendRawAsync(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		t.Parallel()
		dispatcher := &mockQDispatcher{publishErr: nil}
		qb := bus.NewQueryBus(dispatcher)
		ctx := context.Background()
		payload := "data"
		headers := map[string]string{"x": "y"}

		err := qb.SendRawAsync(ctx, "route", payload, headers)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if dispatcher.lastMsg == nil {
			t.Errorf("expected message to be published")
		}
	})

	t.Run("error", func(t *testing.T) {
		t.Parallel()
		dispatcher := &mockQDispatcher{publishErr: errors.New("fail")}
		qb := bus.NewQueryBus(dispatcher)
		ctx := context.Background()
		payload := "data"
		headers := map[string]string{"x": "y"}

		err := qb.SendRawAsync(ctx, "route", payload, headers)
		if err == nil {
			t.Errorf("expected error, got nil")
		}
	})
}
