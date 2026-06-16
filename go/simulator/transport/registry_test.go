package transport

import (
	"encoding/json"
	"errors"
	"testing"
)

func TestRegister_TypedDecode(t *testing.T) {
	type Payload struct {
		Value string `json:"value"`
	}
	reg := make(map[string]Handler)
	Register(reg, "test.typed", func(p Payload) (any, error) {
		return p.Value, nil
	})

	h, ok := reg["test.typed"]
	if !ok {
		t.Fatal("handler not registered")
	}
	result, err := h(json.RawMessage(`{"value":"hello"}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "hello" {
		t.Fatalf("expected hello, got %v", result)
	}
}

func TestRegister_BadPayload(t *testing.T) {
	type Payload struct {
		Count int `json:"count"`
	}
	reg := make(map[string]Handler)
	Register(reg, "test.badpayload", func(p Payload) (any, error) {
		return nil, nil
	})

	h := reg["test.badpayload"]
	_, err := h(json.RawMessage(`{not valid json}`))
	if err == nil {
		t.Fatal("expected error for bad payload")
	}
	if !errors.Is(err, ErrBadPayload) {
		t.Fatalf("expected ErrBadPayload, got: %v", err)
	}
}

func TestRegistry_UnknownCommand(t *testing.T) {
	reg := make(map[string]Handler)
	_, ok := reg["no.such.command"]
	if ok {
		t.Fatal("expected miss for unknown command")
	}
}
