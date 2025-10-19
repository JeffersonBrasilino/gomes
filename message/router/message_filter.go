// Package router provides message routing components for the message system.
//
// This package implements various routing patterns from Enterprise Integration
// Patterns, enabling flexible message routing and processing through different
// channels and handlers. It provides composite routing, recipient list routing,
// and message filtering capabilities.
//
// The MessageFilter implementation supports:
// - Message filtering based on custom criteria
// - Conditional message processing
// - Flexible filtering strategies
// - Integration with routing pipelines
package router

import (
	"context"

	"github.com/hex-api-go/pkg/core/infrastructure/gomes/message"
)

// FilterFunc defines the contract for message filtering functions.
// It takes a message and returns true if the message should be processed,
// false if it should be filtered out.
type FilterFunc func(message.Message) bool

// messageFilter implements message filtering functionality, allowing messages
// to be conditionally processed based on custom filtering criteria.
type messageFilter struct {
	filterFunc FilterFunc
}

// NewMessageFilter creates a new message filter instance with the specified
// filtering function.
//
// Parameters:
//   - filterFunc: the function to use for filtering messages
//
// Returns:
//   - *messageFilter: configured message filter
func NewMessageFilter(filterFunc FilterFunc) *messageFilter {
	return &messageFilter{filterFunc: filterFunc}
}

// Handle processes a message by applying the filter function. If the filter
// returns true, the message is passed through; otherwise, it returns nil.
//
// Parameters:
//   - ctx: context for timeout/cancellation control (unused in current implementation)
//   - msg: the message to be filtered
//
// Returns:
//   - *message.Message: the message if it passes the filter, nil otherwise
//   - error: always nil in current implementation
func (f *messageFilter) Handle(
	ctx context.Context,
	msg *message.Message,
) (*message.Message, error) {
	filterResult := f.filterFunc(*msg)
	if filterResult {
		return msg, nil
	}

	return nil, nil
}
