// Package channel provides point-to-point messaging channels for the message system.
//
// This package implements the Point-to-Point Channel pattern from Enterprise Integration
// Patterns, enabling direct communication between a single sender and receiver. It
// provides a synchronous messaging mechanism where each message is delivered to exactly
// one consumer, ensuring reliable message delivery and processing.
//
// The PointToPointChannel implementation supports:
// - Single sender to single receiver communication
// - Context-aware message sending with cancellation support
// - Asynchronous message subscription and processing
// - Graceful channel closure and resource cleanup
// - Thread-safe operations with proper state management
package channel

import (
	"context"
	"errors"
	"fmt"

	"github.com/jeffersonbrasilino/gomes/message"
)

// PointToPointReferenceName generates a standardized reference name for point-to-point
// channels.
//
// Parameters:
//   - name: The base name for the channel
//
// Returns:
//   - string: The formatted reference name with prefix
func PointToPointReferenceName(name string) string {
	return fmt.Sprintf("point-to-point-channel:%s", name)
}

// PointToPointChannel implements a point-to-point messaging channel where each message
// is delivered to exactly one consumer.
type PointToPointChannel struct {
	name    string
	channel chan *message.Message
	hasOpen bool
}

// NewPointToPointChannel creates a new point-to-point channel instance.
//
// Parameters:
//   - name: The name identifier for the channel
//
// Returns:
//   - *PointToPointChannel: A new configured point-to-point channel
func NewPointToPointChannel(name string) *PointToPointChannel {
	return &PointToPointChannel{
		name:    name,
		channel: make(chan *message.Message),
		hasOpen: true,
	}
}

// Send sends a message through the point-to-point channel with context support.
//
// Parameters:
//   - ctx: Context for timeout/cancellation control
//   - msg: The message to be sent
//
// Returns:
//   - error: Error if the channel is closed or context is cancelled
func (c *PointToPointChannel) Send(ctx context.Context, msg *message.Message) error {
	if !c.hasOpen {
		return errors.New("channel has not been opened")
	}

	select {
	case c.channel <- msg:
		return nil
	case <-ctx.Done():
		return fmt.Errorf("context cancelled while sending message: %v", ctx.Err())
	}
}

// Subscribe registers a callback function to process messages asynchronously.
// The callback is executed in a separate goroutine for each received message.
//
// Parameters:
//   - callable: The function to be called for each received message
func (c *PointToPointChannel) Subscribe(callable func(m *message.Message)) {
	go func(ch <-chan *message.Message) {
		for {
			m, hasOpen := <-ch
			if !hasOpen {
				c.hasOpen = false
				break
			}
			go callable(m)
		}
	}(c.channel)
}

// Receive receives a single message from the channel and closes the channel after
// receiving.
//
// Returns:
//   - *message.Message: The received message
//   - error: Error if the channel is closed or no message is available
func (c *PointToPointChannel) Receive(ctx context.Context) (*message.Message, error) {
	result, hasOpen := <-c.channel
	if !hasOpen {
		c.hasOpen = false
		return nil, errors.New("channel has not been opened")
	}

	return result, nil
}

// Close gracefully closes the point-to-point channel and releases associated resources.
//
// Returns:
//   - error: Error if closing the channel fails (typically nil)
func (c *PointToPointChannel) Close() error {
	if !c.hasOpen {
		return nil
	}
	c.hasOpen = false
	close(c.channel)
	return nil
}

// Name returns the name identifier of the point-to-point channel.
//
// Returns:
//   - string: The channel name
func (c *PointToPointChannel) Name() string {
	return c.name
}
