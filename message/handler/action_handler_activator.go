// Package handler provides message handling components for the message system.
//
// This package implements various message handlers that process and route messages
// through the system. It provides specialized handlers for different message
// processing scenarios including action handling, context management, and error
// handling patterns.
//
// The ActionHandlerActivator implementation supports:
// - Generic action handling with type safety
// - Action routing and processing
// - Reply channel integration
// - Error handling and response management
package handler

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jeffersonbrasilino/gomes/container"
	"github.com/jeffersonbrasilino/gomes/message"
	"github.com/jeffersonbrasilino/gomes/message/channel"
)

// Action defines the contract for actions that can be processed by the system.
type Action interface {
	Name() string
}

// ActionHandler defines the contract for handling specific action types with
// generic input and output types.
type ActionHandler[T Action, U any] interface {
	Handle(ctx context.Context, payload T) (U, error)
}

// ActionHandleActivatorBuilder provides a builder pattern for creating action
// handler activators with specific configurations.
type ActionHandleActivatorBuilder[TInput Action, TOutput any] struct {
	referenceName string
	handler       ActionHandler[TInput, TOutput]
}

type MessageHeaderAccessor interface {
	SetMessageHeader(header message.Header)
}

// ActionHandleActivator processes actions by delegating to the appropriate
// handler and managing the response through reply channels.
type ActionHandleActivator[
	THandler ActionHandler[TInput, TOutput],
	TInput Action,
	TOutput any,
] struct {
	handler THandler
}

// NewActionHandleActivatorBuilder creates a new action handler activator builder
// instance.
//
// Parameters:
//   - referenceName: unique identifier for the activator
//   - handler: the action handler to use for processing
//
// Returns:
//   - *ActionHandleActivatorBuilder[TInput, TOutput]: configured builder instance
func NewActionHandleActivatorBuilder[TInput Action, TOutput any](
	referenceName string,
	handler ActionHandler[TInput, TOutput],
) *ActionHandleActivatorBuilder[TInput, TOutput] {
	return &ActionHandleActivatorBuilder[TInput, TOutput]{
		referenceName: referenceName,
		handler:       handler,
	}
}

// NewActionHandlerActivator creates a new action handler activator instance.
//
// Parameters:
//   - handler: the action handler to use for processing
//
// Returns:
//   - *ActionHandleActivator[THandler, TInput, TOutput]: configured activator
func NewActionHandlerActivator[THandler ActionHandler[TInput, TOutput], TInput Action, TOutput any](
	handler THandler,
) *ActionHandleActivator[THandler, TInput, TOutput] {
	return &ActionHandleActivator[THandler, TInput, TOutput]{
		handler: handler,
	}
}

// ReferenceName returns the reference name of the activator builder.
//
// Returns:
//   - string: the reference name
func (b *ActionHandleActivatorBuilder[TInput, TOutput]) ReferenceName() string {
	return b.referenceName
}

// Build constructs an action handler activator from the dependency container.
//
// Parameters:
//   - container: dependency container containing required components
//
// Returns:
//   - message.PublisherChannel: configured publisher channel for the activator
//   - error: error if construction fails
func (b *ActionHandleActivatorBuilder[TInput, TOutput]) Build(
	container container.Container[any, any],
) (message.PublisherChannel, error) {
	handlerActivator := NewActionHandlerActivator(b.handler)
	chn := channel.NewPointToPointChannel(b.referenceName)
	chn.Subscribe(func(msg *message.Message) {
		handlerActivator.Handle(msg.GetContext(), msg)
	})
	return chn, nil
}

// Handle processes an action message by delegating to the appropriate handler
// and managing the response through reply channels.
//
// When it is an external message, the payload MUST be of type []byte
//
// Parameters:
//   - ctx: context for timeout/cancellation control
//   - msg: the message containing the action to be processed
//
// Returns:
//   - *message.Message: the result message
//   - error: error if processing fails
func (c *ActionHandleActivator[THandler, TInput, TOutput]) Handle(
	ctx context.Context,
	msg *message.Message,
) (*message.Message, error) {

	var action TInput
	action, ok := msg.GetPayload().(TInput)
	resultMessageBuilder := message.NewMessageBuilder().
		WithChannelName(msg.GetInternalReplyChannel().Name()).
		WithMessageType(message.Document).
		WithCorrelationId(msg.GetHeader().Get(message.HeaderCorrelationId))

	if replyToMessage := msg.GetHeader().Get(message.HeaderReplyTo); replyToMessage != "" {
		resultMessageBuilder.WithChannelName(replyToMessage)
	}

	if !ok {
		payload, ok := msg.GetPayload().([]byte)
		if !ok {
			err := fmt.Errorf(
				"[action-handler] cannot process action: incorrect contract data",
			)
			resultMessageBuilder.WithPayload(err)
			c.sendResponseToReplyChannel(ctx, msg, resultMessageBuilder.Build())
			return nil, err
		}

		errUnmsl := json.Unmarshal(payload, &action)
		if errUnmsl != nil {
			err := fmt.Errorf(
				"[action-handler] cannot process action: %v", errUnmsl.Error(),
			)
			resultMessageBuilder.WithPayload(err)
			c.sendResponseToReplyChannel(ctx, msg, resultMessageBuilder.Build())
			return nil, err
		}
	}

	if accessor, ok := any(c.handler).(MessageHeaderAccessor); ok {
		accessor.SetMessageHeader(msg.GetHeader())
	}

	output, err := c.executeAction(ctx, action)

	if err != nil {
		resultMessageBuilder.WithPayload(err)
	} else {
		resultMessageBuilder.WithPayload(output)
	}

	resultMessage := resultMessageBuilder.Build()
	c.sendResponseToReplyChannel(ctx, msg, resultMessage)

	return resultMessage, err
}

// executeAction executes the action using the configured handler.
//
// Parameters:
//   - ctx: context for timeout/cancellation control
//   - args: the action to be executed
//
// Returns:
//   - TOutput: the result of the action execution
//   - error: error if execution fails
func (c *ActionHandleActivator[THandler, TInput, TOutput]) executeAction(
	ctx context.Context,
	args TInput,
) (TOutput, error) {
	result, err := c.handler.Handle(ctx, args)
	return result, err
}

func (c *ActionHandleActivator[THandler, TInput, TOutput]) sendResponseToReplyChannel(
	ctx context.Context,
	requestMessage,
	responseMessage *message.Message,
) {
	replyChannel := requestMessage.GetInternalReplyChannel()
	if replyChannel != nil {
		replyChannel.Send(ctx, responseMessage)
	}
}
