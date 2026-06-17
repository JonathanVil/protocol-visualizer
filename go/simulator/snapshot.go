package simulator

import (
	"fmt"
	"reflect"
	"strings"
)

type ActorSnapshot struct {
	ID       int            `json:"id"`
	TypeName string         `json:"typeName"`
	Fields   map[string]any `json:"fields"`
}

type MessageSnapshot struct {
	Id            string
	From          int
	To            int
	Payload       any
	SentTick      int
	DeliverAtTick int
}

type SettingsSnapshot struct {
	TickDurationMs int `json:"tickDurationMs"`
	TransitTicks   int `json:"transitTicks"`
}

type Snapshot struct {
	Tick     int
	Running  bool
	Settings SettingsSnapshot
	Actors   []ActorSnapshot
	Messages []MessageSnapshot
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
		})
	}

	messages := make([]MessageSnapshot, 0)
	for tick, queue := range s.tickQueues {
		for _, m := range queue {
			messages = append(messages, MessageSnapshot{
				Id:            m.ID,
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
		name := field.Name
		if tag := field.Tag.Get("json"); tag != "" && tag != "-" {
			name = strings.Split(tag, ",")[0]
		}
		m[name] = val.Field(i).Interface()
	}
	return m, nil
}
