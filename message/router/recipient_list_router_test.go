package router

import (
	"context"
	"errors"
	"testing"

	"github.com/jeffersonbrasilino/gomes/container"
	"github.com/jeffersonbrasilino/gomes/message"
)

type dummyChannel struct {
	msgReceived chan *message.Message
	shouldError bool
}

func (d *dummyChannel) Send(_ context.Context, msg *message.Message) error {
	if d.shouldError {
		return errors.New("erro de canal")
	}
	d.msgReceived <- msg
	return nil
}

func (d *dummyChannel) Receive(_ context.Context) (*message.Message, error) {
	return <-d.msgReceived, nil
}

func (d *dummyChannel) Name() string {
	return "canal1"
}

func TestNewRecipientListRouter(t *testing.T) {
	t.Parallel()
	container := container.NewGenericContainer[any, any]()
	r := NewRecipientListRouter(container)
	if r == nil {
		t.Error("NewRecipientListRouter should return a non-nil instance")
	}
}

func TestHandle(t *testing.T) {
	msg := message.NewMessageBuilder().
		WithPayload("payload").
		WithRoute("rota1").
		WithMessageType(1).
		WithCorrelationId("cid").
		WithChannelName("canal1").
		WithReplyChannelName("rch").
		WithContext(context.Background()).
		Build()
	container := container.NewGenericContainer[any, any]()
	chn := make(chan *message.Message, 50)
	ch := &dummyChannel{msgReceived: chn}
	container.Set("canal1", ch)
	container.Set("rota1", ch)
	t.Cleanup(func() { close(chn) })
	t.Run("should route message to channel via ChannelName", func(t *testing.T) {
		t.Parallel()
		r := NewRecipientListRouter(container)
		result, err := r.Handle(context.Background(), msg)
		if err != nil {
			t.Errorf("Handle should return nil error, got: %v", err)
		}
		if result != msg {
			t.Error("Handle should return the original message if channel exists")
		}
		if <-ch.msgReceived != msg {
			t.Error("Channel should receive the sent message")
		}
	})
	t.Run("should route message to channel via Route", func(t *testing.T) {
		t.Parallel()
		r := NewRecipientListRouter(container)
		result, err := r.Handle(context.Background(), msg)
		if err != nil {
			t.Errorf("Handle should return nil error, got: %v", err)
		}
		if result != msg {
			t.Error("Handle should return the original message if route exists")
		}
		if <-ch.msgReceived != msg {
			t.Error("Channel should receive the sent message")
		}
	})
	t.Run("should return error if channel does not exist", func(t *testing.T) {
		t.Parallel()
		r := NewRecipientListRouter(container)
		msg := message.NewMessageBuilderFromMessage(msg).WithChannelName("dont_exists").Build()
		result, err := r.Handle(context.Background(), msg)
		if err == nil {
			t.Error("Handle should return error if channel does not exist")
		}
		if result != nil {
			t.Error("Handle should return nil if channel does not exist")
		}
	})
	t.Run("should use route when ChannelName is empty", func(t *testing.T) {
		t.Parallel()
		r := NewRecipientListRouter(container)
		msg := message.NewMessageBuilderFromMessage(msg).
			WithChannelName("").
			WithRoute("rota1").
			Build()
		_, err := r.Handle(context.Background(), msg)
		if err != nil {
			t.Errorf("Handle should return nil error, got: %v", err)
		}
		if <-ch.msgReceived != msg {
			t.Error("Channel referenced by ChannelName should receive the message")
		}
	})
}
