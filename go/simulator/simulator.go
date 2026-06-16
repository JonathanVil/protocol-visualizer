package simulator

import (
	"fmt"
	"log"
	"maps"
	"math/rand/v2"
	"reflect"
	"slices"
	"sync"
	"sync/atomic"
	"time"
)

// CommandResult is the outcome of a queued command executed on the sim goroutine.
type CommandResult struct {
	Value any
	Err   error
}

// QueuedCommand is a unit of work delivered to the sim goroutine via the command channel.
type QueuedCommand struct {
	Execute func() (any, error)
	Reply   chan<- CommandResult
}

type Simulator struct {
	mu              sync.Mutex
	actors          map[int]Actor
	actorTypeNames  map[int]string
	tickQueues      map[int][]Message
	deliveredIDs    map[string]bool
	tick            int
	running         atomic.Bool
	stopCh          chan struct{}
	events          chan Event
	history         []Message
	actorTypes      map[string]reflect.Type
	commands        chan QueuedCommand
	TransitTicks    int
	speedMultiplier float64
}

func New() *Simulator {
	return &Simulator{
		actors:          make(map[int]Actor),
		actorTypeNames:  make(map[int]string),
		tickQueues:      make(map[int][]Message),
		deliveredIDs:    make(map[string]bool),
		tick:            0,
		stopCh:          make(chan struct{}),
		events:          make(chan Event, 64),
		actorTypes:      make(map[string]reflect.Type),
		commands:        make(chan QueuedCommand, 32),
		TransitTicks:    1,
		speedMultiplier: 1.0,
	}
}

// Events returns a read-only view of the simulation event stream.
func (s *Simulator) Events() <-chan Event { return s.events }

// Enqueue schedules fn to run on the sim goroutine and returns a channel that
// delivers exactly one CommandResult when fn completes.
// If the sim is not running, fn is executed synchronously on the caller's goroutine
// (safe because there is no concurrent tick processing when stopped).
func (s *Simulator) Enqueue(fn func() (any, error)) <-chan CommandResult {
	ch := make(chan CommandResult, 1)
	if !s.running.Load() {
		res, err := fn()
		ch <- CommandResult{Value: res, Err: err}
		return ch
	}
	s.commands <- QueuedCommand{Execute: fn, Reply: ch}
	return ch
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
	if id == -1 {
		id = len(s.actors)
	}

	actualType := s.actorTypes[typ]
	if actualType == nil {
		log.Printf("Unknown actor type %s\n", typ)
		return nil
	}

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
	defer func() {
		s.tick++
	}()

	fmt.Printf("-- Tick %d --\n", s.tick)

	// get queue of messages due this tick
	queue := s.tickQueues[s.tick]
	s.mu.Unlock()
	if queue == nil {
		return
	}

	for _, msg := range queue {
		// deliver it
		actor := s.actors[msg.To]
		if actor == nil {
			fmt.Printf("No actor found %d\n", msg.To)
			continue
		}
		actor.OnMessage(msg)
		s.events <- Event{Tick: s.tick, Message: msg}
		s.history = append(s.history, msg)
	}
}

func (s *Simulator) Start() {
	s.mu.Lock()
	if s.running.Load() {
		s.mu.Unlock()
		return // already running; idempotent
	}
	s.stopCh = make(chan struct{})
	s.running.Store(true)
	s.mu.Unlock()

	fmt.Println("Simulator started")

	for {
		interval := time.Duration(float64(time.Second) / s.speedMultiplier)
		timer := time.NewTimer(interval)
		select {
		case <-s.stopCh:
			timer.Stop()
			return
		case <-timer.C:
			s.doTick()
		case qc := <-s.commands:
			timer.Stop()
			res, err := qc.Execute()
			qc.Reply <- CommandResult{Value: res, Err: err}
		}
	}
}

func (s *Simulator) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.running.CompareAndSwap(true, false) {
		close(s.stopCh)
		fmt.Println("Simulator stopped")
	}
}

func (s *Simulator) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.tick = 0
	s.tickQueues = make(map[int][]Message)
	s.deliveredIDs = make(map[string]bool)
}

// --- Internal helpers ---

func (s *Simulator) doTick() {
	fmt.Printf("-- Tick %d --\n", s.tick)
	queue := s.tickQueues[s.tick]
	delete(s.tickQueues, s.tick)
	for _, msg := range queue {
		s.deliverMessage(msg)
	}
	s.tick++
}

func (s *Simulator) deliverMessage(msg Message) {
	actor := s.actors[msg.To]
	if actor == nil {
		fmt.Printf("No actor found %d\n", msg.To)
		return
	}
	actor.OnMessage(msg)
	s.deliveredIDs[msg.ID] = true
	s.history = append(s.history, msg)
	s.emit(Event{
		Type: EventMessageDelivered,
		Tick: s.tick,
		Payload: MessageDeliveredPayload{
			MessageID: msg.ID,
			From:      msg.From,
			To:        msg.To,
			Payload:   msg.Payload,
		},
	})
}

func (s *Simulator) findMessage(id string) (tick int, idx int, err error) {
	if s.deliveredIDs[id] {
		return 0, 0, ErrMessageAlreadyDelivered
	}
	for t, msgs := range s.tickQueues {
		for i, m := range msgs {
			if m.ID == id {
				return t, i, nil
			}
		}
	}
	return 0, 0, ErrMessageNotFound
}

func (s *Simulator) emit(e Event) {
	select {
	case s.events <- e:
	default:
		log.Printf("event buffer full, dropping event %s", e.Type)
	}
}

// Tick is retained for external callers and tests.
func (s *Simulator) Tick() {
	s.doTick()
}
