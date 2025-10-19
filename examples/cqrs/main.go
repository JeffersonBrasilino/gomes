package main

import (
	"context"
	"fmt"
	"log/slog"
	"os/signal"
	"syscall"
	"time"

	"github.com/jeffersonbrasilino/gomes"
)

// command
type Action struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func CreateAction(Username, Password string) *Action {
	return &Action{
		Username,
		Password,
	}
}

// this function is responsible for the record name and action routing
func (c *Action) Name() string {
	return "createUser"
}

// CQRS acton handler
type ActionHandler struct{}

// response structure
type ResultCm struct {
	Result any
}

func NewComandHandler() *ActionHandler {
	return &ActionHandler{}
}

// note that the link between the action and its handler is the type of the data parameter.
// This indicates that this handler is responsible for this action
func (c *ActionHandler) Handle(ctx context.Context, data *Action) (*ResultCm, error) {
	fmt.Println("process action ok", data.Username)
	time.Sleep(time.Second * 2)
	return &ResultCm{"deu tudo certo"}, nil
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	slog.Info("start message system CQRS....")

	gomes.AddActionHandler(NewComandHandler())

	gomes.Start()

	maxPublishMessages := 5
	for i := 1; i <= maxPublishMessages; i++ {
		fmt.Println("publish command message...")

		busA := gomes.CommandBus()
		busA.Send(context.Background(), CreateAction("COMMAND", "123"))

		busB := gomes.QueryBus()
		busB.Send(context.Background(), CreateAction("QUERY", "123"))

		time.Sleep(time.Second * 3)
	}

	<-ctx.Done()
	gomes.Shutdown()
}
