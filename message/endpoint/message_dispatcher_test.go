package endpoint_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hex-api-go/pkg/core/infrastructure/gomes/container"
	"github.com/hex-api-go/pkg/core/infrastructure/gomes/message"
	"github.com/hex-api-go/pkg/core/infrastructure/gomes/message/endpoint"
)

type dummyHandler struct{}

func (d *dummyHandler) Handle(_ context.Context, msg *message.Message) (*message.Message, error) {
	if msg.GetPayload() == "payload error" {
		return nil, fmt.Errorf("payload error")
	}
	return msg, nil
}

func TestNewMessageDispatcherBuilder(t *testing.T) {
	t.Parallel()
	builder := endpoint.NewMessageDispatcherBuilder("ref", "channel")
	if builder == nil {
		t.Error("NewMessageDispatcherBuilder should return a non-nil instance")
	}
}

func TestNewMessageDispatcherBuilder_Build(t *testing.T) {
	t.Run("should build a new MessageDispatcher successfully", func(t *testing.T) {
		t.Parallel()
		c := container.NewGenericContainer[any, any]()
		builder := endpoint.NewMessageDispatcherBuilder("ref", "channel")
		dispatcher, err := builder.Build(c)
		if err != nil {
			t.Errorf("Build should return nil error, got: %v", err)
		}
		if dispatcher == nil {
			t.Error("Build should return a non-nil instance")
		}
	})
	t.Run("should build a new MessageDispatcher with error", func(t *testing.T) {
		t.Parallel()
		c := container.NewGenericContainer[any, any]()
		builder := endpoint.NewMessageDispatcherBuilder("", "")
		dispatcher, err := builder.Build(c)
		if err != nil {
			t.Errorf("Build should return nil error, got: %v", err)
		}
		if dispatcher == nil {
			t.Error("Build should return a non-nil instance")
		}
	})
}

func TestNewMessageDispatcher(t *testing.T) {
	t.Parallel()
	gw := endpoint.NewGateway(&dummyHandler{}, "", "channel")
	dispatcher := endpoint.NewMessageDispatcher(gw)
	if dispatcher == nil {
		t.Error("NewMessageDispatcher should return a non-nil instance")
	}
}

func TestSendMessage(t *testing.T) {
	gw := endpoint.NewGateway(&dummyHandler{}, "", "channel")
	dispatcher := endpoint.NewMessageDispatcher(gw)
	t.Run("should send with success", func(t *testing.T) {
		t.Parallel()
		msg := message.NewMessageBuilder().
			WithChannelName("channel").
			WithMessageType(message.Command).
			WithPayload("payload").
			Build()
		resp, err := dispatcher.SendMessage(context.Background(), msg)
		if err != nil {
			t.Errorf("SendMessage should return nil error, got: %v", err)
		}
		if resp == nil {
			t.Error("SendMessage should return a response, got: nil")
		}
	})

	t.Run("should send message with error", func(t *testing.T) {
		t.Parallel()
		msg := message.NewMessageBuilder().
			WithChannelName("channel").
			WithMessageType(message.Command).
			WithPayload("payload error").
			Build()
		resp, err := dispatcher.SendMessage(context.Background(), msg)
		if err == nil {
			t.Errorf("SendMessage should return error, got: nil")
		}
		if resp != nil {
			t.Error("SendMessage should nil response, got: %w", resp)
		}
	})
}
func TestPublishMessage(t *testing.T) {
	gw := endpoint.NewGateway(&dummyHandler{}, "", "channel")
	dispatcher := endpoint.NewMessageDispatcher(gw)
	t.Run("should publish message without error", func(t *testing.T) {
		t.Parallel()
		msg := message.NewMessageBuilder().
			WithChannelName("channel").
			WithMessageType(message.Command).
			WithPayload("payload").
			Build()
		err := dispatcher.PublishMessage(context.Background(), msg)
		if err != nil {
			t.Errorf("PublishMessage should return nil error, got: %v", err)
		}
	})

	t.Run("should publish message with error", func(t *testing.T) {
		t.Parallel()
		msg := message.NewMessageBuilder().
			WithChannelName("channel").
			WithMessageType(message.Command).
			WithPayload("payload error").
			Build()
		err := dispatcher.PublishMessage(context.Background(), msg)
		if err == nil {
			t.Error("PublishMessage should return error if payload is invalid")
		}
	})
}
