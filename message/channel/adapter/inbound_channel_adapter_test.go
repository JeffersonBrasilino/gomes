package adapter_test

import (
	"context"
	"errors"
	"testing"

	"github.com/jeffersonbrasilino/gomes/message"
	"github.com/jeffersonbrasilino/gomes/message/channel/adapter"
)

// mockConsumerChannel implements message.ConsumerChannel for tests.
type mockConsumerChannel struct {
	msg      *message.Message
	err      error
	closeErr error
}

func (m *mockConsumerChannel) Receive(ctx context.Context) (*message.Message, error) {
	return m.msg, m.err
}

func (m *mockConsumerChannel) Close() error {
	return m.closeErr
}

func (m mockConsumerChannel) Name() string {
	return "mockChannel"
}

type mockMessageHandler struct{}

func (m mockMessageHandler) Handle(ctx context.Context, msg *message.Message) (*message.Message, error) {
	return msg, nil
}

// mockTranslator implements InboundChannelMessageTranslator for tests.
type mockTranslator struct{}

func (m *mockTranslator) ToMessage(msg string) (*message.Message, error) {
	return message.NewMessageBuilder().
		WithChannelName("channel").
		WithMessageType(message.Command).
		WithPayload(msg).
		Build(), nil
}

func TestNewInboundChannelAdapterBuilder(t *testing.T) {
	t.Parallel()
	translator := &mockTranslator{}
	builder := adapter.NewInboundChannelAdapterBuilder("ref", "chan", translator)
	if builder.ReferenceName() != "chan" {
		t.Errorf("Expected ChannelName 'chan', got '%s'", builder.ReferenceName())
	}
	if builder.MessageTranslator() != translator {
		t.Error("MessageTranslator not assigned correctly")
	}
}

func TestInboundChannelAdapterBuilder_WithDeadLetterChannelName(t *testing.T) {
	t.Parallel()
	translator := &mockTranslator{}
	builder := adapter.NewInboundChannelAdapterBuilder("ref", "chan", translator)
	builder.WithDeadLetterChannelName("dlc")
	b := builder.BuildInboundAdapter(&mockConsumerChannel{})
	if b.DeadLetterChannelName() != "dlc" {
		t.Errorf("Expected DeadLetterChannelName 'dlc', got '%s'", b.DeadLetterChannelName())
	}
}

func TestInboundChannelAdapterBuilder_WithBeforeInterceptors(t *testing.T) {
	t.Parallel()
	translator := &mockTranslator{}
	builder := adapter.NewInboundChannelAdapterBuilder("ref", "chan", translator)
	builder.WithBeforeInterceptors(&mockMessageHandler{})
	b := builder.BuildInboundAdapter(&mockConsumerChannel{})
	if len(b.BeforeProcessors()) != 1 {
		t.Error("BeforeProcessors not assigned correctly")
	}
}

func TestInboundChannelAdapterBuilder_WithAfterInterceptors(t *testing.T) {
	t.Parallel()
	translator := &mockTranslator{}
	builder := adapter.NewInboundChannelAdapterBuilder("ref", "chan", translator)
	builder.WithAfterInterceptors(&mockMessageHandler{})
	b := builder.BuildInboundAdapter(&mockConsumerChannel{})
	if len(b.AfterProcessors()) != 1 {
		t.Error("AfterProcessors not assigned correctly")
	}
}

func TestInboundChannelAdapterBuilder_BuildInboundAdapter(t *testing.T) {
	t.Parallel()
	translator := &mockTranslator{}
	builder := adapter.NewInboundChannelAdapterBuilder("ref", "chan", translator)
	mockChan := &mockConsumerChannel{}
	adapterInstance := builder.BuildInboundAdapter(mockChan)
	if adapterInstance.ReferenceName() != "ref" {
		t.Errorf("Expected ReferenceName 'ref', got '%s'", adapterInstance.ReferenceName())
	}
}

func TestInboundChannelAdapterBuilder_ReferenceName(t *testing.T) {
	t.Parallel()
	translator := &mockTranslator{}
	builder := adapter.NewInboundChannelAdapterBuilder("ref", "chan", translator)
	if builder.ReferenceName() != "chan" {
		t.Errorf("Expected ReferenceName 'chan', got '%s'", builder.ReferenceName())
	}
}

func TestInboundChannelAdapter_ReferenceName(t *testing.T) {
	t.Parallel()
	mockChan := &mockConsumerChannel{}
	adapterInstance := adapter.NewInboundChannelAdapter(mockChan, "ref", "dlc", nil, nil, nil)
	if adapterInstance.ReferenceName() != "ref" {
		t.Errorf("Expected ReferenceName 'ref', got '%s'", adapterInstance.ReferenceName())
	}
}

func TestInboundChannelAdapter_DeadLetterChannelName(t *testing.T) {
	t.Parallel()
	mockChan := &mockConsumerChannel{}
	adapterInstance := adapter.NewInboundChannelAdapter(mockChan, "ref", "dlc", nil, nil, nil)
	if adapterInstance.DeadLetterChannelName() != "dlc" {
		t.Errorf("Expected DeadLetterChannelName 'dlc', got '%s'", adapterInstance.DeadLetterChannelName())
	}
}

func TestInboundChannelAdapter_BeforeProcessors(t *testing.T) {
	t.Parallel()
	mockChan := &mockConsumerChannel{}
	beforeHandlers := []message.MessageHandler{&mockMessageHandler{}}
	adapterInstance := adapter.NewInboundChannelAdapter(mockChan, "ref", "dlc", beforeHandlers, nil, nil)
	if len(adapterInstance.BeforeProcessors()) != 1 {
		t.Error("BeforeProcessors not assigned correctly")
	}
}

func TestInboundChannelAdapter_AfterProcessors(t *testing.T) {
	t.Parallel()
	mockChan := &mockConsumerChannel{}
	afterHandlers := []message.MessageHandler{&mockMessageHandler{}}
	adapterInstance := adapter.NewInboundChannelAdapter(mockChan, "ref", "dlc", nil, afterHandlers, nil)
	if len(adapterInstance.AfterProcessors()) != 1 {
		t.Error("AfterProcessors not assigned correctly")
	}
}

func TestInboundChannelAdapter_ReceiveMessage(t *testing.T) {
	msg := message.NewMessageBuilder().
		WithChannelName("channel").
		WithMessageType(message.Command).
		WithPayload("payload").
		Build()
	t.Run("success", func(t *testing.T) {
		t.Parallel()
		mockChan := &mockConsumerChannel{msg: msg}
		adapterInstance := adapter.NewInboundChannelAdapter(mockChan, "ref", "dlc", nil, nil, nil)
		ctx := context.Background()
		m, err := adapterInstance.ReceiveMessage(ctx)
		if err != nil {
			t.Errorf("Expected success, got error: %v", err)
		}
		if m != msg {
			t.Error("Received message does not match")
		}
	})
	t.Run("context cancel", func(t *testing.T) {
		t.Parallel()
		mockChan := &mockConsumerChannel{}
		adapterInstance := adapter.NewInboundChannelAdapter(mockChan, "ref", "dlc", nil, nil, nil)
		ctxCancel, cancel := context.WithCancel(context.Background())
		cancel()
		m, err := adapterInstance.ReceiveMessage(ctxCancel)
		if err == nil {
			t.Error("Expected context canceled error")
		}
		if m != nil {
			t.Error("Message should be nil when context is canceled")
		}
	})
}

func TestClose(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		t.Parallel()
		mockChan := &mockConsumerChannel{closeErr: nil}
		adapterInstance := adapter.NewInboundChannelAdapter(mockChan, "ref", "dlc", nil, nil, nil)
		if err := adapterInstance.Close(); err != nil {
			t.Errorf("Expected success on close, got error: %v", err)
		}
	})
	t.Run("error", func(t *testing.T) {
		t.Parallel()
		errClose := errors.New("erro ao fechar")
		mockChan2 := &mockConsumerChannel{closeErr: errClose}
		adapterInstance2 := adapter.NewInboundChannelAdapter(mockChan2, "ref", "dlc", nil, nil, nil)
		if err := adapterInstance2.Close(); err != errClose {
			t.Errorf("Expected close error, got: %v", err)
		}
	})
}
