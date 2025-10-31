// Package kafka provides Kafka integration for the message system.
//
// This package implements Kafka-specific channel adapters and connections for
// publishing and consuming messages through Apache Kafka. It provides outbound
// and inbound channel adapters with message translation capabilities.
//
// The Connection implementation supports:
// - Kafka producer and consumer connection management
// - Connection pooling and reuse
// - Error handling and connection lifecycle
// - Configuration management for Kafka clients
package kafka

import (
	"github.com/segmentio/kafka-go"
)

// connection manages Kafka producer and consumer connections with lifecycle
// management capabilities.
type connection struct {
	name             string
	host             []string
	producerInstance *kafka.Writer
	consumerConfig   *kafka.ReaderConfig
}

// conInstance holds the singleton connection instance for reuse across
// the application.
var conInstance *connection

// NewConnection creates a new Kafka connection instance. This implementation
// uses a singleton pattern to reuse the same connection across the application.
//
// Parameters:
//   - name: the connection name identifier
//   - host: list of Kafka broker addresses
//
// Returns:
//   - *connection: the connection instance
func NewConnection(name string, host []string) *connection {
	if conInstance != nil {
		return conInstance
	}
	conInstance = &connection{
		name: name,
		host: host,
	}
	return conInstance
}

// Connect establishes connections to Kafka brokers for both producer and consumer.
// It configures the Kafka client with appropriate settings for reliability.
//
// Returns:
//   - error: error if connection establishment fails
func (c *connection) Connect() error {
	c.producerInstance = &kafka.Writer{
		Addr: kafka.TCP(c.host...),
	}
	c.consumerConfig = &kafka.ReaderConfig{
		Brokers:  c.host,
		MaxBytes: 10e6,
	}
	return nil
}

// GetProducer returns the Kafka sync producer instance.
//
// Returns:
//   - sarama.SyncProducer: the Kafka producer
func (c *connection) Producer() *kafka.Writer {
	return c.producerInstance
}

// GetConsumer returns the Kafka consumer instance.
//
// Returns:
//   - *kafka.Reader: the Kafka consumer
func (c *connection) Consumer(topic string, groupId string) *kafka.Reader {
	consumerConfig := *c.consumerConfig
	consumerConfig.GroupID = groupId
	consumerConfig.Topic = topic
	return kafka.NewReader(consumerConfig)
}

// Disconnect closes the Kafka connections and releases associated resources.
//
// Returns:
//   - error: error if disconnection fails (typically nil)
func (c *connection) Disconnect() error {
	return c.producerInstance.Close()
}

// ReferenceName returns the connection name identifier.
//
// Returns:
//   - string: the connection name
func (c *connection) ReferenceName() string {
	return c.name
}
