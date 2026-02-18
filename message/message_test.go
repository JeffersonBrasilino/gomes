package message_test

import (
	"context"
	"testing"
	"time"

	"github.com/jeffersonbrasilino/gomes/message"
	"github.com/jeffersonbrasilino/gomes/message/channel"
)

func TestMessageTypeString(t *testing.T) {
	cases := []struct {
		description string
		should      message.MessageType
		want        string
	}{
		{"should Command", message.Command, "Command"},
		{"should Query", message.Query, "Query"},
		{"should Event", message.Event, "Event"},
		{"should Document", message.Document, "Document"},
	}

	for _, c := range cases {
		t.Run(c.description, func(t *testing.T) {
			t.Parallel()
			if c.should.String() != c.want {
				t.Errorf("%s.String() should return '%s'", c.should, c.want)
			}
		})
	}
}

func TestMessageHeaders(t *testing.T) {
	t.Run("should create success", func(t *testing.T) {
		t.Parallel()
		timestamp := time.Now().Format("2006-01-02 15:04:05")
		data := map[string]string{
			message.HeaderOrigin:        "dummy",
			message.HeaderMessageId:     "dummy",
			message.HeaderRoute:         "dummy",
			message.HeaderMessageType:   message.Command.String(),
			message.HeaderReplyTo:       "dummy",
			message.HeaderCorrelationId: "dummy",
			message.HeaderChannelName:   "dummy",
			message.HeaderTimestamp:     timestamp,
			message.HeaderVersion:       "dummy",
		}
		header := message.NewHeader(data)
		for key, val := range data {
			if header.Get(key) != val {
				t.Errorf("Expected %s '%s', got '%s'", key, val, header.Get(key))
			}
		}
	})

	t.Run("return get data", func(t *testing.T) {
		t.Parallel()
		header := message.NewHeader(nil)
		if header.Get(message.HeaderReplyTo) != "" {
			t.Errorf("Expected empty data got '%s'", header.Get(message.HeaderReplyTo))
		}
	})

	t.Run("return set data success", func(t *testing.T) {
		t.Parallel()
		header := message.NewHeader(nil)
		header.Set("newHeader", "okok")
		if header.Get("newHeader") != "okok" {
			t.Errorf("Expected empty data got '%s'", header.Get(message.HeaderReplyTo))
		}
	})

	t.Run("return set data error", func(t *testing.T) {
		t.Parallel()
		header := message.NewHeader(nil)
		err := header.Set(message.HeaderMessageId, "okok")
		if err == nil {
			t.Errorf("Expected error")
		}
	})
}

func TestNewMessage(t *testing.T) {
	t.Parallel()
	headers := message.NewHeader(nil)
	ctx := context.Background()
	msg := message.NewMessage(ctx, "payload", headers)
	if msg.GetPayload() != "payload" {
		t.Error("GetPayload did not return correct value")
	}
	if msg.GetHeader() == nil {
		t.Error("GetHeader did not return correct value")
	}
	if msg.GetContext() != ctx {
		t.Error("GetContext did not return correct value")
	}
}

func TestMessage_SetContext(t *testing.T) {
	t.Parallel()
	headers := message.NewHeader(nil)
	msg := message.NewMessage(nil, "payload", headers)
	ctx := context.Background()
	msg.SetContext(ctx)
	if msg.GetContext() != ctx {
		t.Error("SetContext did not set context correctly")
	}
}

func TestMessage_Getters(t *testing.T) {
	headers := message.NewHeader(nil)
	ctx := context.Background()
	msg := message.NewMessage(ctx, "payload", headers)
	cases := []struct {
		description string
		should      func() any
		want        any
	}{
		{"GetPayload should return correct value", func() any {
			return msg.GetPayload()
		}, "payload"},
		{"GetContext should return correct value", func() any {
			return msg.GetContext()
		}, ctx},
	}

	for _, c := range cases {
		t.Run(c.description, func(t *testing.T) {
			t.Parallel()
			if c.should() != c.want {
				t.Errorf("%s() should return '%v'", c.description, c.want)
			}
		})
	}
}

func TestMessage_ReplyRequired(t *testing.T) {
	cases := []struct {
		description string
		should      message.MessageType
		want        bool
	}{
		{"ReplyRequired should return true for Command", message.Command, true},
		{"ReplyRequired should return true for Query", message.Query, true},
		{"ReplyRequired should return false for Event", message.Event, false},
		{"ReplyRequired should return false for Document", message.Document, false},
	}
	for _, c := range cases {
		t.Run(c.description, func(t *testing.T) {
			t.Parallel()
			headers := message.NewHeader(map[string]string{
				message.HeaderMessageType: c.should.String(),
			})
			msg := message.NewMessage(nil, "payload", headers)
			if msg.ReplyRequired() != c.want {
				t.Errorf("%q: got %v, want %v", c.description, msg.ReplyRequired(), c.want)
			}
		})
	}
}

func TestMessage_InternalReplyChannel(t *testing.T) {
	msg := message.NewMessage(context.TODO(), "payload", nil)
	msg.SetInternalReplyChannel(channel.NewPointToPointChannel("tst"))

	if msg.GetInternalReplyChannel() == nil {
		t.Error("Expected internal reply channel to be set, got nil")
	}
}
