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

// mockAction implements handler.Action for tests.
type mockAction struct {
	name string
}

func (a mockAction) Name() string {
	return a.name
}

// mockActionHandler implements handler.ActionHandler for tests.
type mockActionHandler struct {
	result string
}

func (h *mockActionHandler) Handle(ctx context.Context, action *mockAction) (any, error) {
	if h.result == "failure" {
		return nil, fmt.Errorf("handler error")
	}
	return h.result, nil
}

func TestNewActionHandleActivatorBuilder(t *testing.T) {
	t.Parallel()
	action := &mockActionHandler{result: "ok"}
	builder := handler.NewActionHandleActivatorBuilder("ref", action)
	if builder.ReferenceName() != "ref" {
		t.Errorf("Expected ReferenceName 'ref', got '%s'", builder.ReferenceName())
	}
}

func TestActionHandleActivatorBuilder_Build(t *testing.T) {
	t.Parallel()
	action := &mockActionHandler{result: "ok"}
	builder := handler.NewActionHandleActivatorBuilder("ref", action)
	cont := container.NewGenericContainer[any, any]()
	chn, err := builder.Build(cont)
	if err != nil {
		t.Errorf("Expected success, got error: %v", err)
	}
	if chn == nil {
		t.Error("Expected channel instance, got nil")
	}
}

func TestActionHandleActivator_Handle(t *testing.T) {
	cases := []struct {
		description         string
		expectError         bool
		expectChanMessage   bool
		expectResultMessage bool
	}{
		{"success", false, true, true},
		{"failure", true, false, false},
		{"invalid payload", true, false, false},
	}
	resChn := make(chan *message.Message, 50)
	for _, c := range cases {
		t.Run(c.description, func(t *testing.T) {
			t.Parallel()
			r := "ok"
			if c.expectError {
				r = "failure"
			}
			actionHandler := &mockActionHandler{result: r}
			activator := handler.NewActionHandlerActivator(actionHandler)
			replyChan := channel.NewPointToPointChannel(fmt.Sprintf("reply-%v", c.description))
			go func(chn chan<- *message.Message, replyChn *channel.PointToPointChannel) {
				r, _ := replyChn.Receive(context.TODO())
				chn <- r
			}(resChn, replyChan)
			msg := message.NewMessageBuilder().
				WithChannelName("channel").
				WithMessageType(message.Command).
				WithPayload(&mockAction{name: "test"}).
				WithReplyChannel(replyChan)

			if c.description == "invalid payload" {
				msg.WithPayload("teste")
			}
			ctx := context.Background()
			result, err := activator.Handle(ctx, msg.Build())
			if c.expectError && err == nil {
				t.Error("Expected error, got nil")
			}
			if c.expectChanMessage && <-resChn == nil {
				t.Error("Expected channel message, got nil")
			}
			if c.expectResultMessage && result == nil {
				t.Error("Expected result message, got nil")
			}
		})
	}
}
