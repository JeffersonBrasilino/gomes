// Package router provides message routing components for the message system.
//
// This package implements various routing patterns from Enterprise Integration
// Patterns, enabling flexible message routing and processing through different
// channels and handlers. It provides composite routing, recipient list routing,
// and message filtering capabilities.
//
// The RecipientListRouter implementation supports:
// - Dynamic message routing based on message headers
// - Container-based channel resolution
// - Flexible routing strategies
// - Error handling for missing channels
package router

import (
	"context"
	"fmt"

	"github.com/jeffersonbrasilino/gomes/container"
	"github.com/jeffersonbrasilino/gomes/message"
)

// recipientListRouter implements the Recipient List pattern, routing messages
// to specific channels based on message headers and container configuration.
type recipientListRouter struct {
	gomesContainer container.Container[any, any]
}

// NewRecipientListRouter creates a new recipient list router instance.
//
// Parameters:
//   - gomesContainer: container for resolving channel references
//
// Returns:
//   - *recipientListRouter: configured recipient list router
func NewRecipientListRouter(
	gomesContainer container.Container[any, any],
) *recipientListRouter {
	return &recipientListRouter{gomesContainer: gomesContainer}
}

// Handle routes a message to the appropriate channel based on message headers.
// The router determines the target channel using channel name or route information.
//
// Parameters:
//   - ctx: context for timeout/cancellation control
//   - msg: the message to be routed
//
// Returns:
//   - *message.Message: the original message if routing succeeds
//   - error: error if the target channel is not found
func (r *recipientListRouter) Handle(
	ctx context.Context,
	msg *message.Message,
) (*message.Message, error) {
	route := r.chooseRoute(msg)
	actionChannel, err := r.gomesContainer.Get(route)

	if err != nil {
		return nil, fmt.Errorf(
			"[recipient-list-router] unprocessable message, handler for action %v not exists",
			route,
		)
	}

	channel, ok := actionChannel.(message.PublisherChannel)
	if !ok {
		return nil, fmt.Errorf(
			"[recipient-list-router] unprocessable message, channel for action %v does not implement PublisherChannel",
			route,
		)
	}

	channel.Send(ctx, msg)

	return msg, nil
}

// chooseRoute determines the appropriate route for a message based on its headers.
// It prioritizes ChannelName over Route if both are present.
//
// Parameters:
//   - msg: the message to determine routing for
//
// Returns:
//   - string: the determined route name
func (r *recipientListRouter) chooseRoute(msg *message.Message) string {
	var route string
	if msg.GetHeader().Get(message.HeaderChannelName) != "" {
		route = msg.GetHeader().Get(message.HeaderChannelName)
	}

	if msg.GetHeader().Get(message.HeaderRoute) != "" && route == "" {
		route = msg.GetHeader().Get(message.HeaderRoute)
	}
	return route
}
