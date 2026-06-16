package simulator

import (
	"fmt"
	"reflect"
	"strings"
)

type ActorSnapshot struct {
	id       int
	typeName string
	fields   map[string]any
}

type MessageSnapshot struct {
	Id            string
	From          int
	To            int
	Payload       any
	DeliverAtTick int
}

type Snapshot struct {
	Tick     int
	Running  bool
	Actors   []ActorSnapshot
	Messages []MessageSnapshot
}

func (s *Simulator) GetSnapshot() Snapshot {
	actors := make([]ActorSnapshot, len(s.actors))

	// create a snapshot of all actors
	for _, a := range s.actors {
		fields, err := structToMap(a)
		if err != nil {
			panic(err)
		}

		actors = append(actors, ActorSnapshot{
			id:       a.ID(),
			typeName: s.actorTypeNames[a.ID()],
			fields:   fields,
		})
	}

	// create a snapshot of all messages
	messages := make([]MessageSnapshot, len(s.history))
	for tick, queue := range s.tickQueues {
		for _, m := range queue {
			messages = append(messages, MessageSnapshot{
				Id:            m.ID,
				From:          m.From,
				To:            m.To,
				Payload:       m.Payload,
				DeliverAtTick: tick,
			})
		}
	}

	return Snapshot{
		Tick:     s.tick,
		Running:  s.running.Load(),
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
