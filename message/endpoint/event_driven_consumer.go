// Package endpoint implements the event-driven-consumer pattern for message processing
// systems.
//
// This package provides a structure for consuming messages asynchronously and scalably,
// using multiple processors and integration with gateways and input channels. It
// facilitates the consumption, processing, and routing of messages in event-driven
// systems, with support for timeout, dead letter channels, and interceptors.
//
// The EventDrivenConsumer implementation supports:
// - Asynchronous message consumption with multiple concurrent processors
// - Integration with inbound channel adapters and gateways
// - Configurable processing timeouts and error handling
// - Graceful shutdown and resource cleanup
// - Dead letter channel support for failed messages
package endpoint

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/jeffersonbrasilino/gomes/container"
	"github.com/jeffersonbrasilino/gomes/message"
	"github.com/jeffersonbrasilino/gomes/message/handler"
	"github.com/jeffersonbrasilino/gomes/otel"
)

// EventDrivenConsumerBuilder is responsible for building EventDrivenConsumer instances.
// referenceName identifies the input channel to be consumed.
type EventDrivenConsumerBuilder struct {
	referenceName string
}

// EventDrivenConsumer represents an event-driven-consumer.
// Manages multiple processors, processing queue, and integration with gateway and
// input channel.
type EventDrivenConsumer struct {
	referenceName                 string
	processingTimeoutMilliseconds int
	gateway                       *Gateway
	inboundChannelAdapter         InboundChannelAdapter
	amountOfProcessors            int
	processingQueue               chan *message.Message
	processorsWaitGroup           sync.WaitGroup
	stopOnError                   bool
	RunCtx                        context.Context
	cancelRunCtx                  context.CancelFunc
	isRunning                     bool
	otelTrace                     otel.OtelTrace
}

// NewEventDrivenConsumerBuilder creates a new EventDrivenConsumerBuilder instance.
//
// Parameters:
//   - referenceName: reference name of the input channel
//
// Returns:
//   - *EventDrivenConsumerBuilder: pointer to EventDrivenConsumerBuilder
func NewEventDrivenConsumerBuilder(referenceName string) *EventDrivenConsumerBuilder {
	return &EventDrivenConsumerBuilder{
		referenceName: referenceName,
	}
}

// NewEventDrivenConsumer creates a new EventDrivenConsumer instance.
//
// Parameters:
//   - referenceName: reference name of the input channel
//   - gateway: pointer to the associated Gateway
//   - inboundChannelAdapter: input channel adapter
//
// Returns:
//   - *EventDrivenConsumer: pointer to EventDrivenConsumer
func NewEventDrivenConsumer(
	referenceName string,
	gateway *Gateway,
	inboundChannelAdapter InboundChannelAdapter,
) *EventDrivenConsumer {
	consumer := &EventDrivenConsumer{
		referenceName:                 referenceName,
		processingTimeoutMilliseconds: 100000,
		gateway:                       gateway,
		inboundChannelAdapter:         inboundChannelAdapter,
		amountOfProcessors:            1,
		stopOnError:                   true,
		isRunning:                     false,
		otelTrace:                     otel.InitTrace("event-driven-consumer"),
	}
	return consumer
}

// Build constructs an EventDrivenConsumer from the dependency container.
//
// Parameters:
//   - container: dependency container
//
// Returns:
//   - *EventDrivenConsumer: pointer to EventDrivenConsumer
//   - error: error if any occurs
func (b *EventDrivenConsumerBuilder) Build(
	container container.Container[any, any],
) (*EventDrivenConsumer, error) {

	anyChannel, err := container.Get(b.referenceName)
	if err != nil {
		return nil,
			fmt.Errorf(
				"[event-driven-consumer] consumer channel %s not found.",
				b.referenceName,
			)
	}

	inboundChannel, ok := anyChannel.(InboundChannelAdapter)
	if !ok {
		return nil,
			fmt.Errorf(
				"[event-driven-consumer] consumer channel %s is not a consumer channel.",
				b.referenceName,
			)
	}

	gatewayBuilder := NewGatewayBuilder(inboundChannel.ReferenceName(), "")

	if inboundChannel.DeadLetterChannelName() != "" {
		gatewayBuilder.WithDeadLetterChannel(inboundChannel.DeadLetterChannelName())
	}

	if len(inboundChannel.BeforeProcessors()) > 0 {
		gatewayBuilder.WithBeforeInterceptors(inboundChannel.BeforeProcessors()...)
	}

	if len(inboundChannel.AfterProcessors()) > 0 {
		gatewayBuilder.WithAfterInterceptors(inboundChannel.AfterProcessors()...)
	}

	if len(inboundChannel.RetryAttempts()) > 0 {
		gatewayBuilder.WithRetry(inboundChannel.RetryAttempts())
	}

	if ackChannel, ok := inboundChannel.(handler.ChannelMessageAcknowledgment); ok {
		gatewayBuilder.WithAcknowledge(ackChannel)
	}

	gateway, err := gatewayBuilder.Build(container)
	if err != nil {
		return nil, err
	}

	consumer := NewEventDrivenConsumer(
		b.referenceName,
		gateway,
		inboundChannel,
	)

	return consumer, nil
}

// WithMessageProcessingTimeout sets the message processing timeout in milliseconds.
//
// Parameters:
//   - milliseconds: timeout in milliseconds
//
// Returns:
//   - *EventDrivenConsumer: pointer to EventDrivenConsumer for method chaining
func (b *EventDrivenConsumer) WithMessageProcessingTimeout(
	milliseconds int,
) *EventDrivenConsumer {
	if milliseconds > 0 {
		b.processingTimeoutMilliseconds = milliseconds
	}
	return b
}

// WithAmountOfProcessors sets the number of concurrent processors.
//
// default value: 1
//
// Warning: If the order of message processing is crucial (such as data streaming),
// it is not recommended to configure this setting, as we do not guarantee the
// processing order in parallel goroutines.
//
// Parameters:
//   - value: number of processors
//
// Returns:
//   - *EventDrivenConsumer: pointer to EventDrivenConsumer for method chaining
func (b *EventDrivenConsumer) WithAmountOfProcessors(value int) *EventDrivenConsumer {
	if value > 1 {
		b.amountOfProcessors = value
	}
	return b
}

// WithStopOnError sets the stop run when error occured.
//
// default value: true
//
// Parameters:
//   - value: flag(bool)
//
// Returns:
//   - *EventDrivenConsumer: pointer to EventDrivenConsumer for method chaining
func (b *EventDrivenConsumer) WithStopOnError(value bool) *EventDrivenConsumer {
	b.stopOnError = value
	return b
}

// Run starts processing messages received from the input channel.
//
// Parameters:
//   - ctx: context for cancellation and timeout control
//
// Returns:
//   - error: error if any occurs
func (e *EventDrivenConsumer) Run(ctx context.Context) {
	e.isRunning = true
	slog.Info(
		"[event-driven-consumer] started.",
		"consumerName", e.referenceName,
	)

	e.RunCtx, e.cancelRunCtx = context.WithCancel(ctx)
	defer e.shutdown()
	defer e.cancelRunCtx()

	e.processingQueue = make(chan *message.Message, e.amountOfProcessors)
	e.startProcessorsNodes()

	for {
		beforeReceiveContextIsDone, consumerCtxErr := e.handleContext(e.RunCtx)
		if consumerCtxErr != nil {
			slog.Error("[event-driven-consumer] run error",
				"consumerName", e.referenceName,
				"error", consumerCtxErr,
			)
			if e.stopOnError {
				return
			}
		}

		if beforeReceiveContextIsDone {
			return
		}

		msg, err := e.inboundChannelAdapter.ReceiveMessage(e.RunCtx)
		if err != nil {
			if err != context.Canceled {
				slog.Error("[event-driven-consumer] message receive error",
					"consumerName", e.referenceName,
					"error", err,
				)
			}
			if e.stopOnError {
				return
			}
		}
		afterReceiveContextIsDone, _ := e.handleContext(e.RunCtx)
		if afterReceiveContextIsDone {
			return
		}

		select {
		case <-e.RunCtx.Done():
			return
		case e.processingQueue <- msg:
		}
	}
}

// sendToGateway sends the message to the gateway for processing.
//
// Parameters:
//   - msg: message to be processed
//   - nodeId: processor identifier
func (e *EventDrivenConsumer) sendToGateway(
	msg *message.Message,
	nodeId int,
) {
	contextRunHasDone, _ := e.handleContext(e.RunCtx)
	if contextRunHasDone {
		return
	}
	opCtx, cancel := context.WithTimeout(
		e.RunCtx,
		time.Duration(e.processingTimeoutMilliseconds)*time.Millisecond,
	)
	defer cancel()

	var span otel.OtelSpan
	if msg.GetContext() != nil {
		opCtx, span = e.otelTrace.Start(
			msg.GetContext(),
			fmt.Sprintf("Receive message %s", msg.GetHeaders().Route),
			otel.WithMessagingSystemType(otel.MessageSystemTypeInternal),
			otel.WithSpanOperation(otel.SpanOperationReceive),
			otel.WithSpanKind(otel.SpanKindConsumer),
			otel.WithMessage(msg),
		)
		defer span.End()
	}

	slog.Info("[event-driven-consumer] message processing started.",
		"consumer.name", e.referenceName,
		"consumer.nodeId", nodeId,
		"consumer.messageId", msg.GetHeaders().MessageId,
	)
	_, err := e.gateway.Execute(opCtx, msg)
	spanStatus := otel.SpanStatusOK
	if err != nil {
		spanStatus = otel.SpanStatusError
		slog.Error("[event-driven-consumer] processing message error.",
			"consumer.name", e.referenceName,
			"consumer.nodeId", nodeId,
			"consumer.messageId", msg.GetHeaders().MessageId,
			"consumer.error", err.Error(),
		)

		if span != nil {
			span.Error(err, "[event-driven-consumer] processing message error.")
		}

		if e.stopOnError {
			e.cancelRunCtx()
			return
		}
	}

	if span != nil {
		span.SetStatus(spanStatus, "[event-driven-consumer] message processed completed.")
	}

	slog.Info("[event-driven-consumer] message processed completed.",
		"consumer.name", e.referenceName,
		"consumer.nodeId", nodeId,
		"consumer.messageId", msg.GetHeaders().MessageId,
	)
}

func (e *EventDrivenConsumer) handleContext(ctx context.Context) (bool, error) {
	select {
	case <-ctx.Done():
		if ctx.Err() != nil {
			return true, ctx.Err()
		}
		return true, nil
	default:
	}
	return false, nil
}

// Stop requests the consumer to stop by canceling the internal context.
func (e *EventDrivenConsumer) Stop() {
	if e.cancelRunCtx == nil {
		return
	}
	e.cancelRunCtx()
}

// shutdown ends processing, closes the input channel and waits for processors to finish.
func (e *EventDrivenConsumer) shutdown() {

	if !e.isRunning {
		return
	}
	e.isRunning = false
	slog.Info("[event-driven-consumer] shutting down.",
		"consumerName", e.referenceName,
	)
	e.inboundChannelAdapter.Close()
	close(e.processingQueue)
	e.processorsWaitGroup.Wait()
}

// startProcessorsNodes starts concurrent processors to consume messages from the queue.
func (e *EventDrivenConsumer) startProcessorsNodes() {
	for i := 0; i < e.amountOfProcessors; i++ {
		e.processorsWaitGroup.Add(1)
		go func(workerId int) {
			defer e.processorsWaitGroup.Done()
			for {
				ctxIsDone, ctxErr := e.handleContext(e.RunCtx)
				if ctxIsDone {
					if ctxErr != nil {
						slog.Error("[event-driven-consumer] processor stopping",
							"consumer.name", e.referenceName,
							"consumer.nodeId", workerId,
							"error", ctxErr,
						)
					}
					return
				}
				msg, ok := <-e.processingQueue
				if !ok {
					slog.Debug("[event-driven-consumer] processor stopping",
						"consumer.name", e.referenceName,
						"consumer.nodeId", workerId,
						"reason", "queue closed",
					)
					return
				}
				if msg != nil {
					e.sendToGateway(msg, workerId)
				}
			}
		}(i)
	}
}
