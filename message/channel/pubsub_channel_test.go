package channel_test

import (
	"context"
	"testing"
	"time"

	"github.com/hex-api-go/pkg/core/infrastructure/gomes/message"
	"github.com/hex-api-go/pkg/core/infrastructure/gomes/message/channel"
)

func TestNewPubSubChannel(t *testing.T) {
	t.Parallel()
	ch := channel.NewPubSubChannel("chan1")
	if ch == nil {
		t.Error("NewPubSubChannel should return a non-nil instance")
	}
	if ch.Name() != "chan1" {
		t.Error("Channel name should be set correctly")
	}
}

func TestPubSub_Send(t *testing.T) {
	t.Run("should send message successfully", func(t *testing.T) {
		t.Parallel()
		ch := channel.NewPubSubChannel("chan1")
		ctx := context.Background()
		msg := &message.Message{}
		ch.Subscribe()
		ch.Send(ctx, msg)
		t.Cleanup(func() {
			ch.Unsubscribe()
		})
	})

	t.Run("should error when send message with context cancel", func(t *testing.T) {
		t.Parallel()
		ch := channel.NewPubSubChannel("chan1")
		msg := &message.Message{}
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		err := ch.Send(ctx, msg)
		if err.Error() != "context cancelled while sending message: context canceled" {
			t.Errorf("Send should return nil error, got: %v", err)
		}
		t.Cleanup(func() {
			ch.Unsubscribe()
		})
	})

	t.Run("should error when channel is closed", func(t *testing.T) {
		t.Parallel()
		ch := channel.NewPubSubChannel("chan1")
		ch.Unsubscribe()
		msg := &message.Message{}
		ctx := context.Background()
		err := ch.Send(ctx, msg)
		if err == nil {
			t.Error("Send should return error if channel is closed")
		}

		t.Cleanup(func() {
			ch.Unsubscribe()
		})
	})
}

func TestPubSub_Subscribe(t *testing.T) {
	t.Run("should receive message successfully", func(t *testing.T) {
		t.Parallel()
		ch := channel.NewPubSubChannel("chan1")
		msg := &message.Message{}
		ctx := context.Background()
		processed := make(chan bool)
		ch.Subscribe(func(m *message.Message) {
			if m == msg {
				processed <- true
			}
		})
		time.Sleep(100 * time.Millisecond)
		ch.Send(ctx, msg)
		<-processed
		ch.Unsubscribe()
	})

	t.Run("should stop when channel is closed", func(t *testing.T) {
		t.Parallel()
		ch := channel.NewPubSubChannel("chan1")
		ctx := context.Background()
		msg := &message.Message{}
		ch.Subscribe()
		time.Sleep(100 * time.Millisecond)
		ch.Send(ctx, msg)
		ch.Unsubscribe()
	})
}
