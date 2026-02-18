package otel

import (
    "context"
    "testing"

    "github.com/jeffersonbrasilino/gomes/message"
)

func TestOtelHelpers(t *testing.T) {
    t.Parallel()

    t.Run("makeAttributes empty and non-empty", func(t *testing.T) {
        t.Parallel()
        // empty attributes
        _ = makeAttributes([]OtelAttribute{})

        // non-empty attributes via NewOtelAttr
        attr := NewOtelAttr("k", "v")
        opt := makeAttributes([]OtelAttribute{attr})
        _ = opt
    })

    t.Run("NewOtelAttr fields", func(t *testing.T) {
        t.Parallel()
        a := NewOtelAttr("kk", "vv")
        if a.key != "kk" {
            t.Fatalf("expected key kk, got %s", a.key)
        }
        if a.value != "vv" {
            t.Fatalf("expected value vv, got %s", a.value)
        }
    })

    t.Run("makeAttributesFromMessage uses headers and channel name", func(t *testing.T) {
        t.Parallel()
        hdrs := map[string]string{
            message.HeaderRoute:       "route1",
            message.HeaderChannelName: "chan1",
            message.HeaderCorrelationId: "corr-1",
            message.HeaderMessageType: "mt",
            message.HeaderVersion: "v1",
        }
        h := message.NewHeader(hdrs)
        msg := message.NewMessage(context.Background(), nil, h)

        attrs := makeAttributesFromMessage(msg)
        if len(attrs) != 6 {
            t.Fatalf("expected 6 attributes, got %d", len(attrs))
        }
        // destination name should come from channelName
        found := false
        for _, a := range attrs {
            if a.key == "messaging.destination.name" && a.value == "chan1" {
                found = true
            }
        }
        if !found {
            t.Fatalf("expected messaging.destination.name=chan1 among attributes")
        }
    })

    t.Run("GetTraceContextPropagatorByContext and ByTraceParent", func(t *testing.T) {
        t.Parallel()
        // With a background context the carrier is usually empty
        m := GetTraceContextPropagatorByContext(context.Background())
        if m == nil {
            t.Fatalf("expected map, got nil")
        }

        // Provide a valid traceparent header to ensure Extract path runs
        tp := "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01"
        ctx := GetTraceContextPropagatorByTraceParent(context.Background(), tp)
        if ctx == nil {
            t.Fatalf("expected non-nil context from GetTraceContextPropagatorByTraceParent")
        }
    })
}
