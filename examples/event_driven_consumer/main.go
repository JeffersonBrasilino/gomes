package main

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/jeffersonbrasilino/gomes"
	kafka "github.com/jeffersonbrasilino/gomes/channel/kafka"
	"github.com/jeffersonbrasilino/gomes/message"
	"github.com/jeffersonbrasilino/gomes/otel"
)

// Underneath the hood,
// we use the CQRS pattern to process commands, queries, and events.
// First, we need to create an action (command, query, or event) and a handler for this action.
// This is a basic example.

// cqrs action
type Command struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func CreateCommand(Username, Password string) *Command {
	return &Command{
		Username,
		Password,
	}
}

// this function is responsible for the record name and action routing
func (c *Command) Name() string {
	return "createUser"
}

// CQRS acton handler
type CommandHandler struct {
	tracer otel.OtelTrace
	header map[string]string
}

// response structure
type ResultCm struct {
	Result any
}

func NewComandHandler() *CommandHandler {
	return &CommandHandler{
		tracer: otel.InitTrace("command-handler"),
	}
}

// note that the link between the action and its handler is the type of the data parameter.
// This indicates that this handler is responsible for this action
func (c *CommandHandler) Handle(ctx context.Context, data *Command) (*ResultCm, error) {
	time.Sleep(time.Second * 1)
	ctx, span := c.tracer.Start(
		ctx,
		"Handle Command",
	)
	defer span.End()

	slog.Info("processing command...",
		"username", data.Username,
	)
	time.Sleep(time.Second * 5)
	slog.Info("command processed.",
		"username", data.Username,
	)

	return &ResultCm{Result: "DEU BOM"}, nil
	//return nil, fmt.Errorf("DEU RUIM AO PROCESSAR A MENSAGEM")
}

//when async handler and header is required for the processing.
//Gomes message core inject the header using this method (satifying the MessageHeaderAccessor contract) before handle message.
func (c *CommandHandler) SetMessageHeader(header message.Header) {
	c.header = header
}

func main() {

	ctx, stop := context.WithCancel(context.Background())
	defer stop()
	slog.Info("start message system consumer....")

	//connection channels
	gomes.AddChannelConnection(
		kafka.NewConnection("defaultConKafka", []string{"kafka:9092"}),
	)

	// consumer channel config 
	topicConsumerChannel := kafka.NewConsumerChannelAdapterBuilder(
		"defaultConKafka",
		"gomes.topic",
		"test_consumer",
	)

	//configure reply channel using the replyTo header of the message. This way, we can have dynamic reply channels.
	//topicConsumerChannel.WithSendReplyUsingReplyTo()
	
	//configure retries
	//topicConsumerChannel.WithRetryTimes(2_000, 5_000)

	//configure dlq channel name. This way, we can have a default dlq channel for all consumer channels that don't have a specific dlq channel configured.
	//topicConsumerChannel.WithDeadLetterChannelName("gomes.dlq")

	//add consumer channel to the system
	gomes.AddConsumerChannel(topicConsumerChannel)

	//response channel configuration. This channel will be used to send the response of the message processing.
	/* responseChannel := kafka.NewPublisherChannelAdapterBuilder(
		"defaultConKafka",
		"gomes.response",
	)
	gomes.AddPublisherChannel(responseChannel)

	//DLQ channel configuration.
	dlqChannel := kafka.NewPublisherChannelAdapterBuilder(
		"defaultConKafka",
		"gomes.dlq",
	)
	gomes.AddPublisherChannel(dlqChannel) */

	// Register CQRS action and action handler.
	gomes.AddActionHandler(NewComandHandler())

	//enable otel trace for the message system
	//gomes.EnableOtelTrace()

	//start the message system
	gomes.Start()

	go publishMessage()

	//For the consumer channel endpoint,
	//the advantage of having an abstraction between the consumer channel and the consumer endpoint
	//is that we can have two different endpoints for the same channel (event-driven or polling).
	//Note that the consumerName parameter of the eventDrivenConsumer method is the same as the consumer name of the consumerChannel.
	consumer, err := gomes.EventDrivenConsumer("test_consumer")
	if err != nil {
		panic(err)
	}

	go func() {
		err := consumer.WithAmountOfProcessors(1).
			WithMessageProcessingTimeout(4000).
			WithStopOnError(false).
			Run(ctx)

		fmt.Println("main.go erro no consumer", "erro", err)

		if err != nil {
			stop()
		}
	}()

	<-ctx.Done()
	//message system graceful shutdown
	gomes.Shutdown()
	fmt.Println("CONSUMIDOR STOPPED COM SUCESSO...")
}

func publishMessage() {
	maxPublishMessages := 10
	for i := 1; i <= maxPublishMessages; i++ {
		fmt.Println("publish command message...")
		busA, _ := gomes.CommandBusByChannel("gomes.topic")
		busA.SendAsync(context.Background(), CreateCommand(fmt.Sprintf("message %d", i), "123"))
	}
}
