package adapter_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jeffersonbrasilino/gomes/message"
	"github.com/jeffersonbrasilino/gomes/message/adapter"
	"github.com/jeffersonbrasilino/gomes/message/channel"
)

// mockPublisherChannel implements message.PublisherChannel for tests.
type mockPublisherChannel struct {
	sendErr error
	sentMsg *message.Message
}

func (m *mockPublisherChannel) Send(ctx context.Context, msg *message.Message) error {
	m.sentMsg = msg
	return m.sendErr
}

func (m *mockPublisherChannel) Name() string {
	return "mockPublisherChannel"
}

func (m *mockPublisherChannel) Close() error {
	return nil
}

// mockOutboundMessageHandler implements message.MessageHandler for tests.
type mockOutboundMessageHandler struct{}

func (m mockOutboundMessageHandler) Handle(ctx context.Context, msg *message.Message) (*message.Message, error) {
	return msg, nil
}

// mockOutboundTranslator implements adapter.OutboundChannelMessageTranslator for tests.
type mockOutboundTranslator struct{}

func (m *mockOutboundTranslator) FromMessage(msg *message.Message) (string, error) {
	return "translated", nil
}

func TestOutboundChannelAdapterBuilder_ReferenceName(t *testing.T) {
	t.Parallel()
	translator := &mockOutboundTranslator{}
	builder := adapter.NewOutboundChannelAdapterBuilder("ref", "chan", translator)
	builder = builder.WithReferenceName("newref")
	if builder.ReferenceName() != "newref" {
		t.Errorf("Expected ReferenceName 'newref', got '%s'", builder.ReferenceName())
	}
}

func TestOutboundChannelAdapterBuilder_ChannelName(t *testing.T) {
	t.Parallel()
	translator := &mockOutboundTranslator{}
	builder := adapter.NewOutboundChannelAdapterBuilder("ref", "chan", translator)
	builder.WithChannelName("newchan")
	if builder.ChannelName() != "newchan" {
		t.Errorf("Expected ChannelName 'newchan', got '%s'", builder.ChannelName())
	}
}
func TestOutboundChannelAdapterBuilder_MessageTranslator(t *testing.T) {
	t.Parallel()
	translator := &mockOutboundTranslator{}
	builder := adapter.NewOutboundChannelAdapterBuilder("ref", "chan", translator)
	builder.WithMessageTranslator(translator)
	if builder.MessageTranslator() != translator {
		t.Error("MessageTranslator not assigned correctly")
	}
}

func TestOutboundChannelAdapterBuilder_WithReferenceName(t *testing.T) {
	t.Parallel()
	translator := &mockOutboundTranslator{}
	builder := adapter.NewOutboundChannelAdapterBuilder("ref", "chan", translator)
	builder = builder.WithReferenceName("newref")
	if builder.ReferenceName() != "newref" {
		t.Errorf("Expected ReferenceName 'newref', got '%s'", builder.ReferenceName())
	}
}

func TestOutboundChannelAdapterBuilder_WithChannelName(t *testing.T) {
	t.Parallel()
	translator := &mockOutboundTranslator{}
	builder := adapter.NewOutboundChannelAdapterBuilder("ref", "chan", translator)
	builder = builder.WithChannelName("newchan")
	if builder.ChannelName() != "newchan" {
		t.Errorf("Expected ChannelName 'newchan', got '%s'", builder.ChannelName())
	}
}

func TestOutboundChannelAdapterBuilder_WithMessageTranslator(t *testing.T) {
	t.Parallel()
	translator := &mockOutboundTranslator{}
	builder := adapter.NewOutboundChannelAdapterBuilder("ref", "chan", translator)
	builder.WithMessageTranslator(translator)
	if builder.MessageTranslator() != translator {
		t.Error("MessageTranslator not assigned correctly")
	}
}

func TestOutboundChannelAdapterBuilder_WithReplyChannelName(t *testing.T) {
	t.Parallel()
	translator := &mockOutboundTranslator{}
	builder := adapter.NewOutboundChannelAdapterBuilder("ref", "chan", translator)
	builder.WithReplyChannelName("replychan")
	if builder.ReplyChannelName("") != "replychan" {
		t.Errorf("Expected ReplyChannelName 'replychan', got '%s'", builder.ReplyChannelName(""))
	}
}

func TestOutboundChannelAdapterBuilder_WithBeforeInterceptors(t *testing.T) {
	t.Parallel()
	translator := &mockOutboundTranslator{}
	builder := adapter.NewOutboundChannelAdapterBuilder("ref", "chan", translator)
	before := &mockOutboundMessageHandler{}
	expect := builder.WithBeforeInterceptors(before)
	if expect != builder {
		t.Error("BeforeProcessors not assigned correctly")
	}
}

func TestOutboundChannelAdapterBuilder_WithAfterInterceptors(t *testing.T) {
	t.Parallel()
	translator := &mockOutboundTranslator{}
	builder := adapter.NewOutboundChannelAdapterBuilder("ref", "chan", translator)
	after := &mockOutboundMessageHandler{}
	expect := builder.WithAfterInterceptors(after)
	if expect != builder {
		t.Error("AfterProcessors not assigned correctly")
	}
}
func TestOutboundChannelAdapterBuilder_BuildOutboundAdapter(t *testing.T) {
	t.Parallel()
	translator := &mockOutboundTranslator{}
	builder := adapter.NewOutboundChannelAdapterBuilder("ref", "chan", translator)
	pubChan := &mockPublisherChannel{}
	chn, err := builder.BuildOutboundAdapter(pubChan)
	if err != nil {
		t.Errorf("Expected success, got error: %v", err)
	}
	if chn == nil {
		t.Error("Expected channel instance, got nil")
	}
}

func TestOutboundChannelAdapter_Handle(t *testing.T) {
	msg := message.NewMessageBuilder().
		WithChannelName("channel").
		WithMessageType(message.Command).
		WithPayload("payload").
		Build()
	pubChan := &mockPublisherChannel{}
	adapterInstance := adapter.NewOutboundChannelAdapter(pubChan)
	ctx := context.Background()
	t.Run("success with payload", func(t *testing.T) {
		t.Parallel()
		msg := message.NewMessageBuilder().
			WithChannelName("channel").
			WithMessageType(message.Command).
			WithPayload("payload").
			Build()
		pubChan := &mockPublisherChannel{}
		adapterInstance := adapter.NewOutboundChannelAdapter(pubChan)
		replyChan := channel.NewPointToPointChannel("reply")
		msg.GetHeaders().ReplyChannel = replyChan

		done := make(chan struct{})
		replyChan.Subscribe(func(m *message.Message) {
			defer close(done)
			if m.GetPayload() != "payload" {
				t.Errorf("Expected payload 'payload', got '%v'", m.GetPayload())
			}
		})

		adapterInstance.Send(context.Background(), msg)
		select {
		case <-done:
			// Sucesso: a mensagem chegou a tempo
		case <-time.After(2 * time.Second):
			t.Fatal("Test timed out waiting for reply")
		}
		t.Cleanup(func() {
			replyChan.Close()
		})
	})

	t.Run("success without payload", func(t *testing.T) {
		t.Parallel()
		msg := message.NewMessageBuilder().
			WithChannelName("channel").
			WithMessageType(message.Command).
			WithPayload(nil).
			Build()
		pubChan := &mockPublisherChannel{}
		adapterInstance := adapter.NewOutboundChannelAdapter(pubChan)
		replyChan := channel.NewPointToPointChannel("reply")
		msg.GetHeaders().ReplyChannel = replyChan

		done := make(chan struct{})
		replyChan.Subscribe(func(m *message.Message) {
			defer close(done)
			if m.GetPayload() != nil {
				t.Errorf("Expected payload 'nil', got '%v'", m.GetPayload())
			}
		})
		adapterInstance.Send(context.Background(), msg)

		select {
		case <-done:
			// Sucesso: a mensagem chegou a tempo
		case <-time.After(2 * time.Second):
			t.Fatal("Test timed out waiting for reply")
		}
		t.Cleanup(func() {
			replyChan.Close()
		})
	})
	t.Run("send error", func(t *testing.T) {
		t.Parallel()
		pubChan.sendErr = errors.New("send error")
		err := adapterInstance.Send(ctx, msg)
		if err == nil {
			t.Error("Expected error from publisher, got nil")
		}
		pubChan.sendErr = nil
	})
}
