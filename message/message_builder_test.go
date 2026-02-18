package message_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/jeffersonbrasilino/gomes/message"
	"github.com/jeffersonbrasilino/gomes/message/channel"
)

func TestNewMessageBuilder(t *testing.T) {
	t.Parallel()
	b := message.NewMessageBuilder()
	if b == nil {
		t.Error("NewMessageBuilder() should not return nil")
	}
}

func TestNewMessageBuilderFromMessage(t *testing.T) {
	t.Parallel()
	b := message.NewMessageBuilder().
		WithPayload("payload").
		WithRoute("route").
		WithMessageType(1).
		WithCorrelationId("cid").
		WithChannelName("ch").
		WithReplyTo("rch").
		WithContext(context.Background())
	msg := b.Build()
	b2 := message.NewMessageBuilderFromMessage(msg)
	if b2 == nil {
		t.Error("NewMessageBuilderFromMessage() should not return nil")
	}
}

func TestNewMessageBuilderFromHeaders(t *testing.T) {
	t.Run("create message builder from valid headers", func(t *testing.T) {
		t.Parallel()
		data := map[string]string{
			message.HeaderOrigin:        "dummy",
			message.HeaderMessageId:     "dummy",
			message.HeaderRoute:         "dummy",
			message.HeaderMessageType:   message.Command.String(),
			message.HeaderReplyTo:       "dummy",
			message.HeaderCorrelationId: "dummy",
			message.HeaderChannelName:   "dummy",
			message.HeaderTimestamp:     time.Now().Format("2006-01-02 15:04:05"),
			message.HeaderVersion:       "dummy",
			"customHeader":              "customValue",
		}
		got, err := message.NewMessageBuilderFromHeaders(data)
		if err != nil {
			t.Errorf("NewMessageBuilderFromHeaders() returned an error: %v", err)
		}

		if got == nil {
			t.Error("NewMessageBuilderFromHeaders() should not return nil")
		}
	})

	t.Run("create message builder from nil headers", func(t *testing.T) {
		t.Parallel()
		got, err := message.NewMessageBuilderFromHeaders(nil)
		if err != nil {
			t.Errorf("NewMessageBuilderFromHeaders() returned an error: %v", err)
		}
		if got == nil {
			t.Error("NewMessageBuilderFromHeaders() should not return nil")
		}
	})

	t.Run("error when timestamp is invalid", func(t *testing.T) {
		t.Parallel()
		data := map[string]string{
			message.HeaderTimestamp: "invalid-timestamp",
		}
		got, err := message.NewMessageBuilderFromHeaders(data)
		fmt.Println("err", err)
		if err == nil {
			t.Error("NewMessageBuilderFromHeaders() should return an error for invalid timestamp")
		}
		if got != nil {
			t.Error("NewMessageBuilderFromHeaders() should return nil for invalid timestamp")
		}
		if err.Error() != "[message-builder] header converter error: timestamp - parsing time \"invalid-timestamp\" as \"2006-01-02 15:04:05\": cannot parse \"invalid-timestamp\" as \"2006\"" {
			t.Errorf("NewMessageBuilderFromHeaders() returned unexpected error: %v", err)
		}
	})

	t.Run("create when data is empty", func(t *testing.T) {
		t.Parallel()
		data := map[string]string{
			message.HeaderOrigin:        "",
			message.HeaderMessageId:     "",
			message.HeaderRoute:         "",
			message.HeaderMessageType:   message.Command.String(),
			message.HeaderReplyTo:       "",
			message.HeaderCorrelationId: "",
			message.HeaderChannelName:   "",
			message.HeaderTimestamp:     time.Now().Format("2006-01-02 15:04:05"),
			message.HeaderVersion:       "",
			"customHeader":              "",
		}
		got, err := message.NewMessageBuilderFromHeaders(data)
		if err != nil {
			t.Errorf("NewMessageBuilderFromHeaders() returned an error: %v", err)
		}

		if got == nil {
			t.Error("NewMessageBuilderFromHeaders() should not return nil")
		}
	})
}

func TestMessageBuilder_chooseMessageType(t *testing.T) {
	cases := []struct {
		description string
		want        message.MessageType
		input       string
	}{
		{
			description: "should return Command for 'Command'",
			want:        message.Command,
			input:       "Command",
		},
		{
			description: "should return Query for 'Query'",
			want:        message.Query,
			input:       "Query",
		},
		{
			description: "should return Document for unknown type",
			want:        message.Document,
			input:       "unknown",
		},
		{
			description: "should return Event for event type",
			want:        message.Event,
			input:       "Event",
		},
	}
	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			t.Parallel()
			b, _ := message.NewMessageBuilderFromHeaders(map[string]string{
				message.HeaderMessageType: tc.input,
			})
			got := b.Build()
			got.GetHeader().Get(message.HeaderMessageType)
			if got.GetHeader().Get(message.HeaderMessageType) != tc.want.String() {
				t.Errorf("ChooseMessageType(%q) = %v; want %v", tc.input, got.GetHeader().Get(message.HeaderMessageType), tc.want.String())
			}
		})
	}
}

func TestWithPayload(t *testing.T) {
	t.Parallel()
	data := "payload"
	b := message.NewMessageBuilder().WithPayload(data).Build()
	if b.GetPayload() != data {
		t.Error("WithPayload did not set payload correctly")
	}
}

func TestWithMessageType(t *testing.T) {
	t.Parallel()
	data := message.Command
	b := message.NewMessageBuilder().WithMessageType(data).Build()
	if b.GetHeader().Get(message.HeaderMessageType) != data.String() {
		t.Error("WithMessageType did not set messageType correctly")
	}
}

func TestWithRoute(t *testing.T) {
	t.Parallel()
	b := message.NewMessageBuilder().WithRoute("route").Build()
	if b.GetHeader().Get(message.HeaderRoute) != "route" {
		t.Error("WithRoute did not set route correctly")
	}
}

func TestWithReplyChannel(t *testing.T) {
	t.Parallel()
	var ch message.PublisherChannel
	b := message.NewMessageBuilder().WithInternalReplyChannel(ch).Build()
	if b.GetInternalReplyChannel() != ch {
		t.Error("WithReplyChannel did not set replyChannel correctly")
	}
}

func TestWithCorrelationId(t *testing.T) {
	t.Parallel()
	b := message.NewMessageBuilder().WithCorrelationId("cid")
	if b.Build().GetHeader().Get(message.HeaderCorrelationId) != "cid" {
		t.Error("WithCorrelationId did not set correlationId correctly")
	}
}

func TestWithChannelName(t *testing.T) {
	t.Parallel()
	b := message.NewMessageBuilder().WithChannelName("ch")
	if b.Build().GetHeader().Get(message.HeaderChannelName) != "ch" {
		t.Error("WithChannelName did not set channelName correctly")
	}
}

func TestWithReplyChannelName(t *testing.T) {
	t.Parallel()
	b := message.NewMessageBuilder().WithReplyTo("rch")
	if b.Build().GetHeader().Get(message.HeaderReplyTo) != "rch" {
		t.Error("WithReplyChannelName did not set replyChannelName correctly")
	}
}

func TestWithContext(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	b := message.NewMessageBuilder().WithContext(ctx).Build()
	if b.GetContext() != ctx {
		t.Error("WithContext did not set context correctly")
	}
}

func TestWithTimestamp(t *testing.T) {
	t.Parallel()
	timestamp := time.Now()
	b := message.NewMessageBuilder().WithTimestamp(timestamp).Build()
	if b.GetHeader().Get(message.HeaderTimestamp) != timestamp.Format("2006-01-02 15:04:05") {
		t.Error("WithTimestamp did not set timestamp correctly")
	}
}

func TestWithOrigin(t *testing.T) {
	t.Parallel()
	b := message.NewMessageBuilder().WithOrigin("origin").Build()
	if b.GetHeader().Get(message.HeaderOrigin) != "origin" {
		t.Error("WithOrigin did not set origin correctly")
	}
}

func TestWithVersion(t *testing.T) {
	t.Parallel()
	b := message.NewMessageBuilder().WithVersion("version").Build()
	if b.GetHeader().Get(message.HeaderVersion) != "version" {
		t.Error("WithVersion did not set version correctly")
	}
}

func TestWithRawMessage(t *testing.T) {
	t.Parallel()
	b := message.NewMessageBuilder().WithRawMessage("rawMessage").Build()
	if b.GetRawMessage() != "rawMessage" {
		t.Error("WithRawMessage did not set rawMessage correctly")
	}
}

func TestBuild(t *testing.T) {
	t.Parallel()
	chn :=channel.NewPointToPointChannel("teste")
	defer chn.Close()
	b := message.NewMessageBuilder().
		WithPayload("payload").
		WithRoute("route").
		WithMessageType(1).
		WithCorrelationId("cid").
		WithChannelName("ch").
		WithReplyTo("rch").
		WithContext(context.Background()).
		WithInternalReplyChannel(chn)
	msg := b.Build()
	if msg == nil {
		t.Error("Build() should not return nil")
	}
}
