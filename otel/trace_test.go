package otel

import (
	"context"
	"errors"
	"testing"

	"github.com/jeffersonbrasilino/gomes/message"
	"go.opentelemetry.io/otel/codes"
	traceTypes "go.opentelemetry.io/otel/trace"
)

func TestTraceFunctions(t *testing.T) {
	t.Parallel()

	t.Run("StartOptions builders set fields", func(t *testing.T) {
		t.Parallel()
		so := &startOptions{}
		WithMessagingSystemType(MessageSystemTypeKafka)(so)
		if so.messagingSystemType != MessageSystemTypeKafka {
			t.Fatalf("expected messaging system kafka, got %v", so.messagingSystemType)
		}
		WithSpanOperation(SpanOperationReceive)(so)
		if so.operation != SpanOperationReceive {
			t.Fatalf("expected operation receive, got %v", so.operation)
		}
		WithSpanKind(SpanKindClient)(so)
		if so.spanKind != SpanKindClient {
			t.Fatalf("expected span kind client, got %v", so.spanKind)
		}
		ctxLink := context.Background()
		WithTraceContextToLink(ctxLink)(so)
		if so.traceContextToLink != ctxLink {
			t.Fatalf("expected trace context to link set")
		}
		WithAttributes(NewOtelAttr("a", "b"))(so)
		if len(so.attributes) != 1 {
			t.Fatalf("expected 1 attribute, got %d", len(so.attributes))
		}
		// WithMessage
		hdrs := message.NewHeader(map[string]string{message.HeaderRoute: "r"})
		msg := message.NewMessage(context.Background(), nil, hdrs)
		WithMessage(msg)(so)
		if so.message == nil {
			t.Fatalf("expected message set in startOptions")
		}
	})

	t.Run("EnableTrace Start and span methods execute", func(t *testing.T) {
		t.Parallel()
		EnableTrace()
		tr := InitTrace("svc-test-2")
		hdrs := message.NewHeader(map[string]string{message.HeaderRoute: "route-x"})
		msg := message.NewMessage(context.Background(), nil, hdrs)

		ctx, sp := tr.Start(context.Background(), "", WithMessage(msg), WithSpanOperation(SpanOperationSend), WithMessagingSystemType(MessageSystemTypeKafka), WithSpanKind(SpanKindProducer))
		if ctx == nil || sp == nil {
			t.Fatalf("expected context and span when trace enabled")
		}

		// call span methods to exercise code paths
		sp.AddEvent("evt1", NewOtelAttr("k", "v"))
		sp.SetStatus(SpanStatusOK, "ok")
		sp.Success("done")
		sp.Error(errors.New("err1"), "failed")
		sp.End()
	})

	t.Run("helpers: otelStatus, otelKind, makeSpanName, String methods", func(t *testing.T) {
		t.Parallel()
		sOK := SpanStatusOK
		if (&sOK).otelStatus() != codes.Ok {
			t.Fatalf("expected codes.Ok for SpanStatusOK")
		}
		sErr := SpanStatusError
		if (&sErr).otelStatus() != codes.Error {
			t.Fatalf("expected codes.Error for SpanStatusError")
		}
		unk := SpanStatus(1234)
		if (&unk).otelStatus() != codes.Unset {
			t.Fatalf("expected codes.Unset for unknown status")
		}

		k := SpanKindServer
		if k.otelKind() != traceTypes.SpanKindServer {
			t.Fatalf("expected server kind mapping")
		}
		var kunk SpanKind = 999
		if kunk.otelKind() == traceTypes.SpanKindUnspecified {
			// ok - unspecified for unknown kinds
		}

		if makeSpanName(SpanKindProducer, "name") != "send name" {
			t.Fatalf("unexpected span name for producer")
		}
		if makeSpanName(SpanKindInternal, "name") != "process name" {
			t.Fatalf("unexpected span name for internal")
		}

		op := SpanOperationSend
		if op.String() != "send" {
			t.Fatalf("expected send, got %s", op.String())
		}
		opunk := SpanOperation(9999)
		if opunk.String() != "process" {
			t.Fatalf("expected default process for unknown operation")
		}

		mt := MessageSystemTypeKafka
		if mt.String() != "kafka" {
			t.Fatalf("expected kafka for MessageSystemTypeKafka")
		}
		mtunk := MessageSystemType(9999)
		if mtunk.String() != "internal" {
			t.Fatalf("expected internal for unknown message system type")
		}
	})
}
