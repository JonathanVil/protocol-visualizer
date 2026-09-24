package simulator

import (
	"fmt"
	"reflect"
	"slices"
)

type ActorSnapshot struct {
	ID       int            `json:"id"`
	TypeName string         `json:"typeName"`
	Fields   map[string]any `json:"fields"`
	Methods  []MethodInfo   `json:"methods"`
	Alive    bool           `json:"alive"`
}

type MessageSnapshot struct {
	ID            string `json:"id"`
	From          int    `json:"from"`
	To            int    `json:"to"`
	Payload       any    `json:"payload"`
	SentTick      int    `json:"sentTick"`
	DeliverAtTick int    `json:"deliverAtTick"`
}

type SettingsSnapshot struct {
	TickDurationMs int `json:"tickDurationMs"`
	TransitTicks   int `json:"transitTicks"`
}

type Snapshot struct {
	Tick     int               `json:"tick"`
	Running  bool              `json:"running"`
	Settings SettingsSnapshot  `json:"settings"`
	Actors   []ActorSnapshot   `json:"actors"`
	Messages []MessageSnapshot `json:"inTransit"`
}

func (s *Simulator) GetSnapshot() Snapshot {
	actors := make([]ActorSnapshot, 0, len(s.actors))
	for _, a := range s.actors {
		fields, err := structToMap(a)
		if err != nil {
			fields = nil
		}
		actors = append(actors, ActorSnapshot{
			ID:       a.ID(),
			TypeName: s.actorTypeNames[a.ID()],
			Fields:   fields,
			Methods:  actorMethods(a),
			Alive:    !s.dead[a.ID()],
		})
	}
	slices.SortFunc(actors, func(a, b ActorSnapshot) int { return a.ID - b.ID })

	messages := make([]MessageSnapshot, 0)
	for tick, queue := range s.tickQueues {
		for _, m := range queue {
			messages = append(messages, MessageSnapshot{
				ID:            m.ID,
				From:          m.From,
				To:            m.To,
				Payload:       m.Payload,
				SentTick:      m.SentTick,
				DeliverAtTick: tick,
			})
		}
	}

	return Snapshot{
		Tick:     s.tick,
		Running:  s.running.Load(),
		Settings: SettingsSnapshot{TickDurationMs: int(s.TickDuration.Milliseconds()), TransitTicks: s.TransitTicks},
		Actors:   actors,
		Messages: messages,
	}
}

func structToMap(v any) (map[string]any, error) {
	val := reflect.ValueOf(v)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	if val.Kind() != reflect.Struct {
		return nil, fmt.Errorf("expected struct, got %s", val.Kind())
	}

	typ := val.Type()
	m := make(map[string]any, val.NumField())

	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		if !field.IsExported() {
			continue
		}
		if name := wireName(field); name != "" {
			m[name] = val.Field(i).Interface()
		}
	}
	return m, nil
}
