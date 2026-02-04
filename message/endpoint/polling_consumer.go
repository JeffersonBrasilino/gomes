// // Package endpoint provides polling consumer functionality for message processing.
// //
// // This package implements the Polling Consumer pattern from Enterprise Integration
// // Patterns, enabling applications to periodically check for messages and process
// // them. It provides a configurable polling mechanism with support for timeouts,
// // error handling, and graceful shutdown.
// //
// // The PollingConsumer implementation supports:
// // - Configurable polling intervals and processing delays
// // - Processing timeout management
// // - Error handling with configurable stop-on-error behavior
// // - Graceful shutdown and resource cleanup
// // - Integration with gateways and inbound channel adapters
package endpoint

// import (
// 	"context"
// 	"fmt"
// 	"log/slog"
// 	"sync"
// 	"time"

// 	"github.com/jeffersonbrasilino/gomes/container"
// 	"github.com/jeffersonbrasilino/gomes/message"
// )

// // TODO: URGENTE, REVISAR. refazer o polling consumer, provavlemente este started consumer não precisa mais.
// // startedConsumers tracks which consumers have been started to prevent duplicate
// // initialization.
// var startedConsumers sync.Map

// // PollingConsumerBuilder provides a builder pattern for creating PollingConsumer
// // instances with specific configurations.
// type PollingConsumerBuilder struct {
// 	referenceName string
// }

// // PollingConsumer represents a consumer that periodically polls for messages
// // and processes them through configured gateways.
// type PollingConsumer struct {
// 	referenceName                 string
// 	pollIntervalMilliseconds      int
// 	processingDelayMilliseconds   int
// 	processingTimeoutMilliseconds int
// 	stopOnError                   bool
// 	hasRunning                    bool
// 	gateway                       *Gateway
// 	inboundChannelAdapter         InboundChannelAdapter
// }

// // NewPolllingConsumerBuilder creates a new polling consumer builder instance.
// //
// // Parameters:
// //   - referenceName: unique identifier for the consumer
// //
// // Returns:
// //   - *PollingConsumerBuilder: configured builder instance
// func NewPolllingConsumerBuilder(referenceName string) *PollingConsumerBuilder {
// 	return &PollingConsumerBuilder{
// 		referenceName: referenceName,
// 	}
// }

// // NewPollingConsumer creates a new polling consumer instance.
// //
// // Parameters:
// //   - gateway: the gateway to use for message processing
// //   - inboundChannelAdapter: the adapter for receiving messages
// //   - referenceName: unique identifier for the consumer
// //
// // Returns:
// //   - *PollingConsumer: configured polling consumer
// func NewPollingConsumer(
// 	gateway *Gateway,
// 	inboundChannelAdapter InboundChannelAdapter,
// 	referenceName string,
// ) *PollingConsumer {
// 	return &PollingConsumer{
// 		pollIntervalMilliseconds:      1000,
// 		processingDelayMilliseconds:   0,
// 		processingTimeoutMilliseconds: 100000,
// 		stopOnError:                   true,
// 		gateway:                       gateway,
// 		inboundChannelAdapter:         inboundChannelAdapter,
// 		referenceName:                 referenceName,
// 	}
// }

// // Build constructs a PollingConsumer from the dependency container.
// //
// // Parameters:
// //   - container: dependency container containing required components
// //
// // Returns:
// //   - *PollingConsumer: configured polling consumer
// //   - error: error if construction fails
// func (b *PollingConsumerBuilder) Build(
// 	container container.Container[any, any],
// ) (*PollingConsumer, error) {

// 	_, hasExists := startedConsumers.Load(b.referenceName)
// 	if hasExists {
// 		return nil, fmt.Errorf("consumer %s already started", b.referenceName)
// 	}

// 	channel, ok := container.Get(b.referenceName)
// 	if ok != nil {
// 		panic(fmt.Sprintf("consumer channel %s not found.", b.referenceName))
// 	}

// 	inboundChannel, instance := channel.(InboundChannelAdapter)
// 	if !instance {
// 		panic(fmt.Sprintf("consumer channel %s is not a consumer channel.", b.referenceName))
// 	}

// 	gateway, err := container.Get(GatewayReferenceName(b.referenceName))
// 	if err != nil {
// 		return nil, fmt.Errorf(
// 			"[polling-consumer] gateway %s does not exist",
// 			b.referenceName,
// 		)
// 	}

// 	startedConsumers.Store(b.referenceName, true)

// 	return NewPollingConsumer(
// 		gateway.(*Gateway),
// 		inboundChannel,
// 		b.referenceName,
// 	), nil
// }

// // WithPollIntervalMilliseconds sets the polling interval in milliseconds.
// //
// // Parameters:
// //   - value: polling interval in milliseconds
// //
// // Returns:
// //   - *PollingConsumer: consumer instance for method chaining
// func (b *PollingConsumer) WithPollIntervalMilliseconds(value int) *PollingConsumer {
// 	b.pollIntervalMilliseconds = value
// 	return b
// }

// // WithProcessingDelayMilliseconds sets the processing delay in milliseconds.
// //
// // Parameters:
// //   - value: processing delay in milliseconds
// //
// // Returns:
// //   - *PollingConsumer: consumer instance for method chaining
// func (b *PollingConsumer) WithProcessingDelayMilliseconds(value int) *PollingConsumer {
// 	b.processingDelayMilliseconds = value
// 	return b
// }

// // WithStopOnError sets whether the consumer should stop on processing errors.
// //
// // Parameters:
// //   - value: true to stop on error, false to continue
// //
// // Returns:
// //   - *PollingConsumer: consumer instance for method chaining
// func (b *PollingConsumer) WithStopOnError(value bool) *PollingConsumer {
// 	b.stopOnError = value
// 	return b
// }

// // WithProcessingTimeoutMilliseconds sets the processing timeout in milliseconds.
// //
// // Parameters:
// //   - value: processing timeout in milliseconds
// //
// // Returns:
// //   - *PollingConsumer: consumer instance for method chaining
// func (b *PollingConsumer) WithProcessingTimeoutMilliseconds(value int) *PollingConsumer {
// 	b.processingTimeoutMilliseconds = value
// 	return b
// }

// // Run starts the polling consumer and begins processing messages.
// //
// // Parameters:
// //   - ctx: context for cancellation and timeout control
// //
// // Returns:
// //   - error: error if polling fails or context is cancelled
// func (c *PollingConsumer) Run(ctx context.Context) error {
// 	slog.Info("Starting polling consumer", "consumerName", c.referenceName)
// 	c.hasRunning = true

// 	ticker := time.NewTicker(
// 		time.Millisecond * time.Duration(c.pollIntervalMilliseconds),
// 	)
// 	defer ticker.Stop()

// 	for c.hasRunning {
// 		select {
// 		case <-ctx.Done():
// 			c.Stop()
// 			return ctx.Err()
// 		case <-ticker.C:
// 			msg, err := c.inboundChannelAdapter.ReceiveMessage(ctx)
// 			if err != nil {
// 				slog.Error("Error receiving message", "error", err, "name", c.referenceName)
// 				if c.stopOnError {
// 					c.Stop()
// 					return err
// 				}
// 				continue
// 			}
// 			if msg == nil {
// 				slog.Info("no message received", "consumerName", c.referenceName)
// 				continue
// 			}

// 			if c.processingDelayMilliseconds > 0 {
// 				time.Sleep(
// 					time.Millisecond * time.Duration(c.processingDelayMilliseconds),
// 				)
// 			}

// 			go c.sendToGateway(ctx, msg)
// 		}
// 	}

// 	return nil
// }

// // sendToGateway sends a message to the gateway for processing with timeout support.
// //
// // Parameters:
// //   - ctx: context for timeout/cancellation control
// //   - msg: the message to be processed
// func (c *PollingConsumer) sendToGateway(ctx context.Context, msg *message.Message) {
// 	fmt.Println("sendToGateway", msg)
// 	opCtx, cancel := context.WithTimeout(
// 		ctx,
// 		time.Duration(c.processingTimeoutMilliseconds)*time.Millisecond,
// 	)
// 	defer cancel()
// 	_, err := c.gateway.Execute(opCtx, msg)
// 	if err != nil {
// 		slog.Error("failed to process message",
// 			"error", err,
// 			"name", c.referenceName,
// 			"messageId", msg.GetHeaders().MessageId,
// 		)
// 		return
// 	}
// 	slog.Debug("message processed", "name", c.referenceName)
// }

// // Stop stops the polling consumer and cleans up resources.
// func (c *PollingConsumer) Stop() {
// 	c.hasRunning = false
// 	startedConsumers.Delete(c.referenceName)
// }
