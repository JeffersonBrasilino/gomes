package endpoint_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/jeffersonbrasilino/gomes/container"
	"github.com/jeffersonbrasilino/gomes/message"
	"github.com/jeffersonbrasilino/gomes/message/channel"
	"github.com/jeffersonbrasilino/gomes/message/endpoint"
)

// fakeInboundAdapter is a lightweight test double for InboundChannelAdapter.
type fakeInboundAdapter struct {
	ch *channel.PointToPointChannel
}

func (f *fakeInboundAdapter) ReferenceName() string {
	if f.ch == nil {
		return "nil-channel"
	}

	return f.ch.Name()
}
func (f *fakeInboundAdapter) DeadLetterChannelName() string {
	return ""
}
func (f *fakeInboundAdapter) AfterProcessors() []message.MessageHandler {
	return []message.MessageHandler{&dummyEventDrivenGatewayHandler{nil}}
}
func (f *fakeInboundAdapter) BeforeProcessors() []message.MessageHandler {
	return []message.MessageHandler{&dummyEventDrivenGatewayHandler{nil}}
}
func (f *fakeInboundAdapter) RetryAttempts() []int {
	return []int{0}
}
func (f *fakeInboundAdapter) Close() error {
	if f.ch != nil {
		f.ch.Close()
	}
	return nil
}
func (f *fakeInboundAdapter) ReceiveMessage(ctx context.Context) (*message.Message, error) {

	if f.ch == nil {
		return nil, nil
	}

	msg, err := f.ch.Receive(ctx)
	if err != nil {
		return nil, nil
	}

	if msg.GetPayload() == "error" {
		return nil, fmt.Errorf("error receiving message")
	}
	return msg, nil
}

type dummyEventDrivenGatewayHandler struct {
	response chan any
}

func (d *dummyEventDrivenGatewayHandler) Handle(
	_ context.Context,
	msg *message.Message,
) (*message.Message, error) {
	if d.response == nil {
		return nil, nil
	}
	time.Sleep(time.Second * 1)
	if msg.GetPayload() == "payload error" {
		d.response <- fmt.Errorf("payload error")
		return nil, fmt.Errorf("payload error")
	}
	d.response <- msg
	return msg, nil
}

func TestNewEventDrivenConsumerBuilder_Build(t *testing.T) {
	t.Run("builds EventDrivenConsumer successfully", func(t *testing.T) {
		t.Parallel()
		cont := container.NewGenericContainer[any, any]()
		in := &fakeInboundAdapter{nil}
		cont.Set("ref", in)
		got, err := endpoint.NewEventDrivenConsumerBuilder("ref").
			Build(cont)

		if err != nil {
			t.Errorf("Expected success, got error: %v", err)
		}
		if got == nil {
			t.Error("Expected EventDrivenConsumer instance, got nil")
		}
	})
	t.Run("Fails to build when gateway not found", func(t *testing.T) {
		t.Parallel()
		cont := container.NewGenericContainer[any, any]()
		got, err := endpoint.NewEventDrivenConsumerBuilder("ref").
			Build(cont)
		if got != nil {
			t.Errorf("Expected nil, got: %v", got)
		}
		if err == nil || err.Error() != "[event-driven-consumer] consumer channel ref not found." {
			t.Errorf("Expected error '[event-driven-consumer] consumer channel ref not found.', got: %v", err)
		}
	})
	t.Run("Fails to build when channel adapter is invalid", func(t *testing.T) {
		t.Parallel()
		cont := container.NewGenericContainer[any, any]()
		in := "invalid adapter"
		cont.Set("ref", in)
		got, err := endpoint.NewEventDrivenConsumerBuilder("ref").
			Build(cont)
		fmt.Println(got, err)
		if got != nil {
			t.Errorf("Expected nil, got: %v", got)
		}
		if err == nil || err.Error() != "[event-driven-consumer] consumer channel ref is not a consumer channel." {
			t.Errorf("Expected error '[event-driven-consumer] consumer channel ref is not a consumer channel.', got: %v", err)
		}
	})
}

func TestEventDrivenConsumer_Run(t *testing.T) {
	t.Run("processes message successfully", func(t *testing.T) {
		t.Parallel()

		inChannel := channel.NewPointToPointChannel("in")
		outChannel := make(chan any)
		in := &fakeInboundAdapter{ch: inChannel}

		gw := endpoint.NewGateway(&dummyEventDrivenGatewayHandler{response: outChannel}, "", "")
		consumer := endpoint.NewEventDrivenConsumer("ref", gw, in)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		go consumer.Run(ctx)

		msg := message.NewMessageBuilder().
			WithChannelName("in").
			WithMessageType(message.Command).
			WithPayload("payload").
			Build()
		inChannel.Send(ctx, msg)

		select {
		case res := <-outChannel:
			resMsg, ok := res.(*message.Message)
			if !ok {
				t.Errorf("expected a message response, got: %v", res)
			} else if resMsg.GetPayload() != "payload" {
				t.Errorf("expected 'payload', got: %v", resMsg.GetPayload())
			}
		case <-time.After(2 * time.Second):
			t.Fatal("timeout waiting for handler response")
		}

		t.Cleanup(func() {
			consumer.Stop()
			close(outChannel)
		})
	})
	t.Run("Receive message Error", func(t *testing.T) {

		inChannel := channel.NewPointToPointChannel("in")
		outChannel := make(chan any)
		in := &fakeInboundAdapter{ch: inChannel}

		gw := endpoint.NewGateway(&dummyEventDrivenGatewayHandler{response: outChannel}, "", "")
		consumer := endpoint.NewEventDrivenConsumer("ref", gw, in)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		errChan := make(chan error, 1)
		go func() {
			errChan <- consumer.Run(ctx)
		}()

		msg := message.NewMessageBuilder().
			WithChannelName("in").
			WithMessageType(message.Command).
			WithPayload("error").
			Build()
		inChannel.Send(ctx, msg)

		got := <-errChan

		if got == nil || got.Error() != "error receiving message" {
			t.Errorf("expected error 'error receiving message', got: %v", got)
		}

		t.Cleanup(func() {
			consumer.Stop()
			close(outChannel)
		})
	})
}

func TestEventDrivenConsumer_ConfigFunctions(t *testing.T) {
	configFunctions := []struct {
		name           string
		functionConfig func(*endpoint.EventDrivenConsumer) *endpoint.EventDrivenConsumer
	}{
		{
			"WithMessageProcessingTimeout",
			func(c *endpoint.EventDrivenConsumer) *endpoint.EventDrivenConsumer {
				return c.WithMessageProcessingTimeout(5)
			},
		},
		{
			"WithAmountOfProcessors",
			func(c *endpoint.EventDrivenConsumer) *endpoint.EventDrivenConsumer {
				return c.WithAmountOfProcessors(5)
			},
		},
		{
			"WithStopOnError",
			func(c *endpoint.EventDrivenConsumer) *endpoint.EventDrivenConsumer {
				return c.WithStopOnError(true)
			},
		},
	}

	for _, cf := range configFunctions {
		t.Run(cf.name, func(t *testing.T) {
			t.Parallel()
			got := cf.functionConfig(endpoint.NewEventDrivenConsumer("", nil, nil))
			if got == nil {
				t.Errorf("expected type *EventDrivenConsumer, got nil")
			}
		})
	}
}
