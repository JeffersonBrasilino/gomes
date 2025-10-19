package endpoint_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hex-api-go/pkg/core/infrastructure/gomes/container"
	"github.com/hex-api-go/pkg/core/infrastructure/gomes/message"
	"github.com/hex-api-go/pkg/core/infrastructure/gomes/message/channel"
	"github.com/hex-api-go/pkg/core/infrastructure/gomes/message/endpoint"
)

type dummyGatewayHandler struct{}

func (d *dummyGatewayHandler) Handle(_ context.Context, msg *message.Message) (*message.Message, error) {
	if msg.GetPayload() == "payload error" {
		return nil, fmt.Errorf("payload error")
	}
	if msg.GetPayload() == "invalid payload" {
		return nil, nil
	}
	return msg, nil
}

func TestGatewayReferenceName(t *testing.T) {
	t.Parallel()
	refName := "myGateway"
	expected := "gateway:myGateway"
	result := endpoint.GatewayReferenceName(refName)
	if result != expected {
		t.Errorf("GatewayReferenceName(%s) = %s; want %s", refName, result, expected)
	}
}

func TestNewGatewayBuilder(t *testing.T) {
	t.Parallel()
	builder := endpoint.NewGatewayBuilder("ref", "channel")
	if builder == nil {
		t.Error("NewGatewayBuilder should return a non-nil instance")
	}
}

func TestMessageBuilder_WithBeforeInterceptors(t *testing.T) {
	t.Parallel()
	t.Run("should add before interceptors correctly", func(t *testing.T) {
		container := container.NewGenericContainer[any, any]()
		interceptor := &dummyGatewayHandler{}
		result, err := endpoint.NewGatewayBuilder("ref", "channel").
			WithBeforeInterceptors(interceptor).
			Build(container)
		if err != nil {
			t.Errorf("Build should return nil error, got: %v", err)
		}
		if result == nil {
			t.Error("WithBeforeInterceptors should add the interceptor")
		}
	})
}
func TestMessageBuilder_ReferenceName(t *testing.T) {
	t.Parallel()
	t.Run("should return the correct reference name", func(t *testing.T) {
		referenceName := "myGateway"
		expected := "gateway:myGateway"
		builder := endpoint.NewGatewayBuilder(referenceName, "channel")
		if expected != builder.ReferenceName() {
			t.Errorf("ReferenceName() = %s; want %s", expected, builder.ReferenceName())
		}
	})
}

func TestMessageBuilder_WithAfterInterceptors(t *testing.T) {
	t.Parallel()
	t.Run("should add after interceptors correctly", func(t *testing.T) {
		container := container.NewGenericContainer[any, any]()
		interceptor := &dummyGatewayHandler{}
		result, err := endpoint.NewGatewayBuilder("ref", "channel").
			WithAfterInterceptors(interceptor).
			Build(container)
		if err != nil {
			t.Errorf("Build should return nil error, got: %v", err)
		}
		if result == nil {
			t.Error("WithAfterInterceptors should add the interceptor")
		}
	})
}
func TestMessageBuilder_WithDeadLetterChannel(t *testing.T) {
	t.Parallel()
	t.Run("should add dead letter channel correctly", func(t *testing.T) {
		container := container.NewGenericContainer[any, any]()
		dlq := channel.NewPointToPointChannel("deadLetterChannel")
		container.Set("deadLetterChannel", dlq)
		result, err := endpoint.NewGatewayBuilder("ref", "channel").
			WithDeadLetterChannel("deadLetterChannel").
			Build(container)
		if err != nil {
			t.Errorf("Build should return nil error, got: %v", err)
		}
		if result == nil {
			t.Error("WithDeadLetterChannel should add the dead letter channel")
		}

		t.Cleanup(func() {
			dlq.Close()
		})
	})

	t.Run("should return error if dead letter channel does not exist", func(t *testing.T) {
		container := container.NewGenericContainer[any, any]()
		_, err := endpoint.NewGatewayBuilder("ref", "channel").
			WithDeadLetterChannel("nonExistentChannel").
			Build(container)
		if err == nil {
			t.Error("Build should return an error if dead letter channel does not exist")
		}
	})
}
func TestMessageBuilder_WithReplyChannel(t *testing.T) {
	t.Parallel()
	t.Run("should add reply channel correctly", func(t *testing.T) {
		container := container.NewGenericContainer[any, any]()
		result, err := endpoint.NewGatewayBuilder("ref", "channel").
			WithReplyChannel("replyChannel").
			Build(container)
		if err != nil {
			t.Errorf("Build should return nil error, got: %v", err)
		}
		if result == nil {
			t.Error("WithReplyChannel should add the reply channel")
		}
	})
}
func TestNewGateway(t *testing.T) {
	t.Parallel()
	gw := endpoint.NewGateway(&dummyGatewayHandler{}, "ref", "channel")
	if gw == nil {
		t.Error("NewGateway should return a non-nil instance")
	}
}
func TestGateway_Execute(t *testing.T) {
	gw := endpoint.NewGateway(&dummyGatewayHandler{}, "ref", "channel")
	t.Run("should execute the message successfully", func(t *testing.T) {
		t.Parallel()
		msg := message.NewMessageBuilder().
			WithChannelName("channel").
			WithMessageType(message.Command).
			WithPayload("payload").
			Build()
		ctx := context.Background()
		res, err := gw.Execute(ctx, msg)
		if err != nil {
			t.Error("Execute should return a non-nil error, got:", err)
		}
		if res == nil {
			t.Error("Execute should return a non-nil result, got: nil")
		}
	})

	t.Run("should error when execute the message", func(t *testing.T) {
		t.Parallel()
		msg := message.NewMessageBuilder().
			WithChannelName("channel").
			WithMessageType(message.Command).
			WithPayload("payload error").
			Build()
		ctx := context.Background()
		res, err := gw.Execute(ctx, msg)
		if err == nil {
			t.Error("Execute should return a non-nil error, got: nil")
		}
		if res != nil {
			t.Error("Execute should return a nil result, got:", res)
		}
	})

	t.Run("should cancel the execution when context is done", func(t *testing.T) {
		t.Parallel()
		msg := message.NewMessageBuilder().
			WithChannelName("channel").
			WithMessageType(message.Command).
			WithPayload("payload").
			Build()
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		res, err := gw.Execute(ctx, msg)
		if err.Error() != "context canceled" {
			t.Error("Execute should return a non-nil error due to context cancellation, got: nil")
		}
		if res != nil {
			t.Error("Execute should return a nil result due to context cancellation, got:", res)
		}
	})
}
