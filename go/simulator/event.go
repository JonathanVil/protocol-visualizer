package simulator

type EventType string

const (
	EventMessageSent        EventType = "message.sent"
	EventMessageDelivered   EventType = "message.delivered"
	EventMessageDropped     EventType = "message.dropped"
	EventMessageDelayed     EventType = "message.delayed"
	EventActorSpawned       EventType = "actor.spawned"
	EventActorFieldChanged  EventType = "actor.fieldChanged"
	EventActorKilled        EventType = "actor.killed"
	EventActorRevived       EventType = "actor.revived"
	EventSimSettingsChanged EventType = "sim.settingsChanged"
	EventSnapshot           EventType = "snapshot"
)

type Event struct {
	Type    EventType
	Tick    int
	Payload any
}

// Typed payload structs (json tags are metadata only — sim does not import encoding/json).

type MessageSentPayload struct {
	MessageID     string `json:"messageId"`
	From          int    `json:"from"`
	To            int    `json:"to"`
	Payload       any    `json:"payload"`
	SentTick      int    `json:"sentTick"`
	DeliverAtTick int    `json:"deliverAtTick"`
}

type MessageDeliveredPayload struct {
	MessageID string `json:"messageId"`
	From      int    `json:"from"`
	To        int    `json:"to"`
	Payload   any    `json:"payload"`
}

type MessageDroppedPayload struct {
	MessageID string `json:"messageId"`
	Reason    string `json:"reason,omitempty"` // set when the simulator, not a user, dropped it
}

type MessageDelayedPayload struct {
	MessageID      string `json:"messageId"`
	NewDeliverTick int    `json:"newDeliverTick"`
}

type ActorSpawnedPayload struct {
	ActorID  int    `json:"actorId"`
	TypeName string `json:"typeName"`
}

type ActorKilledPayload struct {
	ActorID int `json:"actorId"`
}

type ActorRevivedPayload struct {
	ActorID int `json:"actorId"`
}

type ActorFieldChangedPayload struct {
	ActorID int    `json:"actorId"`
	Field   string `json:"field"`
	Value   any    `json:"value"`
}

type SimSettingsChangedPayload struct {
	TickDurationMs int `json:"tickDurationMs"`
	TransitTicks   int `json:"transitTicks"`
}
