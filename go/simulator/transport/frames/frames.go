package frames

import "encoding/json"

// --- Inbound (frontend → backend) ---

// Envelope is the outer wrapper for every command frame. Payload is decoded
// lazily so the router can dispatch before knowing the concrete type.
type Envelope struct {
	ID      string          `json:"id"`
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

// --- Outbound (backend → frontend) ---

// AckFrame confirms a command succeeded.
type AckFrame struct {
	Type    string `json:"type"`    // always "ack"
	ReplyTo string `json:"replyTo"` // mirrors command id
	Result  any    `json:"result"`  // null or a return value
}

// ErrorFrame signals a command was rejected or failed.
type ErrorFrame struct {
	Type    string `json:"type"`    // always "error"
	ReplyTo string `json:"replyTo"` // mirrors command id
	Code    string `json:"code"`    // error code
	Message string `json:"message"` // human-readable error message
}

// EventFrame carries a single simulation event.
type EventFrame struct {
	Seq     int64  `json:"seq"`
	Tick    int    `json:"tick"`
	Type    string `json:"type"`
	Payload any    `json:"payload"`
}

// SnapshotFrame is sent once on connect before any events.
type SnapshotFrame struct {
	Type    string `json:"type"` // always "snapshot"
	Seq     int64  `json:"seq"`
	Tick    int    `json:"tick"`
	Payload any    `json:"payload"`
}

// --- Command payload structs ---

type MessageDropPayload struct {
	MessageID string `json:"messageId"`
}

type MessageDelayPayload struct {
	MessageID string `json:"messageId"`
	Ticks     int    `json:"ticks"`
}

type MessageDeliverNowPayload struct {
	MessageID string `json:"messageId"`
}

type ActorSetFieldPayload struct {
	ActorID int    `json:"actorId"`
	Field   string `json:"field"`
	Value   any    `json:"value"`
}

type ActorInvokePayload struct {
	ActorID int    `json:"actorId"`
	Method  string `json:"method"`
	Args    any    `json:"args"`
}

type SimSetSpeedPayload struct {
	Multiplier float64 `json:"multiplier"`
}

type SimSetTransitTimePayload struct {
	Ticks int `json:"ticks"`
}

type SpawnPayload struct {
	Name string `json:"name"`
}
