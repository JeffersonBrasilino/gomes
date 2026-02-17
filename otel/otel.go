// Package otel provides helpers to instrument application code using
// OpenTelemetry. Intent: offer a small, opinionated helper layer to create
// tracers, spans and events consistently across the codebase. Objective: make
// it simple to start spans, add events and propagate trace context for
// messages and HTTP flows.
package otel

import (
	"context"

	"github.com/jeffersonbrasilino/gomes/message"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	traceTypes "go.opentelemetry.io/otel/trace"
)

const GomesOtelTraceEnableFlagName = "gomes.otel.Enable"

// OtelTrace is an interface to start spans and produce OtelSpan instances.
type OtelTrace interface {
	// Start initiates a new span for the given context and name.
	// Parameters:
	//   ctx: propagation context used by the span.
	//   name: span name. If empty, implementation may derive a name.
	//   options: functional options that modify span creation.
	// Returns:
	//   context.Context: context that contains the started span.
	//   OtelSpan: wrapper for the created span.
	Start(ctx context.Context, name string, options ...StartOptions) (context.Context, OtelSpan)
}

// OtelSpan is a minimal wrapper around an OpenTelemetry span that provides
// utility methods to end the span, add events and record status.
type OtelSpan interface {
	// End finalizes the span and records its end time.
	End()
	// AddEvent records an event on the span.
	// Parameters:
	//   eventMessage: event name or description.
	//   attributes: optional list of OtelAttribute to attach to the event.
	AddEvent(eventMessage string, attributes ...OtelAttribute)
	// SetStatus sets the span status and an optional description.
	// Parameters:
	//   status: SpanStatus value to set.
	//   description: human-readable explanation for the status.
	SetStatus(status SpanStatus, description string)
	// Success marks the span as successful with the provided message.
	Success(message string)
	// Error marks the span as errored and records the provided error.
	// Parameters:
	//   err: error instance to record.
	//   message: descriptive error message.
	Error(err error, message string)
}

// OtelAttribute represents a key/value pair used as a span or event attribute.
type OtelAttribute struct {
	key   string
	value string
}

// makeAttributes converts a list of OtelAttribute into an OpenTelemetry
// span/event option that attaches those attributes.
// Parameters:
//
//	attributes: list of OtelAttribute to convert.
//
// Returns:
//
//	traceTypes.SpanStartEventOption: an option that can be passed to span start
//	or event creation functions.
func makeAttributes(attributes []OtelAttribute) traceTypes.SpanStartEventOption {
	var attrs []attribute.KeyValue
	if len(attributes) > 0 {
		for _, attr := range attributes {
			attrs = append(attrs, attribute.String(attr.key, attr.value))
		}
	}
	return traceTypes.WithAttributes(attrs...)
}

// NewOtelAttr creates an OtelAttribute from a key and value.
// Parameters:
//
//	key: attribute key.
//	value: attribute value.
//
// Returns:
//
//	OtelAttribute: constructed attribute instance.
func NewOtelAttr(key string, value string) OtelAttribute {
	return OtelAttribute{
		key:   key,
		value: value,
	}
}

// makeAttributesFromMessage builds a slice of OtelAttribute extracted from a
// message.Message headers. It maps common messaging header fields to
// semantic attribute keys used for tracing.
func makeAttributesFromMessage(msg *message.Message) []OtelAttribute {
	messageHeaders := msg.GetHeader()
	destinationName := messageHeaders.Get(message.HeaderRoute)
	if messageHeaders.Get(message.HeaderChannelName) != "" {
		destinationName = messageHeaders.Get(message.HeaderChannelName)
	}
	return []OtelAttribute{
		NewOtelAttr("messaging.message.id", messageHeaders.Get(message.HeaderMessageId)),
		NewOtelAttr("messaging.message.correlationId", messageHeaders.Get(message.HeaderCorrelationId)),
		NewOtelAttr("command.name", messageHeaders.Get(message.HeaderRoute)),
		NewOtelAttr("messaging.type", messageHeaders.Get(message.HeaderMessageType)),
		NewOtelAttr("command.version", messageHeaders.Get(message.HeaderVersion)),
		NewOtelAttr("messaging.destination.name", destinationName),
	}
}

// GetTraceContextPropagatorByContext extracts the current trace context from
// ctx using the global text map propagator and returns it as a map of header
// keys to values. Useful for attaching trace headers to outgoing messages.
func GetTraceContextPropagatorByContext(ctx context.Context) map[string]string {
	carrier := propagation.HeaderCarrier{}
	propagator := otel.GetTextMapPropagator()
	propagator.Inject(ctx, &carrier)

	var result = make(map[string]string)
	for _, key := range carrier.Keys() {
		result[key] = carrier.Get(key)
	}

	return result
}

// GetTraceContextPropagatorByTraceParent extracts trace context from a trace parent header.
// It creates a new context with the extracted trace information.
//
// Parameters:
//   - ctx: the base context
//   - traceParent: the W3C trace parent header string
//
// Returns:
//   - context.Context: context with extracted trace information
func GetTraceContextPropagatorByTraceParent(
	ctx context.Context,
	traceParent string,
) context.Context {
	carrier := propagation.HeaderCarrier{}
	carrier.Set("Traceparent", traceParent)
	propagator := otel.GetTextMapPropagator()
	return propagator.Extract(ctx, &carrier)
}
