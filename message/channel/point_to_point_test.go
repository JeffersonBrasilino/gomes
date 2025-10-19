package channel_test

import (
	"context"
	"testing"

	"github.com/jeffersonbrasilino/gomes/message"
	"github.com/jeffersonbrasilino/gomes/message/channel"
)

func TestPointToPointReferenceName(t *testing.T) {
	t.Parallel()
	name := "test"
	reference := channel.PointToPointReferenceName(name)
	if reference != "point-to-point-channel:test" {
		t.Errorf("Expected reference name 'point-to-point-channel:test', got: %s", reference)
	}
}

func TestNewPointToPointChannel(t *testing.T) {
	t.Parallel()
	ch := channel.NewPointToPointChannel("chan1")
	if ch == nil {
		t.Error("NewPointToPointChannel should return a non-nil instance")
	}
	if ch.Name() != "chan1" {
		t.Error("Channel name should be set correctly")
	}
	t.Cleanup(func() {
		ch.Close()
	})
}

func TestPointToPoint_Send(t *testing.T) {
	t.Run("should send message successfully", func(t *testing.T) {
		t.Parallel()
		msg := &message.Message{}
		ctx := context.Background()
		ch := channel.NewPointToPointChannel("chan1")
		go ch.Send(ctx, msg)
		ch.Receive(ctx)
		t.Cleanup(func() {
			ch.Close()
		})
	})
	t.Run("should error when send message with context cancel", func(t *testing.T) {
		t.Parallel()
		ch := channel.NewPointToPointChannel("chan1")
		msg := &message.Message{}
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		err := ch.Send(ctx, msg)
		if err.Error() != "context cancelled while sending message: context canceled" {
			t.Errorf("Send should return nil error, got: %v", err)
		}
		t.Cleanup(func() {
			ch.Close()
		})
	})

	t.Run("shoud channel has been closed", func(t *testing.T) {
		t.Parallel()
		ch := channel.NewPointToPointChannel("chan1")
		msg := &message.Message{}
		ctx := context.Background()
		ch.Close()
		err := ch.Send(ctx, msg)
		if err.Error() != "channel has not been opened" {
			t.Error("Send should return error if channel is closed")
		}
	})
}

func TestPointToPoint_Subscribe(t *testing.T) {
	t.Run("should receive message successfully", func(t *testing.T) {
		t.Parallel()
		ch := channel.NewPointToPointChannel("chan1")
		msg := &message.Message{}
		processed := make(chan bool)
		ch.Subscribe(func(m *message.Message) {
			if m == msg {
				processed <- true
			}
		})
		ch.Send(context.Background(), msg)
		<-processed
		t.Cleanup(func() {
			ch.Close()
		})
	})

	t.Run("should stop when channel is closed", func(t *testing.T) {
		t.Parallel()
		ch := channel.NewPointToPointChannel("chan1")
		msg := &message.Message{}
		processed := make(chan bool)
		ch.Subscribe(func(m *message.Message) {
			if m == msg {
				processed <- true
			}
		})
		ch.Send(context.Background(), msg)
		ch.Close()
	})
}

func TestReceive(t *testing.T) {
	t.Run("should receive message successfully", func(t *testing.T) {
		t.Parallel()
		ch := channel.NewPointToPointChannel("chan1")
		msg := &message.Message{}
		go func() {
			ch.Send(context.Background(), msg)
		}()
		receivedMsg, err := ch.Receive(context.TODO())
		if err != nil {
			t.Error("Receive should not return an error")
		}
		if receivedMsg != msg {
			t.Error("Receive should return the sent message")
		}
		t.Cleanup(func() {
			ch.Close()
		})
	})

	t.Run("should error when channel is closed", func(t *testing.T) {
		t.Parallel()
		ch := channel.NewPointToPointChannel("chan1")
		ch.Close()
		_, err := ch.Receive(context.TODO())
		if err.Error() != "channel has not been opened" {
			t.Error("Receive should return error if channel is closed")
		}
	})
}
