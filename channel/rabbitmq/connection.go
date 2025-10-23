package rabbitmq

import (
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/wagslane/go-rabbitmq"
)

type (
	producerChannelType int8
	exchangeType        int8
)

const (
	ProducerQueue producerChannelType = iota
	ProducerExchange

	ExchangeFanout exchangeType = iota
	ExchangeDirect
	ExchangeTopic
	ExchangeHeaders
)

// conInstance holds the singleton connection instance for reuse across
// the application.
var conInstance *connection

type connection struct {
	name string
	host string
	conn *amqp.Connection
	/* vhost string
	ssl bool
	ssl_verify bool
	ssl_verify_name bool
	ssl_verify_peer bool
	ssl_verify_peer_name bool
	ssl_verify_peer_name_callback func(cert *x509.Certificate, host string) bool */
}

// NewConnection creates a new RabbitMQ connection instance. This implementation
// uses a singleton pattern to reuse the same connection across the application.
//
// Parameters:
//   - name: the connection name identifier
//   - host: list of RabbitMQ broker addresses
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

func (c *connection) Connect() error {
	con, err := amqp.Dial(fmt.Sprintf("amqp://%s", c.host))
	/* con, err := rabbitmq.NewConn(
		fmt.Sprintf("amqp://%s", c.host),
		rabbitmq.WithConnectionOptionsLogging,
	) */
	if err != nil {
		return err
	}
	c.conn = con
	return nil
}

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
		ch.QueueDeclare(
			channelName,
			true,  // durable
			false, // delete when unused
			false, // exclusive
			false, // no-wait
			nil,   // arguments
		)
		return ch, nil
	}

	ch.ExchangeDeclare(channelName,
		exchangeType.Type(),
		true,  // durable
		false, // auto-deleted
		false, // internal
		false, // no-wait
		nil,
	)

	return ch, nil
}

func (c *connection) Consumer(queueName string) *rabbitmq.Consumer {
	/* consumer, _ := rabbitmq.NewConsumer(
		c.conn,
		queueName,
		rabbitmq.WithConsumerOptionsRoutingKey("my_routing_key"),
		rabbitmq.WithConsumerOptionsExchangeName("events"),
		rabbitmq.WithConsumerOptionsExchangeDeclare,
	)
	return consumer */
	return nil
}

func (c *connection) Disconnect() error {
	return c.conn.Close()
}

func (c *connection) ReferenceName() string {
	return c.name
}
