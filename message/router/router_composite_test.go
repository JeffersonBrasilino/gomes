package router

import (
	"context"
	"errors"
	"testing"

	"github.com/hex-api-go/pkg/core/infrastructure/gomes/message"
)

type dummyHandler struct {
	returnNil   bool
	shouldError bool
}

func (d *dummyHandler) Handle(_ context.Context, msg *message.Message) (*message.Message, error) {
	if d.shouldError {
		return nil, errors.New("erro de handler")
	}
	if d.returnNil {
		return nil, nil
	}
	return msg, nil
}

func TestNewRouter(t *testing.T) {
	t.Parallel()
	r := NewRouter()
	if r == nil {
		t.Error("NewRouter deve retornar instância não nula")
	}
}

func TestAddHandler(t *testing.T) {
	t.Parallel()
	r := NewRouter()
	h := &dummyHandler{}
	r.AddHandler(h)
	if len(r.handlers) != 1 {
		t.Error("AddHandler deve adicionar handler ao slice")
	}
	if r.handlers[0] != h {
		t.Error("Handler adicionado não corresponde ao esperado")
	}
}

func TestHandle_AllSuccess(t *testing.T) {
	t.Parallel()
	r := NewRouter()
	h1 := &dummyHandler{}
	h2 := &dummyHandler{}
	r.AddHandler(h1).AddHandler(h2)
	msg := &message.Message{}
	result, err := r.Handle(context.Background(), msg)
	if err != nil {
		t.Errorf("Handle deve retornar erro nil, retornou: %v", err)
	}
	if result != msg {
		t.Error("Handle deve retornar a mensagem original se todos os handlers passarem")
	}
}

func TestHandle_HandlerReturnsNil(t *testing.T) {
	t.Parallel()
	r := NewRouter()
	h1 := &dummyHandler{returnNil: true}
	h2 := &dummyHandler{}
	r.AddHandler(h1).AddHandler(h2)
	msg := &message.Message{}
	result, err := r.Handle(context.Background(), msg)
	if err != nil {
		t.Errorf("Handle deve retornar erro nil quando handler retorna nil")
	}
	if result != nil {
		t.Error("Handle deve retornar nil se algum handler retornar nil")
	}
}

func TestHandle_HandlerReturnsError(t *testing.T) {
	t.Parallel()
	r := NewRouter()
	h1 := &dummyHandler{shouldError: true}
	h2 := &dummyHandler{}
	r.AddHandler(h1).AddHandler(h2)
	msg := &message.Message{}
	result, err := r.Handle(context.Background(), msg)
	if err == nil {
		t.Error("Handle deve retornar erro se algum handler retornar erro")
	}
	if result != nil {
		t.Error("Handle deve retornar nil se algum handler retornar erro")
	}
}
