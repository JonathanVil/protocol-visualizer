package simulator

import (
	"errors"
)

var (
	ErrMessageNotFound         = errors.New("message not found")
	ErrMessageAlreadyDelivered = errors.New("message already delivered")
	ErrActorNotFound           = errors.New("actor not found")
	ErrFieldNotFound           = errors.New("field not found")
	ErrMethodNotFound          = errors.New("method not found")
	ErrInvalidArgument         = errors.New("invalid argument")
)

func GetCode(err error) string {
	switch {
	case errors.Is(err, ErrMessageNotFound):
		return "MESSAGE_NOT_FOUND"
	case errors.Is(err, ErrMessageAlreadyDelivered):
		return "MESSAGE_ALREADY_DELIVERED"
	case errors.Is(err, ErrActorNotFound):
		return "ACTOR_NOT_FOUND"
	case errors.Is(err, ErrFieldNotFound):
		return "FIELD_NOT_FOUND"
	case errors.Is(err, ErrMethodNotFound):
		return "METHOD_NOT_FOUND"
	case errors.Is(err, ErrInvalidArgument):
		return "INVALID_ARGUMENT"
	default:
		return "UNKNOWN"
	}
}
