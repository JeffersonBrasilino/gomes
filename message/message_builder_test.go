package message_test

import (
	"context"
	"testing"
	"time"

	"github.com/jeffersonbrasilino/gomes/message"
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
	b := message.NewMessageBuilder().
		WithPayload("payload").
		WithRoute("route").
		WithMessageType(1).
		WithCorrelationId("cid").
		WithChannelName("ch").
		WithReplyTo("rch").
		WithContext(context.Background())
	msg := b.Build()
	if msg == nil {
		t.Error("Build() should not return nil")
	}
}
