// Package gomes provides the core message system implementation with
// comprehensive message routing, processing, and management capabilities.
//
// This package implements a complete message system that supports command/query
// separation (CQRS), event-driven processing, and various messaging patterns.
// It provides centralized management of channels, connections, and endpoints
// with support for both inbound and outbound message processing.
//
// The gomes implementation supports:
// - Command and Query bus management
// - Event-driven consumer processing
// - Channel connection management
// - Action handler registration
// - Default endpoint configuration
// - System lifecycle management
package gomes

import (
	"fmt"
	"log/slog"

	"github.com/jeffersonbrasilino/gomes/bus"
	"github.com/jeffersonbrasilino/gomes/container"
	"github.com/jeffersonbrasilino/gomes/message"
	"github.com/jeffersonbrasilino/gomes/message/adapter"
	"github.com/jeffersonbrasilino/gomes/message/endpoint"
	"github.com/jeffersonbrasilino/gomes/message/handler"
	"github.com/jeffersonbrasilino/gomes/otel"
)

// Default channel names for the message system.
const (
	defaultCommandChannelName = "default.channel.command"
	defaultQueryChannelName   = "default.channel.query"
)

// Global containers for managing message system components.
var (
	outboundChannelBuilders = container.NewGenericContainer[
		string,
		BuildableComponent[endpoint.OutboundChannelAdapter],
	]()
	inboundChannelBuilders = container.NewGenericContainer[
		string,
		BuildableComponent[*adapter.InboundChannelAdapter],
	]()
	channelConnections = container.NewGenericContainer[
		string,
		adapter.ChannelConnection,
	]()
	gomesContainer  = container.NewGenericContainer[any, any]()
	activeEndpoints = container.NewGenericContainer[string, any]()
	actionHandlers  = container.NewGenericContainer[
		string,
		BuildableComponent[message.PublisherChannel],
	]()
)

// BuildableComponent defines the contract for components that can be built
// from a dependency container.
type BuildableComponent[T any] interface {
	Build(container container.Container[any, any]) (T, error)
	ReferenceName() string
}

// AddPublisherChannel registers a publisher channel builder with the message
// system. The channel builder will be used to create outbound channel adapters
// during system initialization.
//
// Parameters:
//   - publisher: the publisher channel builder to register
//
// Returns:
//   - error: error if a channel with the same reference name already exists
func AddPublisherChannel(
	publisher BuildableComponent[endpoint.OutboundChannelAdapter],
) error {
	if outboundChannelBuilders.Has(publisher.ReferenceName()) {
		return fmt.Errorf(
			"[publisher-channel] channel %s already exists",
			publisher.ReferenceName(),
		)
	}
	outboundChannelBuilders.Set(publisher.ReferenceName(), publisher)
	return nil
}

// buildOutboundChannels builds all registered outbound channels and adds them
// to the message system container. This function is called during system
// initialization and processes all registered publisher channel builders.
//
// Parameters:
//   - container: the dependency container to add built channels to
//
// Returns:
//   - error: error if building any channel fails
func buildOutboundChannels(
	container container.Container[any, any],
) error {
	for _, v := range outboundChannelBuilders.GetAll() {
		outboundChannel, err := v.Build(container)
		if err != nil {
			return fmt.Errorf(
				"[publisher-channel] %s",
				err,
			)
		}
		container.Set(v.ReferenceName(), outboundChannel)
	}
	return nil
}

// registerDefaultEndpoints registers the default command and query endpoints
// with the message system. These endpoints are used when no specific channel
// is specified for command or query operations.
//
// Parameters:
//   - container: the dependency container to register endpoints with
//
// Returns:
//   - error: error if endpoint registration fails
func registerDefaultEndpoints(
	container container.Container[any, any],
) error {
	commandDispatcher, err := endpoint.NewMessageDispatcherBuilder(
		defaultCommandChannelName,
		"",
	).Build(container)
	if err != nil {
		return fmt.Errorf(
			"[message-dispatcher] failed to build command dispatcher: %w",
			err,
		)
	}

	err = activeEndpoints.Set(
		defaultCommandChannelName,
		bus.NewCommandBus(commandDispatcher),
	)
	if err != nil {
		return fmt.Errorf(
			"[message-dispatcher] failed to register command bus: %w",
			err,
		)
	}

	queryDispatcher, err := endpoint.NewMessageDispatcherBuilder(
		defaultQueryChannelName,
		"",
	).Build(container)
	if err != nil {
		return fmt.Errorf(
			"[message-dispatcher] failed to build query dispatcher: %w",
			err,
		)
	}

	err = activeEndpoints.Set(
		defaultQueryChannelName,
		bus.NewQueryBus(queryDispatcher),
	)
	if err != nil {
		return fmt.Errorf(
			"[message-dispatcher] failed to register query bus: %w",
			err,
		)
	}

	return nil
}

// AddChannelConnection registers a channel connection with the message system.
// The connection will be established during system initialization. Multiple
// connections can be registered, each with a unique reference name.
//
// Parameters:
//   - con: the channel connection to register
//
// Returns:
//   - error: error if a connection with the same reference name already exists
func AddChannelConnection(con adapter.ChannelConnection) error {
	if channelConnections.Has(con.ReferenceName()) {
		return fmt.Errorf(
			"[channel-module] connection %s already exists",
			con.ReferenceName(),
		)
	}
	channelConnections.Set(con.ReferenceName(), con)
	return nil
}

// buildChannelConnections builds all registered channel connections and adds
// them to the message system container. This function establishes connections
// to messaging brokers (Kafka, RabbitMQ, etc.) and is called during system
// initialization.
//
// Parameters:
//   - container: the dependency container to add built connections to
//
// Returns:
//   - error: error if any connection establishment fails
func buildChannelConnections(
	container container.Container[any, any],
) error {
	for _, v := range channelConnections.GetAll() {
		err := v.Connect()
		if err != nil {
			return fmt.Errorf(
				"[channel-module] %s",
				err,
			)
		}
		container.Set(v.ReferenceName(), v)
	}
	return nil
}

// AddConsumerChannel registers a consumer channel builder with the message
// system. The channel builder will be used to create inbound channel adapters
// for consuming messages from messaging brokers.
//
// Parameters:
//   - inboundChannel: the consumer channel builder to register
//
// Returns:
//   - error: error if a consumer with the same reference name already exists
func AddConsumerChannel(
	inboundChannel BuildableComponent[*adapter.InboundChannelAdapter],
) error {
	if inboundChannelBuilders.Has(inboundChannel.ReferenceName()) {
		return fmt.Errorf(
			"[consumer-channel] consumer for channel %s already exists",
			inboundChannel.ReferenceName(),
		)
	}
	inboundChannelBuilders.Set(inboundChannel.ReferenceName(), inboundChannel)
	return nil
}

// buildInboundChannels builds all registered inbound channels and adds them
// to the message system container. This function processes all registered
// consumer channel builders and is called during system initialization.
//
// Parameters:
//   - container: the dependency container to add built channels to
//
// Returns:
//   - error: error if building any channel fails
func buildInboundChannels(
	container container.Container[any, any],
) error {
	for _, v := range inboundChannelBuilders.GetAll() {
		inboundChannel, err := v.Build(container)
		if err != nil {
			return fmt.Errorf("[consumer-channel] %s", err)
		}
		container.Set(inboundChannel.ReferenceName(), inboundChannel)
	}
	return nil
}

// AddActionHandler registers an action handler with the message system.
// Action handlers process commands, queries, or events based on the action
// type. Each action type can have only one handler registered.
//
// Parameters:
//   - handlerAction: the action handler to register (must not be nil)
//
// Returns:
//   - error: error if handler is nil or a handler for the same action already
//     exists
func AddActionHandler[T handler.Action, U any](
	handlerAction handler.ActionHandler[T, U],
) error {
	if handlerAction == nil {
		return fmt.Errorf("handler cannot be nil")
	}

	action := *new(T)
	if outboundChannelBuilders.Has(action.Name()) {
		return fmt.Errorf(
			"handler for %s already exists",
			action.Name(),
		)
	}

	actionHandlers.Set(
		action.Name(),
		handler.NewActionHandleActivatorBuilder(
			action.Name(),
			handlerAction,
		),
	)
	return nil
}

// buildActionHandlers builds all registered action handlers and adds them to
// the message system container. This function processes all registered handlers
// and is called during system initialization.
//
// Parameters:
//   - container: the dependency container to add built handlers to
//
// Returns:
//   - error: error if building any handler fails
func buildActionHandlers(
	container container.Container[any, any],
) error {
	for _, v := range actionHandlers.GetAll() {
		actionHandler, err := v.Build(container)
		if err != nil {
			return fmt.Errorf(
				"[action-handler] failed to build handler: %w",
				err,
			)
		}
		err = container.Set(actionHandler.Name(), actionHandler)
		if err != nil {
			return fmt.Errorf(
				"[action-handler] failed to register handler: %w",
				err,
			)
		}
	}
	return nil
}

// Start initializes the message system by building all registered components
// and registering default endpoints. This function must be called after
// registering all channels, connections, and handlers, and before using any
// bus or consumer functionality.
//
// The initialization process follows this order:
// 1. Register default command and query endpoints
// 2. Build action handlers
// 3. Build channel connections
// 4. Build outbound channels
// 5. Build inbound channels
//
// Returns:
//   - error: error if any component fails to build or initialize
func Start() error {
	buildFunctions := []func(container container.Container[any, any]) error{
		registerDefaultEndpoints,
		buildActionHandlers,
		buildChannelConnections,
		buildOutboundChannels,
		buildInboundChannels,
	}

	for _, buildFunc := range buildFunctions {
		err := buildFunc(gomesContainer)
		if err != nil {
			return err
		}
	}

	return nil
}

// CommandBus returns the default command bus instance. The default command bus
// uses an internal channel and does not require external messaging infrastructure.
// This is useful for local command processing without message brokers.
//
// Returns:
//   - *bus.CommandBus: the default command bus (never nil, but may panic if
//     system is not initialized)
func CommandBus() (*bus.CommandBus, error) {
	cb, err := CommandBusByChannel(defaultCommandChannelName)
	if err != nil {
		// This should not happen if Start() was called correctly
		return nil, fmt.Errorf(
			"[gomes] failed to get default command bus: %v",
			err,
		)
	}
	return cb, nil
}

// QueryBus returns the default query bus instance. The default query bus uses
// an internal channel and does not require external messaging infrastructure.
// This is useful for local query processing without message brokers.
//
// Returns:
//   - *bus.QueryBus: the default query bus (never nil, but may panic if system
//     is not initialized)
func QueryBus() (*bus.QueryBus, error) {
	qb, err := QueryBusByChannel(defaultQueryChannelName)
	if err != nil {
		// This should not happen if Start() was called correctly
		return nil, fmt.Errorf(
			"[gomes] failed to get default query bus: %v",
			err,
		)
	}
	return qb, nil
}

// CommandBusByChannel returns or creates a command bus for the specified
// channel. If a bus already exists for the channel, it is returned. Otherwise,
// a new bus is created and registered. The channel must have a corresponding
// publisher channel registered.
//
// Parameters:
//   - channelName: the name of the channel for the command bus
//
// Returns:
//   - *bus.CommandBus: the command bus for the specified channel
//   - error: error if channel does not exist or is not a command channel
func CommandBusByChannel(
	channelName string,
) (*bus.CommandBus, error) {
	dispatcher, err := activeEndpoints.Get(channelName)
	if err != nil {
		dispatcher, err := endpoint.NewMessageDispatcherBuilder(
			channelName,
			channelName,
		).Build(gomesContainer)
		if err != nil {
			return nil, err
		}

		commandBus := bus.NewCommandBus(dispatcher)
		activeEndpoints.Set(channelName, commandBus)
		return commandBus, nil
	}

	commandDispatcher, ok := dispatcher.(*bus.CommandBus)
	if !ok {
		return nil, fmt.Errorf("channel %s is not command channel", channelName)
	}
	return commandDispatcher, nil
}

// QueryBusByChannel returns or creates a query bus for the specified channel.
// If a bus already exists for the channel, it is returned. Otherwise, a new
// bus is created and registered. The channel must have a corresponding
// publisher channel registered.
//
// Parameters:
//   - channelName: the name of the channel for the query bus
//
// Returns:
//   - *bus.QueryBus: the query bus for the specified channel
//   - error: error if channel does not exist or is not a query channel
func QueryBusByChannel(
	channelName string,
) (*bus.QueryBus, error) {
	dispatcher, err := activeEndpoints.Get(channelName)
	if err != nil {
		dispatcher, err := endpoint.NewMessageDispatcherBuilder(
			channelName,
			channelName,
		).Build(gomesContainer)
		if err != nil {
			return nil, err
		}

		queryBus := bus.NewQueryBus(dispatcher)
		activeEndpoints.Set(channelName, queryBus)
		return queryBus, nil
	}

	queryDispatcher, ok := dispatcher.(*bus.QueryBus)
	if !ok {
		return nil, fmt.Errorf("channel %s is not query channel", channelName)
	}
	return queryDispatcher, nil
}

// EventBusByChannel returns or creates an event bus for the specified channel.
// If a bus already exists for the channel, it is returned. Otherwise, a new
// bus is created and registered. The channel must have a corresponding
// publisher channel registered.
//
// Parameters:
//   - channelName: the name of the channel for the event bus
//
// Returns:
//   - *bus.EventBus: the event bus for the specified channel
//   - error: error if channel does not exist or is not an event channel
func EventBusByChannel(
	channelName string,
) (*bus.EventBus, error) {
	dispatcher, err := activeEndpoints.Get(channelName)
	if err != nil {
		dispatcher, err := endpoint.NewMessageDispatcherBuilder(
			channelName,
			channelName,
		).Build(gomesContainer)
		if err != nil {
			return nil, err
		}

		eventBus := bus.NewEventBus(dispatcher)
		activeEndpoints.Set(channelName, eventBus)
		return eventBus, nil
	}

	eventDispatcher, ok := dispatcher.(*bus.EventBus)
	if !ok {
		return nil, fmt.Errorf("channel %s is not publish event channel", channelName)
	}
	return eventDispatcher, nil
}

// EventDrivenConsumer creates and returns an event-driven consumer for the
// specified consumer name. The consumer must have a corresponding inbound
// channel adapter registered. The consumer processes messages asynchronously
// and supports multiple concurrent processors.
//
// Parameters:
//   - consumerName: the name of the consumer to create (must match a registered
//     inbound channel adapter reference name)
//
// Returns:
//   - *endpoint.EventDrivenConsumer: the created event-driven consumer
//   - error: error if consumer already exists or creation fails
func EventDrivenConsumer(
	consumerName string,
) (*endpoint.EventDrivenConsumer, error) {
	consumerActive, err := activeEndpoints.Get(consumerName)
	if err == nil && consumerActive != nil {
		return nil, fmt.Errorf(
			"consumer for %s already exists",
			consumerName,
		)
	}

	consumer, err := endpoint.
		NewEventDrivenConsumerBuilder(consumerName).
		Build(gomesContainer)

	if err != nil {
		return nil, err
	}

	activeEndpoints.Set(consumerName, consumer)

	return consumer, nil
}

// Shutdown gracefully shuts down the message system by stopping all active
// consumers and closing all channels. This function should be called during
// application shutdown to ensure proper cleanup of resources. All consumers
// are stopped first, followed by closing of all channels.
func Shutdown() {
	slog.Info("[message-system] shutting down...")
	for k, v := range activeEndpoints.GetAll() {
		if inboundChannel, ok := v.(*endpoint.EventDrivenConsumer); ok {
			slog.Info("[message-system] stop consumer", "name", k)
			inboundChannel.Stop()
		}
	}

	for k, v := range gomesContainer.GetAll() {
		switch c := v.(type) {
		case message.ConsumerChannel:
			slog.Info("[message-system] close consumer channel", "name", c.Name())
			c.Close()
		case message.SubscriberChannel:
			slog.Info("[message-system] unsubscribe channel", "name", c.Name())
			c.Unsubscribe()
		case endpoint.OutboundChannelAdapter:
			slog.Info("[message-system] close outbound channel", "name", k)
			c.Close()
		case adapter.ChannelConnection:
			slog.Info("[message-system] disconnect channel connection", "name", k)
			c.Disconnect()
		}
	}
	slog.Info("[message-system] shutdown completed")
}

// ShowActiveEndpoints displays all currently active endpoints in the message
// system. This function is useful for debugging and monitoring purposes,
// showing all registered endpoints and their types.
func ShowActiveEndpoints() {
	fmt.Println("\n---[Message System] Active Endpoints ---")
	fmt.Printf("%-30s | %-10s\n", "Endpoint Name", "Type")
	fmt.Println("-------------------------------------------")
	for name, ep := range activeEndpoints.GetAll() {
		endpointType := "undefined"
		switch ep.(type) {
		case *endpoint.EventDrivenConsumer:
			endpointType = "[inbound] Event-Driven"
		case *bus.CommandBus:
			endpointType = "[outbound] Command-Bus"
		case *bus.QueryBus:
			endpointType = "[outbound] Query-Bus"
		case *bus.EventBus:
			endpointType = "[outbound] Event-Bus"
		}
		fmt.Printf("%-30s | %-10s\n", name, endpointType)
	}
	fmt.Println("-------------------------------------------")
}

// EnableOtelTrace enables OpenTelemetry distributed tracing for the message
// system. This function must be called before Start() if observability is
// desired. It requires that an OpenTelemetry TracerProvider has been
// configured globally.
func EnableOtelTrace() {
	otel.EnableTrace()
}
