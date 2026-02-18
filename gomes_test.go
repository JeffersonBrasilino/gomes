package gomes_test

import (
	"testing"

	"github.com/jeffersonbrasilino/gomes"
	"github.com/jeffersonbrasilino/gomes/container"
	"github.com/jeffersonbrasilino/gomes/message/adapter"
	"github.com/jeffersonbrasilino/gomes/message/endpoint"
	"github.com/jeffersonbrasilino/gomes/message/handler"
)

// fakeOutboundBuilder implements gomes.BuildableComponent[endpoint.OutboundChannelAdapter]
type fakeOutboundBuilder struct{ name string }

func (f *fakeOutboundBuilder) Build(c container.Container[any, any]) (endpoint.OutboundChannelAdapter, error) {
	return nil, nil
}

func (f *fakeOutboundBuilder) ReferenceName() string { return f.name }

// fakeInboundBuilder implements gomes.BuildableComponent[*adapter.InboundChannelAdapter]
type fakeInboundBuilder struct{ name string }

func (f *fakeInboundBuilder) Build(c container.Container[any, any]) (*adapter.InboundChannelAdapter, error) {
	// return a real adapter instance with the reference name set to avoid nil deref during Start
	return adapter.NewInboundChannelAdapter(nil, f.name, "", nil, nil, nil, false), nil
}

func (f *fakeInboundBuilder) ReferenceName() string { return f.name }

type dummyConn struct{ name string }

func (d *dummyConn) ReferenceName() string { return d.name }
func (d *dummyConn) Connect() error        { return nil }
func (d *dummyConn) Disconnect() error     { return nil }

func TestAddPublisherChannel_Duplicate(t *testing.T) {
	b := &fakeOutboundBuilder{name: "pub.chan.dup"}
	if err := gomes.AddPublisherChannel(b); err != nil {
		t.Fatalf("unexpected error on first add: %v", err)
	}
	if err := gomes.AddPublisherChannel(b); err == nil {
		t.Fatal("expected error when adding duplicate publisher channel, got nil")
	}
}

func TestAddConsumerChannel_Duplicate(t *testing.T) {
	b := &fakeInboundBuilder{name: "in.chan.dup"}
	if err := gomes.AddConsumerChannel(b); err != nil {
		t.Fatalf("unexpected error on first add: %v", err)
	}
	if err := gomes.AddConsumerChannel(b); err == nil {
		t.Fatal("expected error when adding duplicate consumer channel, got nil")
	}
}

func TestAddChannelConnection_Duplicate(t *testing.T) {
	c := &dummyConn{name: "conn.dup"}
	if err := gomes.AddChannelConnection(c); err != nil {
		t.Fatalf("unexpected error on first add: %v", err)
	}
	if err := gomes.AddChannelConnection(c); err == nil {
		t.Fatal("expected error when adding duplicate channel connection, got nil")
	}
}

func TestAddActionHandler_Nil(t *testing.T) {
	// Passing nil should return an error
	err := gomes.AddActionHandler[handler.Action, any](nil)
	if err == nil {
		t.Fatal("expected error when registering nil action handler, got nil")
	}
}

func TestEnableOtelTraceAndShowShutdown(t *testing.T) {
	// Call EnableOtelTrace, ShowActiveEndpoints and Shutdown to ensure they run without panic.
	gomes.EnableOtelTrace()
	gomes.ShowActiveEndpoints()
	gomes.Shutdown()
}

func TestStartAndBuses(t *testing.T) {
	t.Run("start and default buses", func(t *testing.T) {
		t.Parallel()
		if err := gomes.Start(); err != nil {
			t.Fatalf("Start should not return error, got: %v", err)
		}

		if _, err := gomes.CommandBus(); err != nil {
			t.Fatalf("CommandBus should be available after Start: %v", err)
		}

		if _, err := gomes.QueryBus(); err != nil {
			t.Fatalf("QueryBus should be available after Start: %v", err)
		}
	})
}

func TestBusByChannel_TypeMismatchAndEventBus(t *testing.T) {
	t.Run("query-then-command-mismatch", func(t *testing.T) {
		t.Parallel()
		// create query bus for channel
		if _, err := gomes.QueryBusByChannel("mismatch.chan1"); err != nil {
			t.Fatalf("unexpected error creating query bus: %v", err)
		}
		// now command bus by same channel should return type error
		if _, err := gomes.CommandBusByChannel("mismatch.chan1"); err == nil {
			t.Fatal("expected error when creating command bus on query channel, got nil")
		}
	})

	t.Run("command-then-query-mismatch", func(t *testing.T) {
		t.Parallel()
		if _, err := gomes.CommandBusByChannel("mismatch.chan2"); err != nil {
			t.Fatalf("unexpected error creating command bus: %v", err)
		}
		if _, err := gomes.QueryBusByChannel("mismatch.chan2"); err == nil {
			t.Fatal("expected error when creating query bus on command channel, got nil")
		}
	})

	t.Run("event-bus-creation", func(t *testing.T) {
		t.Parallel()
		if _, err := gomes.EventBusByChannel("ev.chan.test"); err != nil {
			t.Fatalf("EventBusByChannel should create event bus: %v", err)
		}
	})
}

func TestEventDrivenConsumer_AlreadyExists(t *testing.T) {
	t.Run("consumer already exists", func(t *testing.T) {
		t.Parallel()
		name := "consumer.exists"
		// create a command bus under this name so EventDrivenConsumer sees it as existing
		if _, err := gomes.CommandBusByChannel(name); err != nil {
			t.Fatalf("unexpected creating command bus: %v", err)
		}

		if _, err := gomes.EventDrivenConsumer(name); err == nil {
			t.Fatal("expected error when creating consumer for existing endpoint, got nil")
		}
	})
}
