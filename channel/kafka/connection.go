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
	"crypto/tls"

	"github.com/segmentio/kafka-go"
)

// connection manages Kafka producer and consumer connections with lifecycle
// management capabilities.
type connection struct {
	name      string
	host      []string
	tlsConfig *tls.Config
	transport *kafka.Transport
	dialer    *kafka.Dialer
}

// ConnectionOptions is a functional option for configuring Kafka connections
// with TLS.
type ConnectionOptions func(*connectionOptions)

type connectionOptions struct {
	tlsConfig *tls.Config
}

// WithTlsConfig sets the TLS configuration for the Kafka connection.
// This allows secure communication with Kafka brokers.
//
// Parameters:
//   - tlsConfig: TLS configuration to use for the connection
//
// Returns:
//   - ConnectionOptions: configured option function
func WithTlsConfig(tlsConfig *tls.Config) ConnectionOptions {
	return func(opt *connectionOptions) {
		opt.tlsConfig = tlsConfig
	}
}

// NewConnection creates a new Kafka connection instance. This implementation
// uses a singleton pattern to reuse the same connection across the application.
//
// Parameters:
//   - name: the connection name identifier
//   - host: list of Kafka broker addresses
//
// Returns:
//   - *connection: the connection instance
func NewConnection(name string, host []string, opts ...ConnectionOptions) *connection {
	connectionOptions := &connectionOptions{}
	for _, opt := range opts {
		opt(connectionOptions)
	}

	return &connection{
		name:      name,
		host:      host,
		tlsConfig: connectionOptions.tlsConfig,
	}
}

// Connect establishes connections to Kafka brokers for both producer and consumer.
// It configures the Kafka client with appropriate settings for reliability.
//
// Returns:
//   - error: error if connection establishment fails
func (c *connection) Connect() error {
	c.dialer = &kafka.Dialer{
		ClientID:  c.name,
		DualStack: true,
		TLS:       c.tlsConfig,
	}

	c.transport = &kafka.Transport{
		ClientID: c.name,
		Dial:     c.dialer.DialFunc,
		TLS:      c.tlsConfig,
	}
	return nil
}

// getTransport returns the Kafka transport configured for this connection.
// The transport is used by producers to send messages to Kafka.
func (c *connection) getTransport() *kafka.Transport {
	return c.transport
}

// getDialer returns the Kafka dialer configured for this connection.
// The dialer is used to establish connections to Kafka brokers.
func (c *connection) getDialer() *kafka.Dialer {
	return c.dialer
}

// getHost returns the list of Kafka broker addresses for this connection.
func (c *connection) getHost() []string {
	return c.host
}

// ReferenceName returns the connection name identifier.
//
// Returns:
//   - string: the connection name
func (c *connection) ReferenceName() string {
	return c.name
}

func (c *connection) Disconnect() error {
	return nil
}
