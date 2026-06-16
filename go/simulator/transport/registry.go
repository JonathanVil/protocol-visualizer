package transport

import (
	"encoding/json"
	"errors"
	"fmt"
)

// ErrBadPayload is returned when a command payload cannot be decoded.
var ErrBadPayload = errors.New("bad payload")

// Handler executes a decoded command and returns an optional result value.
type Handler func(payload json.RawMessage) (any, error)

// Register associates msgType in reg with a handler that receives a typed payload T.
// If the raw payload cannot be decoded into T, the handler returns ErrBadPayload.
// reg is a per-Server map; this is a free function because Go does not support
// generic methods.
func Register[T any](reg map[string]Handler, msgType string, fn func(cmd T) (any, error)) {
	reg[msgType] = func(raw json.RawMessage) (any, error) {
		var cmd T
		// Skip unmarshal when payload is absent or null; use the zero value of T.
		// This lets payload-less commands (start, stop, requestTypes) omit the field.
		if len(raw) > 0 && string(raw) != "null" {
			if err := json.Unmarshal(raw, &cmd); err != nil {
				return nil, fmt.Errorf("%w: %s: %v", ErrBadPayload, msgType, err)
			}
		}
		return fn(cmd)
	}
}
