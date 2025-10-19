package main

import (
	"context"
	"fmt"
	"log/slog"
	"os/signal"
	"syscall"
	"time"

	"github.com/jeffersonbrasilino/gomes"
	kafka "github.com/jeffersonbrasilino/gomes/channel/kafka"
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
type CommandHandler struct{}

// response structure
type ResultCm struct {
	Result any
}

func NewComandHandler() *CommandHandler {
	return &CommandHandler{}
}

// note that the link between the action and its handler is the type of the data parameter.
// This indicates that this handler is responsible for this action
func (c *CommandHandler) Handle(ctx context.Context, data *Command) (*ResultCm, error) {
	fmt.Println("process command ok")
	time.Sleep(time.Second * 2)
	return &ResultCm{"deu tudo certo"}, nil
}

func main() {

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	slog.Info("start message system consumer....")

	//create kafka connection
	//The connection can, and is even recommended,
	//be registered only once after registration.
	//To use it in your channels, simply use its name in the channel reference name.
	gomes.AddChannelConnection(
		kafka.NewConnection("defaultConKafka", []string{"kafka:9092"}),
	)

	//create DLQ publisher channel
	publisherDlqChannel := kafka.NewPublisherChannelAdapterBuilder(
		"defaultConKafka",
		"gomes.dlq",
	)

	//create publisher channel
	publisherChannel := kafka.NewPublisherChannelAdapterBuilder(
		"defaultConKafka",
		"gomes.topic",
	)

	//create consumer channel on message system
	//For the consumer channel,
	//there are two resilience approaches: the retry pattern and the dead-letter pattern.
	//You can use both together or opt for just one of the options.
	topicConsumerChannel := kafka.NewConsumerChannelAdapterBuilder(
		"defaultConKafka",
		"gomes.topic",
		"test_consumer",
	)
	topicConsumerChannel.WithRetryTimes(2_000, 3_000)
	topicConsumerChannel.WithDeadLetterChannelName("gomes.dlq")

	//register publisher channel on message system
	gomes.AddPublisherChannel(publisherChannel)

	//register dlq channel on message system
	gomes.AddPublisherChannel(publisherDlqChannel)

	//register consumer channel on message system
	gomes.AddConsumerChannel(topicConsumerChannel)

	// Register CQRS action and action handler.
	gomes.AddActionHandler(NewComandHandler())

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

	//Run the event-driven consumer. Note that we have a few settings:
	//- WithMessageProcessingTimeout: Sets the message processing timeout
	//- WithAmountOfProcessors: Sets the number of parallel processing nodes
	//- WithStopOnError: If a processing error occurs, the consumer is shut down (default is true)
	go consumer.WithAmountOfProcessors(1).
		WithMessageProcessingTimeout(50000).
		WithStopOnError(false).
		Run(ctx)

	<-ctx.Done()
	//message system graceful shutdown
	gomes.Shutdown()
}

func publishMessage() {
	maxPublishMessages := 5
	for i := 1; i <= maxPublishMessages; i++ {
		fmt.Println("publish command message...")
		//get command bus
		//the message type is defined by bus(command/query/event)
		busA := gomes.CommandBusByChannel("gomes.topic")
		busA.SendAsync(context.Background(), CreateCommand("teste", "123"))
		time.Sleep(time.Second * 3)
	}
}
