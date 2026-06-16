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
	s.actorTypeNames[id] = typ
	s.Register(actor)
	s.emit(EventActorSpawned, ActorSpawnedPayload{ActorID: id, TypeName: typ})
	s.emit(EventSnapshot, s.GetSnapshot())
	return actor
}

func (s *Simulator) Send(from, to int, payload any) {
	s.mu.Lock()
	defer s.mu.Unlock()

	msg := Message{ID: newMessageID(), From: from, To: to, Payload: payload}
	idx := s.tick + s.TransitTicks
	i := rand.IntN(len(s.tickQueues[idx]) + 1)
	s.tickQueues[idx] = append(s.tickQueues[idx][:i], append([]Message{msg}, s.tickQueues[idx][i:]...)...)
}

func (s *Simulator) GetActorTypes() []string {
	return slices.Collect(maps.Keys(s.actorTypes))
}

// --- Sim-loop-only methods (called from commands channel, no locking needed) ---

// DropMessage removes an in-transit message without delivering it.
func (s *Simulator) DropMessage(id string) error {
	tick, idx, err := s.findMessage(id)
	if err != nil {
		return err
	}
	s.tickQueues[tick] = append(s.tickQueues[tick][:idx], s.tickQueues[tick][idx+1:]...)
	s.deliveredIDs[id] = true
	s.emit(EventMessageDropped, MessageDroppedPayload{MessageID: id})
	return nil
}

// DelayMessage moves an in-transit message's delivery tick forward by ticks.
func (s *Simulator) DelayMessage(id string, ticks int) error {
	if ticks <= 0 {
		return fmt.Errorf("%w: ticks must be > 0", ErrInvalidArgument)
	}
	srcTick, idx, err := s.findMessage(id)
	if err != nil {
		return err
	}
	msg := s.tickQueues[srcTick][idx]
	s.tickQueues[srcTick] = append(s.tickQueues[srcTick][:idx], s.tickQueues[srcTick][idx+1:]...)
	newTick := srcTick + ticks
	s.tickQueues[newTick] = append(s.tickQueues[newTick], msg)
	s.emit(EventMessageDelayed, MessageDelayedPayload{MessageID: id, NewDeliverTick: newTick})
	return nil
}

// DeliverNow immediately delivers an in-transit message.
func (s *Simulator) DeliverNow(id string) error {
	srcTick, idx, err := s.findMessage(id)
	if err != nil {
		return err
	}
	msg := s.tickQueues[srcTick][idx]
	s.tickQueues[srcTick] = append(s.tickQueues[srcTick][:idx], s.tickQueues[srcTick][idx+1:]...)
	s.deliverMessage(msg)
	return nil
}

// SetActorField sets a named field on an actor.
func (s *Simulator) SetActorField(actorID int, field string, value any) error {
	actor := s.actors[actorID]
	if actor == nil {
		return ErrActorNotFound
	}
	v := reflect.ValueOf(actor).Elem()
	fieldV := v.FieldByName("Name")
	if !fieldV.IsValid() {
		return ErrFieldNotFound
	}

	switch fieldV.Type().Kind() {
	case reflect.String:
		if reflect.TypeOf(value).Kind() != reflect.String {
			return fmt.Errorf("%w: value must be a string", ErrInvalidArgument)
		}
		fieldV.SetString(value.(string))
		break
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if reflect.TypeOf(value).Kind() != reflect.Int {
			return fmt.Errorf("%w: value must be an int", ErrInvalidArgument)
		}
		fieldV.SetInt(value.(int64))
		break
	case reflect.Bool:
		if reflect.TypeOf(value).Kind() != reflect.Bool {
			return fmt.Errorf("%w: value must be a bool", ErrInvalidArgument)
		}
		fieldV.SetBool(value.(bool))
		break
	case reflect.Float32, reflect.Float64:
		if reflect.TypeOf(value).Kind() != reflect.Float64 {
			return fmt.Errorf("%w: value must be a float64", ErrInvalidArgument)
		}
		fieldV.SetFloat(value.(float64))
		break
	default:
		return fmt.Errorf("%w: unsupported field type %s", ErrInvalidArgument, fieldV.Type().Kind())
	}

	s.emit(EventActorFieldChanged, ActorFieldChangedPayload{ActorID: actorID, Field: field, Value: value})
	return nil
}

// InvokeActor calls a named method on an actor.
func (s *Simulator) InvokeActor(actorID int, method string, args ...any) (any, error) {
	actor := s.actors[actorID]
	if actor == nil {
		return nil, ErrActorNotFound
	}

	v := reflect.ValueOf(actor).Elem()
	methodV := v.MethodByName(method)
	if !methodV.IsValid() {
		return nil, ErrMethodNotFound
	}

	argsV := make([]reflect.Value, len(args))
	for _, v := range args {
		argsV = append(argsV, reflect.ValueOf(v))
	}

	values := methodV.Call(argsV)
	realValues := make([]any, len(values))
	for _, v := range values {
		switch v.Type().Kind() {
		case reflect.String:
			realValues = append(realValues, v.String())
			break
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			realValues = append(realValues, v.Int())
			break
		case reflect.Bool:
			realValues = append(realValues, v.Bool())
			break
		case reflect.Float32, reflect.Float64:
			realValues = append(realValues, v.Float())
			break
		default:
			realValues = append(realValues, v.Interface())
		}
	}

	return methodV.Call(argsV), nil
}

// SetSpeed sets the speed multiplier (ticks per second = multiplier).
func (s *Simulator) SetSpeed(multiplier float64) error {
	if multiplier <= 0 {
		return fmt.Errorf("%w: multiplier must be > 0", ErrInvalidArgument)
	}
	s.speedMultiplier = multiplier
	s.emit(EventSimSettingsChanged, SimSettingsChangedPayload{SpeedMultiplier: multiplier, TransitTicks: s.TransitTicks})
	return nil
}

// SetTransitTime sets the default message transit ticks.
func (s *Simulator) SetTransitTime(ticks int) error {
	if ticks < 0 {
		return fmt.Errorf("%w: ticks must be >= 0", ErrInvalidArgument)
	}
	s.TransitTicks = ticks
	s.emit(EventSimSettingsChanged, SimSettingsChangedPayload{SpeedMultiplier: s.speedMultiplier, TransitTicks: ticks})
	return nil
}

// Settings holds configurable simulation parameters.
type Settings struct {
	SpeedMultiplier float64 `json:"speedMultiplier"`
	TransitTicks    int     `json:"transitTicks"`
}

// --- Sim loop ---

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
	s.tick++
	fmt.Printf("-- Tick %d --\n", s.tick)
	queue := s.tickQueues[s.tick]
	delete(s.tickQueues, s.tick)
	for _, msg := range queue {
		s.deliverMessage(msg)
	}
	s.emit(EventSnapshot, s.GetSnapshot())
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
	s.emit(EventMessageDelivered, MessageDeliveredPayload{
		MessageID: msg.ID,
		From:      msg.From,
		To:        msg.To,
		Payload:   msg.Payload,
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

func (s *Simulator) emit(e EventType, payload any) {
	select {
	case s.events <- Event{Tick: s.tick, Type: e, Payload: payload}:
	default:
		log.Printf("event buffer full, dropping event %s", e)
	}
}

// Tick is retained for external callers and tests.
func (s *Simulator) Tick() {
	s.doTick()
}
