package simulator

import (
	"fmt"
	"log"
	"math/rand/v2"
	"reflect"
	"sync"
	"time"
)

type Simulator struct {
	mu           sync.Mutex
	actors       map[int]Actor
	tickQueues   map[int][]Message
	tick         int
	tickDuration time.Duration
	running      bool
	events       chan Event
	history      []Message
	actorTypes   map[string]reflect.Type
	TransitTicks int
}

func New() *Simulator {
	return &Simulator{
		actors:       make(map[int]Actor),
		tickQueues:   make(map[int][]Message),
		tick:         0,
		tickDuration: time.Second,
		events:       make(chan Event, 16),
		actorTypes:   make(map[string]reflect.Type),
		TransitTicks: 1,
	}
}

func (s *Simulator) Register(a Actor) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.actors[a.ID()] = a
	fmt.Printf("Registered actor %d\n", a.ID())
}

func (s *Simulator) RegisterActorType(typ reflect.Type, name string) string {
	if name == "" {
		name = typ.Name()
	}

	// check if typ implements actor interface
	actorInterface := reflect.TypeOf((*Actor)(nil)).Elem()
	if !typ.Implements(actorInterface) {
		log.Fatalf("Type %s does not implement the actor interface", typ)
	}

	s.actorTypes[name] = typ

	fmt.Printf("Registered actor type %s\n", name)
	return name
}

func (s *Simulator) SpawnActor(typ string, id int) Actor {
	actualType := s.actorTypes[typ]
	v := reflect.New(actualType.Elem())
	actor := v.Interface().(Actor)
	actor.Init(id, s)

	fmt.Printf("Spawned actor %d\n", id)
	s.Register(actor)
	return actor
}

func (s *Simulator) Send(from, to int, payload any) {
	s.mu.Lock()
	defer s.mu.Unlock()

	msg := Message{From: from, To: to, Payload: payload}
	idx := s.tick + s.TransitTicks
	i := rand.IntN(len(s.tickQueues[idx]) + 1)
	s.tickQueues[idx] = append(s.tickQueues[idx][:i], append([]Message{msg}, s.tickQueues[idx][i:]...)...)
}

func (s *Simulator) Tick() {
	s.mu.Lock()
	defer s.mu.Unlock()
	defer func() {
		s.tick++
	}()

	fmt.Printf("-- Tick %d --\n", s.tick)

	// get queue of messages due this tick
	queue := s.tickQueues[s.tick]
	if queue == nil {
		return
	}

	for _, msg := range queue {
		// deliver it
		actor := s.actors[msg.To]
		if actor == nil {
			fmt.Printf("No actor found %d\n", msg.To)
			return
		}
		actor.OnMessage(msg)
		s.events <- Event{Tick: s.tick, Message: msg}
		s.history = append(s.history, msg)
	}
}

func (s *Simulator) Start() {
	s.running = true
	fmt.Println("Simulator started")

	for s.running {
		s.Tick()
		time.Sleep(s.tickDuration)
	}
}

func (s *Simulator) Stop() {
	s.running = false
	fmt.Println("Simulator stopped")
}

func (s *Simulator) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.tick = 0
	s.tickQueues[s.tick] = make([]Message, 0)
}
