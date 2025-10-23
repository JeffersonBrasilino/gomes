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
	"github.com/jeffersonbrasilino/gomes/message/channel/adapter"
	"github.com/jeffersonbrasilino/gomes/message/endpoint"
	"github.com/jeffersonbrasilino/gomes/message/handler"
)

// Default channel names for the message system.
var (
	defaultCommandChannelName = "default.channel.command"
	defaultQueryChannelName   = "default.channel.query"
)

// Global containers for managing system components.
var (
	outboundChannelBuilders = container.NewGenericContainer[string, BuildableComponent[endpoint.OutboundChannelAdapter]]()
	inboundChannelBuilders  = container.NewGenericContainer[string, BuildableComponent[endpoint.InboundChannelAdapter]]()
	channelConnections      = container.NewGenericContainer[string, adapter.ChannelConnection]()
	gomesContainer          = container.NewGenericContainer[any, any]()
	activeEndpoints         = container.NewGenericContainer[string, any]()
	actionHandlers          = container.NewGenericContainer[string, BuildableComponent[message.PublisherChannel]]()
)

// BuildableComponent defines the contract for components that can be built
// from a dependency container.
type BuildableComponent[T any] interface {
	Build(container container.Container[any, any]) (T, error)
	ReferenceName() string
}

// AddPublisherChannel registers a publisher channel builder with the message system.
// Panics if a channel with the same reference name already exists.
//
// Parameters:
//   - publisher: the publisher channel builder to register
func AddPublisherChannel(publisher BuildableComponent[endpoint.OutboundChannelAdapter]) {
	if outboundChannelBuilders.Has(publisher.ReferenceName()) {
		panic(
			fmt.Sprintf(
				"[publisher-channel] channel %s already exists",
				publisher.ReferenceName(),
			),
		)
	}
	outboundChannelBuilders.Set(publisher.ReferenceName(), publisher)
}

// buildOutboundChannels builds all registered outbound channels and adds them
// to the message system container.
//
// Parameters:
//   - container: the dependency container to add built channels to
func buildOutboundChannels(container container.Container[any, any]) {
	for _, v := range outboundChannelBuilders.GetAll() {
		outboundChannel, err := v.Build(container)
		if err != nil {
			panic(
				fmt.Sprintf(
					"[publisher-channel] %s",
					err,
				),
			)
		}
		container.Set(v.ReferenceName(), outboundChannel)
	}
}

// registerDefaultEndpoints registers the default command and query endpoints
// with the message system.
//
// Parameters:
//   - container: the dependency container to register endpoints with
func registerDefaultEndpoints(container container.Container[any, any]) {

	commandDispatcher, _ := endpoint.NewMessageDispatcherBuilder(
		defaultCommandChannelName,
		"",
	).Build(container)

	activeEndpoints.Set(defaultCommandChannelName, bus.NewCommandBus(commandDispatcher))

	queryDispatcher, _ := endpoint.NewMessageDispatcherBuilder(
		defaultQueryChannelName,
		"",
	).Build(container)
	activeEndpoints.Set(defaultQueryChannelName, bus.NewQueryBus(queryDispatcher))
}

// AddChannelConnection registers a channel connection with the message system.
// Panics if a connection with the same reference name already exists.
//
// Parameters:
//   - con: the channel connection to register
func AddChannelConnection(con adapter.ChannelConnection) {
	if channelConnections.Has(con.ReferenceName()) {
		panic(
			fmt.Sprintf(
				"[channel-module] connection %s already exists",
				con.ReferenceName(),
			),
		)
	}
	channelConnections.Set(con.ReferenceName(), con)
}

// buildChannelConnections builds all registered channel connections and adds them
// to the message system container.
//
// Parameters:
//   - container: the dependency container to add built connections to
func buildChannelConnections(container container.Container[any, any]) {
	for _, v := range channelConnections.GetAll() {
		err := v.Connect()
		if err != nil {
			panic(
				fmt.Sprintf(
					"[channel-module] %s",
					err,
				),
			)
		}
		container.Set(v.ReferenceName(), v)
	}
}

// AddConsumerChannel registers a consumer channel builder with the message system.
// Panics if a consumer with the same reference name already exists.
//
// Parameters:
//   - inboundChannel: the consumer channel builder to register
func AddConsumerChannel(inboundChannel BuildableComponent[endpoint.InboundChannelAdapter]) {
	if inboundChannelBuilders.Has(inboundChannel.ReferenceName()) {
		panic(
			fmt.Sprintf(
				"[consumer-channel] consumer for channel %s already exists",
				inboundChannel.ReferenceName(),
			),
		)
	}
	inboundChannelBuilders.Set(inboundChannel.ReferenceName(), inboundChannel)
}

// buildInboundChannels builds all registered inbound channels and adds them
// to the message system container.
//
// Parameters:
//   - container: the dependency container to add built channels to
func buildInboundChannels(container container.Container[any, any]) {
	for _, v := range inboundChannelBuilders.GetAll() {
		inboundChannel, err := v.Build(container)
		if err != nil {
			panic(fmt.Sprintf("[consumer-channel] %s", err))
		}
		container.Set(inboundChannel.ReferenceName(), inboundChannel)
	}
}

// AddActionHandler registers an action handler with the message system.
// Panics if a handler for the same action already exists.
//
// Parameters:
//   - handlerAction: the action handler to register
func AddActionHandler[T handler.Action, U any](handlerAction handler.ActionHandler[T, U]) {
	action := *new(T)
	if outboundChannelBuilders.Has(action.Name()) {
		panic(
			fmt.Sprintf(
				"handler for %s already exists",
				action.Name(),
			),
		)
	}

	actionHandlers.Set(
		action.Name(),
		handler.NewActionHandleActivatorBuilder(
			action.Name(),
			handlerAction,
		),
	)
}

func buildActionHandlers(container container.Container[any, any]) {
	for _, v := range actionHandlers.GetAll() {
		actionHandler, err := v.Build(container)
		if err != nil {
			return
		}
		container.Set(actionHandler.Name(), actionHandler)
	}
}

// Start initializes the message system by building all registered components
// and registering default endpoints.
func Start() {
	registerDefaultEndpoints(gomesContainer)
	buildActionHandlers(gomesContainer)
	buildChannelConnections(gomesContainer)
	buildOutboundChannels(gomesContainer)
	buildInboundChannels(gomesContainer)

	fmt.Println("ALL COMPONENTS", gomesContainer.GetAll())
}

// CommandBus returns the default command bus instance.
//
// Returns:
//   - *bus.CommandBus: the default command bus
func CommandBus() *bus.CommandBus {
	return CommandBusByChannel(defaultCommandChannelName)
}

// QueryBus returns the default query bus instance.
//
// Returns:
//   - *bus.QueryBus: the default query bus
func QueryBus() *bus.QueryBus {
	return QueryBusByChannel(defaultQueryChannelName)
}

// CommandBusByChannel returns or creates a command bus for the specified channel.
//
// Parameters:
//   - channelName: the name of the channel for the command bus
//
// Returns:
//   - *bus.CommandBus: the command bus for the specified channel
func CommandBusByChannel(channelName string) *bus.CommandBus {
	dispatcher, err := activeEndpoints.Get(channelName)
	if err != nil {
		dispatcher, err := endpoint.NewMessageDispatcherBuilder(
			channelName,
			channelName,
		).Build(gomesContainer)
		if err != nil {
			panic(err)
		}

		commandBus := bus.NewCommandBus(dispatcher)
		activeEndpoints.Set(channelName, commandBus)
		return commandBus
	}

	commandDispatcher, ok := dispatcher.(*bus.CommandBus)
	if !ok {
		panic(fmt.Sprintf("channel %s is not command channel", channelName))
	}
	return commandDispatcher
}

// QueryBusByChannel returns or creates a query bus for the specified channel.
//
// Parameters:
//   - channelName: the name of the channel for the query bus
//
// Returns:
//   - *bus.QueryBus: the query bus for the specified channel
func QueryBusByChannel(channelName string) *bus.QueryBus {
	dispatcher, err := activeEndpoints.Get(channelName)
	if err != nil {
		dispatcher, err := endpoint.NewMessageDispatcherBuilder(
			channelName,
			channelName,
		).Build(gomesContainer)
		if err != nil {
			panic(err)
		}

		queryBus := bus.NewQueryBus(dispatcher)
		activeEndpoints.Set(channelName, queryBus)
		return queryBus
	}

	queryDispatcher, ok := dispatcher.(*bus.QueryBus)
	if !ok {
		panic(fmt.Sprintf("channel %s is not query channel", channelName))
	}
	return queryDispatcher
}

// EventBusByChannel returns or creates an event bus for the specified channel.
//
// Parameters:
//   - channelName: the name of the channel for the event bus
//
// Returns:
//   - *bus.EventBus: the event bus for the specified channel
func EventBusByChannel(channelName string) *bus.EventBus {
	dispatcher, err := activeEndpoints.Get(channelName)
	if err != nil {
		dispatcher, err := endpoint.NewMessageDispatcherBuilder(
			channelName,
			channelName,
		).Build(gomesContainer)
		if err != nil {
			panic(err)
		}

		eventBus := bus.NewEventBus(dispatcher)
		activeEndpoints.Set(channelName, eventBus)
		return eventBus
	}

	eventDispatcher, ok := dispatcher.(*bus.EventBus)
	if !ok {
		panic(fmt.Sprintf("channel %s is not publish event channel", channelName))
	}
	return eventDispatcher
}

// EventDrivenConsumer creates and returns an event-driven consumer for the
// specified consumer name.
//
// Parameters:
//   - consumerName: the name of the consumer to create
//
// Returns:
//   - *endpoint.EventDrivenConsumer: the created event-driven consumer
//   - error: error if consumer creation fails
func EventDrivenConsumer(consumerName string) (*endpoint.EventDrivenConsumer, error) {

	consumerActive, _ := activeEndpoints.Get(consumerName)
	if consumerActive != nil {
		return nil, fmt.Errorf("consumer for %s already exists", consumerName)
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
// consumers and closing all channels.
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
		}
	}
	slog.Info("[message-system] shutdown completed")
}

// ShowActiveEndpoints displays all currently active endpoints in the message system.
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
