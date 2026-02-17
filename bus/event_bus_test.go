package bus_test

import (
	"context"
	"errors"
	"testing"

	"github.com/jeffersonbrasilino/gomes/bus"
	"github.com/jeffersonbrasilino/gomes/message"
)

type mockEventDispatcher struct {
	lastMsg    *message.Message
	publishErr error
}

func (m *mockEventDispatcher) PublishMessage(ctx context.Context, msg *message.Message) error {
	m.lastMsg = msg
	return m.publishErr
}
func (m *mockEventDispatcher) SendMessage(ctx context.Context, msg *message.Message) (any, error) {
	m.lastMsg = msg
	return nil, nil
}

func (c *mockEventDispatcher) MessageBuilder(
	messageType message.MessageType,
	payload any,
	headers map[string]string,
) *message.MessageBuilder {
	builder, _ := message.NewMessageBuilderFromHeaders(headers)
	builder.WithMessageType(messageType)
	builder.WithPayload(payload)
	if val, ok := headers[message.HeaderCorrelationId]; !ok || val == "" {
		builder.WithCorrelationId("123445")
	}

	return builder
}

// Mock action for handler.Action
type mockEAction struct {
	name string
}

func (a mockEAction) Name() string { return a.name }

func TestEventBus_Publish(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		t.Parallel()
		dispatcher := &mockEventDispatcher{publishErr: nil}
		eb := bus.NewEventBus(dispatcher)
		ctx := context.Background()
		action := mockEAction{name: "TestEvent"}

		err := eb.Publish(ctx, action)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if dispatcher.lastMsg == nil {
			t.Errorf("expected message to be published")
		}
	})

	t.Run("error", func(t *testing.T) {
		t.Parallel()
		dispatcher := &mockEventDispatcher{publishErr: errors.New("fail")}
		eb := bus.NewEventBus(dispatcher)
		ctx := context.Background()
		action := mockEAction{name: "TestEvent"}

		err := eb.Publish(ctx, action)
		if err == nil {
			t.Errorf("expected error, got nil")
		}
	})
}

func TestEventBus_PublishRaw(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		t.Parallel()
		dispatcher := &mockEventDispatcher{publishErr: nil}
		eb := bus.NewEventBus(dispatcher)
		ctx := context.Background()
		payload := "data"
		headers := map[string]string{"x": "y"}

		err := eb.PublishRaw(ctx, "route", payload, headers)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if dispatcher.lastMsg == nil {
			t.Errorf("expected message to be published")
		}
	})

	t.Run("error", func(t *testing.T) {
		t.Parallel()
		dispatcher := &mockEventDispatcher{publishErr: errors.New("fail")}
		eb := bus.NewEventBus(dispatcher)
		ctx := context.Background()
		payload := "data"
		headers := map[string]string{"x": "y"}

		err := eb.PublishRaw(ctx, "route", payload, headers)
		if err == nil {
			t.Errorf("expected error, got nil")
		}
	})
}
