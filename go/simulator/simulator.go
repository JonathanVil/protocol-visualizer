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
	mu             sync.Mutex
	actors         map[int]Actor
	actorTypeNames map[int]string
	tickQueues     map[int][]Message
	deliveredIDs   map[string]bool
	tick           int
	running        atomic.Bool
	stopCh         chan struct{}
	subsMu         sync.Mutex
	subs           map[chan Event]struct{}
	cmdMu          sync.Mutex
	history        []Message
	actorTypes     map[string]reflect.Type
	commands       chan QueuedCommand
	TransitTicks   int
	TickDuration   time.Duration
}

func New() *Simulator {
	return &Simulator{
		actors:         make(map[int]Actor),
		actorTypeNames: make(map[int]string),
		tickQueues:     make(map[int][]Message),
		deliveredIDs:   make(map[string]bool),
		tick:           0,
		stopCh:         make(chan struct{}),
		subs:           make(map[chan Event]struct{}),
		actorTypes:     make(map[string]reflect.Type),
		commands:       make(chan QueuedCommand, 32),
		TransitTicks:   1,
		TickDuration:   1000 * time.Millisecond,
	}
}

// Subscribe returns a channel that receives every simulation event emitted
// from now on, and a function that unsubscribes it. Each subscriber has its own
// buffer; events are dropped for a subscriber whose buffer is full.
func (s *Simulator) Subscribe() (<-chan Event, func()) {
	ch := make(chan Event, 256)
	s.subsMu.Lock()
	s.subs[ch] = struct{}{}
	s.subsMu.Unlock()
	return ch, func() {
		s.subsMu.Lock()
		delete(s.subs, ch)
		s.subsMu.Unlock()
	}
}

// Enqueue schedules fn to run on the sim goroutine and returns a channel that
// delivers exactly one CommandResult when fn completes.
// If the sim is not running, fn is executed synchronously on the caller's goroutine,
// serialized with other commands (there is no concurrent tick processing when stopped).
func (s *Simulator) Enqueue(fn func() (any, error)) <-chan CommandResult {
	ch := make(chan CommandResult, 1)
	if !s.running.Load() {
		s.cmdMu.Lock()
		res, err := fn()
		s.cmdMu.Unlock()
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

func (s *Simulator) Send(from, to int, payload any) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// check if actors exist
	if _, ok := s.actors[from]; !ok {
		return fmt.Errorf("%w: %d", ErrActorNotFound, from)
	}
	if _, ok := s.actors[to]; !ok {
		return fmt.Errorf("%w: %d", ErrActorNotFound, to)
	}

	msg := Message{ID: newMessageID(), From: from, To: to, Payload: payload, SentTick: s.tick}
	idx := s.tick + s.TransitTicks
	i := rand.IntN(len(s.tickQueues[idx]) + 1)
	s.tickQueues[idx] = append(s.tickQueues[idx][:i], append([]Message{msg}, s.tickQueues[idx][i:]...)...)

	s.emit(EventMessageSent, MessageSentPayload{
		MessageID:     msg.ID,
		From:          msg.From,
		To:            msg.To,
		Payload:       msg.Payload,
		SentTick:      msg.SentTick,
		DeliverAtTick: idx,
	})
	return nil
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
	// The receiving actor's state may have changed outside a tick.
	s.emit(EventSnapshot, s.GetSnapshot())
	return nil
}

// SetActorField sets a named field on an actor. field is either the Go field
// name or its json tag name, matching the keys used in snapshots.
func (s *Simulator) SetActorField(actorID int, field string, value any) error {
	actor := s.actors[actorID]
	if actor == nil {
		return ErrActorNotFound
	}
	fieldV := fieldByWireName(reflect.ValueOf(actor).Elem(), field)
	if !fieldV.IsValid() || !fieldV.CanSet() {
		return ErrFieldNotFound
	}
	converted, err := convertValue(value, fieldV.Type())
	if err != nil {
		return err
	}
	fieldV.Set(converted)

	s.emit(EventActorFieldChanged, ActorFieldChangedPayload{ActorID: actorID, Field: field, Value: converted.Interface()})
	return nil
}

// InvokeActor calls a named exported method on an actor. Arguments are
// converted to the method's parameter types. A trailing error result is
// returned as the error; remaining results are returned as a single value
// (one result) or a slice (several results).
func (s *Simulator) InvokeActor(actorID int, method string, args []any) (result any, err error) {
	actor := s.actors[actorID]
	if actor == nil {
		return nil, ErrActorNotFound
	}
	if isActorInterfaceMethod(method) {
		return nil, ErrMethodNotFound
	}
	methodV := reflect.ValueOf(actor).MethodByName(method)
	if !methodV.IsValid() {
		return nil, ErrMethodNotFound
	}

	methodT := methodV.Type()
	if methodT.IsVariadic() {
		return nil, fmt.Errorf("%w: variadic methods are not supported", ErrInvalidArgument)
	}
	if methodT.NumIn() != len(args) {
		return nil, fmt.Errorf("%w: %s takes %d arguments, got %d", ErrInvalidArgument, method, methodT.NumIn(), len(args))
	}
	argsV := make([]reflect.Value, len(args))
	for i, arg := range args {
		argsV[i], err = convertValue(arg, methodT.In(i))
		if err != nil {
			return nil, fmt.Errorf("argument %d: %w", i, err)
		}
	}

	// Actor code is user code; don't let a panic take down the simulator.
	defer func() {
		if r := recover(); r != nil {
			result, err = nil, fmt.Errorf("%s panicked: %v", method, r)
		}
	}()
	out := methodV.Call(argsV)
	// The method may have changed actor state outside a tick.
	s.emit(EventSnapshot, s.GetSnapshot())

	errorType := reflect.TypeFor[error]()
	if n := len(out); n > 0 && out[n-1].Type() == errorType {
		if !out[n-1].IsNil() {
			return nil, out[n-1].Interface().(error)
		}
		out = out[:n-1]
	}
	switch len(out) {
	case 0:
		return nil, nil
	case 1:
		return out[0].Interface(), nil
	default:
		values := make([]any, len(out))
		for i, v := range out {
			values[i] = v.Interface()
		}
		return values, nil
	}
}

// SetSpeed sets the tick duration
func (s *Simulator) SetSpeed(tickDuration time.Duration) error {
	if tickDuration <= 0 {
		return fmt.Errorf("%w: tickDuration must be > 0", ErrInvalidArgument)
	}
	s.TickDuration = tickDuration
	s.emit(EventSimSettingsChanged, SimSettingsChangedPayload{TickDurationMs: int(tickDuration.Milliseconds()), TransitTicks: s.TransitTicks})
	return nil
}

// SetTransitTime sets the default message transit ticks.
func (s *Simulator) SetTransitTime(ticks int) error {
	if ticks < 0 {
		return fmt.Errorf("%w: ticks must be >= 0", ErrInvalidArgument)
	}
	s.TransitTicks = ticks
	s.emit(EventSimSettingsChanged, SimSettingsChangedPayload{TickDurationMs: int(s.TickDuration.Milliseconds()), TransitTicks: ticks})
	return nil
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
	s.emit(EventSnapshot, s.GetSnapshot())

	for {
		timer := time.NewTimer(s.TickDuration)
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
		s.emit(EventSnapshot, s.GetSnapshot())
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
	s.deliveredIDs[msg.ID] = true
	s.history = append(s.history, msg)
	s.emit(EventMessageDelivered, MessageDeliveredPayload{
		MessageID: msg.ID,
		From:      msg.From,
		To:        msg.To,
		Payload:   msg.Payload,
	})
	actor.OnMessage(msg)
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
	event := Event{Tick: s.tick, Type: e, Payload: payload}
	s.subsMu.Lock()
	defer s.subsMu.Unlock()
	for ch := range s.subs {
		select {
		case ch <- event:
		default:
			log.Printf("event buffer full, dropping event %s", e)
		}
	}
}

// Tick is retained for external callers and tests.
func (s *Simulator) Tick() {
	s.doTick()
}
