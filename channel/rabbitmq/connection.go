// Package rabbitmq provides RabbitMQ integration for the message system,
// implementing connection management, channel adapters, and message translation
// for both producing and consuming messages through RabbitMQ.
//
// This package supports:
//   - Singleton connection pattern for efficient resource management
//   - Queue-based and Exchange-based message publishing
//   - Multiple exchange types (Direct, Fanout, Topic, Headers)
//   - Channel lifecycle management
//   - Message translation between internal and AMQP formats
package rabbitmq

import (
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

// Producer channel type constants.
const (
	// ProducerQueue publishes messages directly to a queue using the default
	// exchange.
	ProducerQueue producerChannelType = iota
	// ProducerExchange publishes messages to a specific exchange with routing
	// keys.
	ProducerExchange
)

// Exchange type constants following AMQP specification.
const (
	// ExchangeFanout broadcasts messages to all bound queues.
	ExchangeFanout exchangeType = iota
	// ExchangeDirect routes messages based on exact routing key match.
	ExchangeDirect
	// ExchangeTopic routes messages based on wildcard routing key patterns.
	ExchangeTopic
	// ExchangeHeaders routes messages based on message header attributes.
	ExchangeHeaders
)

// conInstance holds the singleton connection instance for reuse across the
// application.
var conInstance *connection

// producerChannelType defines the type of producer channel for RabbitMQ.
type producerChannelType int8

// exchangeType defines the type of exchange for RabbitMQ.
type exchangeType int8

// connection manages RabbitMQ broker connections with lifecycle management
// capabilities.
type connection struct {
	name string
	host string
	conn *amqp.Connection
}

// NewConnection creates a new RabbitMQ connection instance using a singleton
// pattern to reuse the same connection across the application.
//
// Parameters:
//   - name: the connection name identifier
//   - host: RabbitMQ broker address (format: host:port or
//     user:password@host:port/vhost)
//
// Returns:
//   - *connection: the connection instance
func NewConnection(name string, host string) *connection {
	if conInstance != nil {
		return conInstance
	}
	conInstance = &connection{
		name: name,
		host: host,
	}
	return conInstance
}

// Type returns the AMQP exchange type string representation.
//
// Returns:
//   - string: AMQP exchange type constant
func (t exchangeType) Type() string {
	switch t {
	case ExchangeFanout:
		return amqp.ExchangeFanout
	case ExchangeHeaders:
		return amqp.ExchangeHeaders
	case ExchangeTopic:
		return amqp.ExchangeTopic
	default:
		return amqp.ExchangeDirect
	}
}

// Connect establishes a connection to the RabbitMQ broker.
//
// Returns:
//   - error: error if connection establishment fails
func (c *connection) Connect() error {
	con, err := amqp.Dial(fmt.Sprintf("amqp://%s", c.host))
	if err != nil {
		return err
	}
	c.conn = con
	return nil
}

// Producer creates a new producer channel for the specified channel name and
// type. If the channel type is ProducerQueue, it declares a durable queue.
// If the channel type is ProducerExchange, it declares a durable exchange.
//
// Parameters:
//   - channelName: name of the queue or exchange
//   - channelType: type of producer channel (ProducerQueue or ProducerExchange)
//   - exchangeType: type of exchange if channelType is ProducerExchange
//
// Returns:
//   - *amqp.Channel: configured RabbitMQ channel
//   - error: error if channel creation or declaration fails
func (c *connection) Producer(
	channelName string,
	channelType producerChannelType,
	exchangeType exchangeType,
) (*amqp.Channel, error) {
	ch, err := c.conn.Channel()
	if err != nil {
		return nil, err
	}

	if channelType == ProducerQueue {
		_, err = ch.QueueDeclare(
			channelName,
			true,  // durable
			false, // delete when unused
			false, // exclusive
			false, // no-wait
			nil,   // arguments
		)
		return ch, err
	}

	err = ch.ExchangeDeclare(
		channelName,
		exchangeType.Type(),
		true,  // durable
		false, // auto-deleted
		false, // internal
		false, // no-wait
		nil,   // arguments
	)

	return ch, err
}

// Consumer creates a new consumer channel for receiving messages from the
// specified queue.
//
// Parameters:
//   - queueName: name of the queue to consume from
//
// Returns:
//   - *amqp.Channel: configured RabbitMQ channel for consuming
//   - error: error if channel creation fails
func (c *connection) Consumer(queueName string) (*amqp.Channel, error) {
	return c.conn.Channel()
}

// Disconnect closes the RabbitMQ connection and releases associated resources.
//
// Returns:
//   - error: error if disconnection fails (typically nil)
func (c *connection) Disconnect() error {
	return c.conn.Close()
}

// ReferenceName returns the connection name identifier.
//
// Returns:
//   - string: the connection name
func (c *connection) ReferenceName() string {
	return c.name
}
