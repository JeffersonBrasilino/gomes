// Package otel provides utilities for OpenTelemetry integration, facilitating instrumentation,
// creation and management of spans, events and status for distributed tracing in Go applications.

// Example usage:
//
//	exporter, err := otel.InitOtelTraceProvider("myService")
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

// OtelTrace defines interface for creating custom spans.
type OtelTrace interface {

	// Start initiates a new trace.
	// Parameters:
	//   ctx context.Context - propagation context.
	//   name string - service trace name.
	//   attributes ...otelAttribute - optional attributes.
	// Returns:
	//   context.Context - updated context.
	//   OtelSpan - created trace span.
	Start(
		ctx context.Context,
		name string,
		options ...StartOptions,
	) (context.Context, OtelSpan)
}

// OtelSpan defines interface for span manipulation.
type OtelSpan interface {
	// End finalizes the span.
	//
	// Example usage:
	//   span.End()
	End()

	// AddEvent adds an event to the span, with optional attributes.
	//
	// Parameters:
	//   eventMessage string - event message.
	//   attributes ...otelAttribute - optional attributes.
	//
	// Example usage:
	//   span.AddEvent("event", attr1)
	AddEvent(eventMessage string, attributes ...OtelAttribute)

	// SetStatus sets the span status (success or error) and a description.
	//
	// Parameters:
	//   status SpanStatus - span status.
	//   description string - status description.
	//
	// Example usage:
	//   span.SetStatus(SpanStatusOK, "ok")
	SetStatus(status SpanStatus, description string)

	// Success marks the span as successful, with descriptive message.
	//
	// Parameters:
	//   message string - success message.
	//
	// Example usage:
	//   span.Success("operation completed")
	Success(message string)

	// Error marks the span as error, with descriptive message.
	//
	// Parameters:
	//   message string - descriptive error message.
	//   err error - error to record.
	//
	// Example usage:
	//   span.Error(err, "operation failed")
	Error(err error, message string)
}

// otelAttribute representa um par chave-valor para atributos de span/evento.
type OtelAttribute struct {
	key   string
	value string
}

// makeAttributes converts a list of otelAttribute into attributes for the span/event.
//
// Parameters:
//
//	attributes []otelAttribute - list of attributes.
//
// Returns:
//
//	traceTypes.SpanStartEventOption - option with attributes for span/event creation.
//
// Example usage:
//
//	opts := makeAttributes([]otelAttribute{core.NewOtelAttr("key", "value")})
func makeAttributes(attributes []OtelAttribute) traceTypes.SpanStartEventOption {
	var attrs []attribute.KeyValue
	if len(attributes) > 0 {
		for _, attr := range attributes {
			attrs = append(attrs, attribute.String(attr.key, attr.value))
		}
	}
	return traceTypes.WithAttributes(attrs...)
}

// NewOtelAttr creates a new otelAttribute with the provided key and value.
//
// Parameters:
//
//	key string - attribute key.
//	value string - attribute value.
//
// Returns:
//
//	otelAttribute - created attribute.
//
// Example usage:
//
//	attr := otel.NewOtelAttr("key", "value")
func NewOtelAttr(key string, value string) OtelAttribute {
	return OtelAttribute{
		key:   key,
		value: value,
	}
}

func makeAttributesFromMessage(message *message.Message) []OtelAttribute {
	messageHeaders := message.GetHeaders()
	destinationName := messageHeaders.Route
	if messageHeaders.ChannelName != "" {
		destinationName = messageHeaders.ChannelName
	}
	return []OtelAttribute{
		NewOtelAttr("messaging.message.id", messageHeaders.MessageId),
		NewOtelAttr("messaging.message.correlationId", messageHeaders.CorrelationId),
		NewOtelAttr("command.name", messageHeaders.Route),
		NewOtelAttr("messaging.type", messageHeaders.MessageType.String()),
		NewOtelAttr("command.version", messageHeaders.Version),
		NewOtelAttr("messaging.destination.name", destinationName),
	}
}

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
