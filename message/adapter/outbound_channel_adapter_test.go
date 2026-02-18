package adapter_test

import (
	"context"
	"errors"
	"testing"

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

type mockClosableChannel struct {
	*mockPublisherChannel
}

func (m *mockClosableChannel) Close() error {
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
	if builder.ReferenceName() != "ref" {
		t.Errorf("Expected ReferenceName 'newref', got '%s'", builder.ReferenceName())
	}
}

func TestOutboundChannelAdapterBuilder_ChannelName(t *testing.T) {
	t.Parallel()
	translator := &mockOutboundTranslator{}
	builder := adapter.NewOutboundChannelAdapterBuilder("ref", "chan", translator)
	if builder.ChannelName() != "chan" {
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

func TestOutboundChannelAdapterBuilder_BuildOutboundAdapter(t *testing.T) {
	t.Parallel()
	translator := &mockOutboundTranslator{}
	builder := adapter.NewOutboundChannelAdapterBuilder("ref", "chan", translator)
	builder.WithReplyChannelName("replychan")
	pubChan := &mockPublisherChannel{}
	chn, err := builder.BuildOutboundAdapter(pubChan)
	if err != nil {
		t.Errorf("Expected success, got error: %v", err)
	}
	if chn == nil {
		t.Error("Expected channel instance, got nil")
	}
}

func TestOutboundChannelAdapter_Name(t *testing.T) {
	t.Parallel()
	translator := &mockOutboundTranslator{}
	builder := adapter.NewOutboundChannelAdapterBuilder("ref", "chan", translator)
	buiderInstance, _ := builder.BuildOutboundAdapter(&mockPublisherChannel{})
	if buiderInstance.Name() != "mockPublisherChannel" {
		t.Errorf("Expected Name 'mockPublisherChannel', got '%s'", buiderInstance.Name())
	}
}

func TestOutboundChannelAdapter_Send(t *testing.T) {
	t.Run("success with payload", func(t *testing.T) {
		t.Parallel()
		msg := message.NewMessageBuilder().
			WithChannelName("channel").
			WithMessageType(message.Command).
			WithPayload("payload").
			Build()
		pubChan := &mockPublisherChannel{}
		adapterInstance := adapter.NewOutboundChannelAdapter(pubChan, "replyChann")

		got := adapterInstance.Send(context.Background(), msg)
		if got != nil {
			t.Errorf("expected nil, got %v", got)
		}
	})

	t.Run("success with Send Internal Channel", func(t *testing.T) {
		t.Parallel()

		internalChannel := channel.NewPointToPointChannel("internalChan")
		msg := message.NewMessageBuilder().
			WithChannelName("channel").
			WithMessageType(message.Command).
			WithPayload("payload").
			WithInternalReplyChannel(internalChannel).
			Build()
		pubChan := &mockPublisherChannel{}

		adapterInstance := adapter.NewOutboundChannelAdapter(pubChan, "replyChann")

		go internalChannel.Receive(context.Background())

		got := adapterInstance.Send(context.Background(), msg)
		if got != nil {
			t.Errorf("expected nil, got %v", got)
		}

		t.Cleanup(func() {
			adapterInstance.Close()
		})
	})

	t.Run("send error", func(t *testing.T) {
		t.Parallel()
		pubChan := &mockPublisherChannel{}

		adapterInstance := adapter.NewOutboundChannelAdapter(pubChan, "replyChann")
		ctx := context.Background()

		internalChannel := channel.NewPointToPointChannel("internalChan")
		go internalChannel.Receive(context.Background())

		msg := message.NewMessageBuilder().
			WithChannelName("channel").
			WithMessageType(message.Command).
			WithPayload("payload").
			WithInternalReplyChannel(internalChannel).
			Build()
		pubChan.sendErr = errors.New("send error")
		err := adapterInstance.Send(ctx, msg)
		if err == nil {
			t.Error("Expected error from publisher, got nil")
		}
		pubChan.sendErr = nil

		t.Cleanup(func() {
			adapterInstance.Close()
		})
	})
}

func TestOutboundChannelAdapter_Close(t *testing.T) {
	t.Run("Close closable channel", func(t *testing.T) {
		t.Parallel()
		pubChan := &mockClosableChannel{&mockPublisherChannel{}}
		adapterInstance := adapter.NewOutboundChannelAdapter(pubChan, "replyChann")
		err := adapterInstance.Close()
		if err != nil {
			t.Errorf("Expected nil, got %v", err)
		}
	})

	t.Run("Close unclosable channel", func(t *testing.T) {
		t.Parallel()
		pubChan := &mockPublisherChannel{}
		adapterInstance := adapter.NewOutboundChannelAdapter(pubChan, "replyChann")
		err := adapterInstance.Close()
		if err != nil {
			t.Errorf("Expected nil, got %v", err)
		}
	})
}
