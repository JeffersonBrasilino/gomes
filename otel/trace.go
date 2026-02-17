// Package otel provides an implementation for OpenTelemetry tracing
// functionality. Intent: present a lightweight API to create and manage
// traces and spans across message and HTTP flows. Objective: simplify span
// creation, status reporting and propagation for the application.
package otel

import (
	"context"
	"fmt"

	"github.com/jeffersonbrasilino/gomes/message"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	traceTypes "go.opentelemetry.io/otel/trace"
)

// SpanStatus represents the status of a span, indicating whether the operation
// was successful or not.
type (
	SpanStatus        int
	SpanKind          int
	SpanOperation     int
	MessageSystemType int
	StartOptions      func(*startOptions)
)

// Constants for span status values
const (
	SpanStatusOK    SpanStatus = iota // Operation completed successfully
	SpanStatusError                   // Operation failed with an error

	SpanKindInternal SpanKind = iota // Internal operation span
	SpanKindServer                   // Server-side span
	SpanKindClient                   // Client-side span
	SpanKindProducer                 // Producer span
	SpanKindConsumer                 // Consumer span

	SpanOperationSend SpanOperation = iota
	SpanOperationReceive
	SpanOperationProcess
	SpanOperationCreate
	SpanOperationSettle

	MessageSystemTypeInternal MessageSystemType = iota
	MessageSystemTypeActiveMQ
	MessageSystemTypeSNS
	MessageSystemTypeSQS
	MessageSystemTypeEventGrid
	MessageSystemTypeEventHubs
	MessageSystemTypeGCPPubSub
	MessageSystemTypeJMS
	MessageSystemTypeKafka
	MessageSystemTypePulsar
	MessageSystemTypeRabbitMQ
	MessageSystemTypeRocketMQ
	MessageSystemTypeServiceBus
)

var (
	traceEnabled bool = false
	msMap             = map[MessageSystemType]string{
		MessageSystemTypeActiveMQ:   "activemq",
		MessageSystemTypeSNS:        "aws.sns",
		MessageSystemTypeSQS:        "aws.sqs",
		MessageSystemTypeEventGrid:  "eventgrid",
		MessageSystemTypeEventHubs:  "eventhubs",
		MessageSystemTypeGCPPubSub:  "gcp_pubsub",
		MessageSystemTypeJMS:        "jms",
		MessageSystemTypeKafka:      "kafka",
		MessageSystemTypePulsar:     "pulsar",
		MessageSystemTypeRabbitMQ:   "rabbitmq",
		MessageSystemTypeRocketMQ:   "rocketmq",
		MessageSystemTypeServiceBus: "servicebus",
	}

	operationMap = map[SpanOperation]string{
		SpanOperationSend:    "send",
		SpanOperationReceive: "receive",
		SpanOperationCreate:  "create",
		SpanOperationSettle:  "settle",
	}

	kindsMap = map[SpanKind]traceTypes.SpanKind{
		SpanKindInternal: traceTypes.SpanKindInternal,
		SpanKindServer:   traceTypes.SpanKindServer,
		SpanKindClient:   traceTypes.SpanKindClient,
		SpanKindProducer: traceTypes.SpanKindProducer,
		SpanKindConsumer: traceTypes.SpanKindConsumer,
	}
)

// EnableTrace enables tracing for the message system.
func EnableTrace() {
	traceEnabled = true
}

// otelTrace implements the OtelTrace interface for creating and managing traces
type otelTrace struct {
	tracer traceTypes.Tracer
}

// otelSpan implements the OtelSpan interface for span operations
type otelSpan struct {
	span traceTypes.Span
	ctx  context.Context
}

type startOptions struct {
	messagingSystemType MessageSystemType
	operation           SpanOperation
	spanKind            SpanKind
	traceContextToLink  context.Context
	attributes          []OtelAttribute
	message             *message.Message
}

// InitTrace creates a new trace instance for the given service.
//
// Parameters:
//   - serviceName string - name of the service for tracing identification
//
// Returns:
//   - *otelTrace - configured trace instance
//
// Example usage:
//
//	tracer := otel.InitTrace("user-service")
func InitTrace(serviceName string) *otelTrace {
	tracer := otel.Tracer(serviceName)
	return &otelTrace{
		tracer: tracer,
	}
}

// WithMessagingSystemType returns a StartOptions that sets the messaging
// system type on the span start options.
func WithMessagingSystemType(mt MessageSystemType) StartOptions {
	return func(so *startOptions) {
		so.messagingSystemType = mt
	}
}

// WithSpanOperation returns a StartOptions that sets the span operation
// type. This option is intended for messaging spans and typically does not
// apply to HTTP spans.
func WithSpanOperation(operation SpanOperation) StartOptions {
	return func(so *startOptions) {
		so.operation = operation
	}
}

// WithSpanKind returns a StartOptions that sets the span kind (server,
// client, producer, consumer, internal).
func WithSpanKind(kind SpanKind) StartOptions {
	return func(so *startOptions) {
		so.spanKind = kind
	}
}

// WithTraceContextToLink returns a StartOptions that links an existing trace
// context to the newly created span, creating an explicit link between
// spans.
func WithTraceContextToLink(ctx context.Context) StartOptions {
	return func(so *startOptions) {
		so.traceContextToLink = ctx
	}
}

// WithAttributes returns a StartOptions that appends OtelAttribute values to
// the span start options.
func WithAttributes(attributes ...OtelAttribute) StartOptions {
	return func(so *startOptions) {
		so.attributes = attributes
	}
}

// WithMessage returns a StartOptions that sets the message associated with
// the span. When provided, the message headers are used to populate common
// messaging attributes.
func WithMessage(message *message.Message) StartOptions {
	return func(so *startOptions) {
		so.message = message
	}
}

// Start initiates a new trace span with the given name and attributes.
//
// Parameters:
//   - ctx context.Context - context for trace propagation
//   - name string - name of the span
//   - attributes ...OtelAttribute - optional span attributes
//
// Returns:
//   - context.Context - updated context with span
//   - OtelSpan - new span instance
//
// Example usage:
//
//	ctx, span := tracer.Start(ctx, "process-user", otel.NewOtelAttr("user_id", "123"))
//	defer span.End()
func (t *otelTrace) Start(
	ctx context.Context,
	name string,
	options ...StartOptions,
) (context.Context, OtelSpan) {
	spanName := name

	if traceEnabled == false {
		return ctx, &otelSpan{}
	}

	startOptions := &startOptions{
		messagingSystemType: MessageSystemTypeInternal,
		spanKind:            SpanKindInternal,
	}

	for _, opt := range options {
		opt(startOptions)
	}

	attributes := startOptions.attributes
	if startOptions.message != nil {
		attributes = append(
			attributes,
			makeAttributesFromMessage(startOptions.message)...,
		)
		if name == "" {
			spanName = makeSpanName(
				startOptions.spanKind,
				startOptions.message.GetHeader().Get(message.HeaderRoute),
			)
		}
	}

	attributes = append(attributes,
		NewOtelAttr("messaging.system", startOptions.messagingSystemType.String()),
	)

	if startOptions.operation != 0 {
		attributes = append(attributes,
			NewOtelAttr("messaging.operation.type", startOptions.operation.String()),
		)
	}

	traceAttribute := []trace.SpanStartOption{
		trace.WithSpanKind(startOptions.spanKind.otelKind()),
		makeAttributes(attributes),
	}

	if startOptions.traceContextToLink != nil {
		spanContext := trace.SpanContextFromContext(startOptions.traceContextToLink)
		if spanContext.IsValid() {
			traceAttribute = append(traceAttribute,
				trace.WithLinks(trace.Link{
					SpanContext: spanContext,
				}),
			)
		}
	}

	ctx, span := t.tracer.Start(
		ctx,
		spanName,
		traceAttribute...,
	)

	return ctx, &otelSpan{
		span: span,
		ctx:  ctx,
	}
}

// End finalizes the span, marking its completion time.
//
// Example usage:
//
//	defer span.End()
func (s *otelSpan) End() {
	if s.span == nil {
		return
	}
	s.span.End()
}

// AddEvent records an event in the span with optional attributes.
//
// Parameters:
//   - eventMessage string - description of the event
//   - attributes ...OtelAttribute - optional event attributes
//
// Example usage:
//
//	span.AddEvent("user-validated", otel.NewOtelAttr("valid", "true"))
func (s *otelSpan) AddEvent(eventMessage string, attributes ...OtelAttribute) {
	if s.span == nil {
		return
	}
	s.span.AddEvent(eventMessage, makeAttributes(attributes))
}

// SetStatus updates the span's status and adds a description.
//
// Parameters:
//   - status SpanStatus - status to set (OK or Error)
//   - description string - explanatory message for the status
//
// Example usage:
//
//	span.SetStatus(SpanStatusOK, "operation completed successfully")
func (s *otelSpan) SetStatus(status SpanStatus, description string) {
	if s.span == nil {
		return
	}
	s.span.SetStatus(status.otelStatus(), description)
}

// Success marks the span as successful with the given message.
//
// Parameters:
//   - message string - success message to record
//
// Example usage:
//
//	span.Success("user data processed successfully")
func (s *otelSpan) Success(message string) {
	s.SetStatus(SpanStatusOK, message)
}

// Error marks the span as failed with the given error message.
//
// Parameters:
//   - err error - error to record
//   - message string - error message to record
//
// Example usage:
//
//	span.Error(err, "failed to process user data: invalid format")
func (s *otelSpan) Error(err error, message string) {
	if s.span == nil {
		return
	}
	s.SetStatus(SpanStatusError, message)
	s.span.RecordError(err)
}

// otelStatus converts SpanStatus to OpenTelemetry status codes.
//
// Returns:
//   - codes.Code - corresponding OpenTelemetry status code
//
// Example usage:
//
//	code := status.otelStatus()
func (s *SpanStatus) otelStatus() codes.Code {
	switch *s {
	case SpanStatusOK:
		return codes.Ok
	case SpanStatusError:
		return codes.Error
	default:
		return codes.Unset
	}
}

func (k *SpanKind) otelKind() traceTypes.SpanKind {

	if kind, exists := kindsMap[*k]; exists {
		return kind
	}
	return traceTypes.SpanKindUnspecified
}

func makeSpanName(kind SpanKind, name string) string {
	if kind == SpanKindProducer {
		return fmt.Sprintf("send %s", name)
	}
	return fmt.Sprintf("process %s", name)
}

func (op *SpanOperation) String() string {

	if operation, exists := operationMap[*op]; exists {
		return operation
	}
	return "process"
}

func (mt *MessageSystemType) String() string {

	if value, exists := msMap[*mt]; exists {
		return value
	}

	return "internal"
}
