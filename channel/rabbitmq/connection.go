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

// conInstance holds the singleton connection instance for reuse across the
// application.
var conInstance *connection


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

func (c *connection) GetConnection() *amqp.Connection {
	return c.conn
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
