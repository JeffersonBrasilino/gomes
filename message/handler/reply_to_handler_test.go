package handler_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/jeffersonbrasilino/gomes/container"
	"github.com/jeffersonbrasilino/gomes/message"
	"github.com/jeffersonbrasilino/gomes/message/channel"
	"github.com/jeffersonbrasilino/gomes/message/handler"
)

type replyTohandlerMock struct{}

func (m *replyTohandlerMock) Handle(
	ctx context.Context,
	msg *message.Message,
) (*message.Message, error) {
	if payload, ok := msg.GetPayload().(string); ok && payload == "error" {
		return nil, fmt.Errorf("error processing message")
	}

	if _, ok := msg.GetPayload().(error); ok {
		responseMessage := message.NewMessageBuilderFromMessage(msg).
			WithPayload(fmt.Errorf("error processing message")).
			Build()
		return responseMessage, nil
	}

	responseMessage := message.NewMessageBuilderFromMessage(msg).
		WithPayload("response").
		Build()
	return responseMessage, nil
}

func TestReplyToHandler_Handle(t *testing.T) {
	t.Run("should be reply to success", func(t *testing.T) {
		t.Parallel()

		responseChannel := channel.NewPointToPointChannel("responseChannel")
		defer responseChannel.Close()

		container := container.NewGenericContainer[any, any]()
		container.Set("responseChannel", responseChannel)

		reqMessage := message.NewMessageBuilder().
			WithReplyTo("responseChannel").
			WithPayload("request").
			Build()

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		go responseChannel.Receive(ctx)
		got := handler.NewSendReplyToHandler(&replyTohandlerMock{}, container)
		result, err := got.Handle(ctx, reqMessage)

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if result == nil {
			t.Fatal("expected a response message, got nil")
		}

		if result.GetPayload() != "response" {
			t.Fatalf("expected payload 'response', got %v", result.GetPayload())
		}

	})

	t.Run("should be channel not specified", func(t *testing.T) {
		t.Parallel()

		reqMessage := message.NewMessageBuilder().
			WithPayload("request").
			Build()

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		container := container.NewGenericContainer[any, any]()
		got := handler.NewSendReplyToHandler(&replyTohandlerMock{}, container)
		result, err := got.Handle(ctx, reqMessage)
		if err == nil {
			t.Fatal("expected an error, got nil")
		}

		if result != nil {
			t.Fatalf("expected no response message, got %v", result)
		}
		if err.Error() != "[send-reply-to-handler] cannot send message: channel not specified" {
			t.Fatalf("expected error message '[send-reply-to-handler] cannot send message: channel not specified', got %v", err.Error())
		}
	})

	t.Run("should be channel not exists", func(t *testing.T) {
		t.Parallel()

		reqMessage := message.NewMessageBuilder().
			WithReplyTo("not exists").
			WithPayload("request").
			Build()

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		container := container.NewGenericContainer[any, any]()
		got := handler.NewSendReplyToHandler(&replyTohandlerMock{}, container)
		result, err := got.Handle(ctx, reqMessage)
		if err == nil {
			t.Fatal("expected an error, got nil")
		}

		if result != nil {
			t.Fatalf("expected no response message, got %v", result)
		}
		if err.Error() != "[send-reply-to-handler] cannot find item not exists" {
			t.Fatalf("expected error message '[send-reply-to-handler] cannot find item not exists', got %v", err.Error())
		}
	})

	t.Run("should be channel is not a publisherChannel", func(t *testing.T) {
		t.Parallel()

		responseChannel := "responseChannelInvalid"
		container := container.NewGenericContainer[any, any]()
		container.Set("responseChannelInvalid", responseChannel)

		reqMessage := message.NewMessageBuilder().
			WithReplyTo("responseChannelInvalid").
			WithPayload("request").
			Build()

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		got := handler.NewSendReplyToHandler(&replyTohandlerMock{}, container)
		result, err := got.Handle(ctx, reqMessage)
		if err == nil {
			t.Fatal("expected an error, got nil")
		}

		if result != nil {
			t.Fatalf("expected no response message, got %v", result)
		}
		if err.Error() != "[send-reply-to-handler] reply channel is not a publisher channel" {
			t.Fatalf("expected error message '[send-reply-to-handler] reply channel is not a publisher channel', got %v", err.Error())
		}
	})

	t.Run("should be send error processing success", func(t *testing.T) {
		t.Parallel()

		responseChannel := channel.NewPointToPointChannel("responseChannel")
		defer responseChannel.Close()

		container := container.NewGenericContainer[any, any]()
		container.Set("responseChannel", responseChannel)

		reqMessage := message.NewMessageBuilder().
			WithReplyTo("responseChannel").
			WithPayload("error").
			Build()

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		go responseChannel.Receive(ctx)
		got := handler.NewSendReplyToHandler(&replyTohandlerMock{}, container)
		result, err := got.Handle(ctx, reqMessage)
		if err == nil {
			t.Fatal("expected an error, got nil")
		}

		if result != nil {
			t.Fatalf("expected no response message, got %v", result)
		}
	})

	t.Run("should be send when payload is error", func(t *testing.T) {
		t.Parallel()

		responseChannel := channel.NewPointToPointChannel("responseChannel")
		defer responseChannel.Close()

		container := container.NewGenericContainer[any, any]()
		container.Set("responseChannel", responseChannel)

		reqMessage := message.NewMessageBuilder().
			WithReplyTo("responseChannel").
			WithPayload(fmt.Errorf("erro de mensagem")).
			Build()

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		go responseChannel.Receive(ctx)
		got := handler.NewSendReplyToHandler(&replyTohandlerMock{}, container)
		result, err := got.Handle(ctx, reqMessage)
		if err == nil {
			t.Fatalf("expected error, got %v", err)
		}

		if result != nil {
			t.Fatal("expected nil a response message, got %w", result)
		}
	})

}
