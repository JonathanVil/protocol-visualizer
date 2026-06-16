package simulator

type EventType string

const (
	EventMessageDelivered   EventType = "message.delivered"
	EventMessageDropped     EventType = "message.dropped"
	EventMessageDelayed     EventType = "message.delayed"
	EventActorSpawned       EventType = "actor.spawned"
	EventActorFieldChanged  EventType = "actor.fieldChanged"
	EventSimSettingsChanged EventType = "sim.settingsChanged"
)

type Event struct {
	Type    EventType
	Tick    int
	Payload any
}

// Typed payload structs (json tags are metadata only — sim does not import encoding/json).

type MessageDeliveredPayload struct {
	MessageID string `json:"messageId"`
	From      int    `json:"from"`
	To        int    `json:"to"`
	Payload   any    `json:"payload"`
}

type MessageDroppedPayload struct {
	MessageID string `json:"messageId"`
}

type MessageDelayedPayload struct {
	MessageID      string `json:"messageId"`
	NewDeliverTick int    `json:"newDeliverTick"`
}

type ActorSpawnedPayload struct {
	ActorID  int    `json:"actorId"`
	TypeName string `json:"typeName"`
}

type ActorFieldChangedPayload struct {
	ActorID int    `json:"actorId"`
	Field   string `json:"field"`
	Value   any    `json:"value"`
}

type SimSettingsChangedPayload struct {
	SpeedMultiplier float64 `json:"speedMultiplier"`
	TransitTicks    int     `json:"transitTicks"`
}
