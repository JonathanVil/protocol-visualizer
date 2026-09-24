package simulator

import (
	"errors"
	"reflect"
	"testing"
)

type lifecycleActor struct {
	id       int
	sim      *Simulator
	Received []any
	Revived  int
}

func (a *lifecycleActor) ID() int                     { return a.id }
func (a *lifecycleActor) Init(id int, sim *Simulator) { a.id, a.sim = id, sim }
func (a *lifecycleActor) OnMessage(m Message)         { a.Received = append(a.Received, m.Payload) }
func (a *lifecycleActor) OnRevive()                   { a.Revived++ }
func (a *lifecycleActor) Ping() string                { return "pong" }

func newLifecycleSim(t *testing.T) (*Simulator, *lifecycleActor, *lifecycleActor) {
	t.Helper()
	sim := New()
	sim.RegisterActorType(reflect.TypeFor[*lifecycleActor](), "l")
	a := sim.SpawnActor("l", -1).(*lifecycleActor)
	b := sim.SpawnActor("l", -1).(*lifecycleActor)
	return sim, a, b
}

// drain returns the types of all events buffered on ch.
func drain(ch <-chan Event) []EventType {
	var types []EventType
	for {
		select {
		case e := <-ch:
			types = append(types, e.Type)
		default:
			return types
		}
	}
}

func count(types []EventType, want EventType) int {
	n := 0
	for _, t := range types {
		if t == want {
			n++
		}
	}
	return n
}

func actorAlive(sim *Simulator, id int) bool {
	for _, a := range sim.GetSnapshot().Actors {
		if a.ID == id {
			return a.Alive
		}
	}
	panic("actor not in snapshot")
}

func TestKillActor(t *testing.T) {
	sim, a, b := newLifecycleSim(t)
	events, unsubscribe := sim.Subscribe()
	defer unsubscribe()

	if err := sim.KillActor(b.ID()); err != nil {
		t.Fatalf("KillActor: %v", err)
	}
	if err := sim.KillActor(b.ID()); err != nil {
		t.Fatalf("killing a dead actor should be a no-op, got %v", err)
	}
	if n := count(drain(events), EventActorKilled); n != 1 {
		t.Errorf("got %d actor.killed events, want 1", n)
	}
	if actorAlive(sim, b.ID()) {
		t.Error("snapshot reports killed actor as alive")
	}

	// Messages to a dead actor are dropped on arrival.
	if err := sim.Send(a.ID(), b.ID(), "hi"); err != nil {
		t.Fatalf("Send to dead actor: %v", err)
	}
	sim.Tick()
	if len(b.Received) != 0 {
		t.Errorf("dead actor received %v", b.Received)
	}
	if n := count(drain(events), EventMessageDropped); n != 1 {
		t.Errorf("got %d message.dropped events, want 1", n)
	}

	// A dead actor can't send or have methods invoked.
	if err := sim.Send(b.ID(), a.ID(), "hi"); !errors.Is(err, ErrActorDead) {
		t.Errorf("Send from dead actor: got %v, want ErrActorDead", err)
	}
	if _, err := sim.InvokeActor(b.ID(), "Ping", nil); !errors.Is(err, ErrActorDead) {
		t.Errorf("InvokeActor on dead actor: got %v, want ErrActorDead", err)
	}

	if err := sim.KillActor(99); !errors.Is(err, ErrActorNotFound) {
		t.Errorf("unknown actor: got %v, want ErrActorNotFound", err)
	}
}

func TestKilledActorsMessagesInTransitStillArrive(t *testing.T) {
	sim, a, b := newLifecycleSim(t)
	if err := sim.Send(a.ID(), b.ID(), "sent before dying"); err != nil {
		t.Fatal(err)
	}
	if err := sim.KillActor(a.ID()); err != nil {
		t.Fatal(err)
	}
	sim.Tick()
	if !reflect.DeepEqual(b.Received, []any{"sent before dying"}) {
		t.Errorf("b received %v", b.Received)
	}
}

func TestReviveActor(t *testing.T) {
	sim, a, b := newLifecycleSim(t)

	if err := sim.ReviveActor(b.ID()); err != nil || b.Revived != 0 {
		t.Fatalf("reviving a live actor should be a no-op: err=%v, OnRevive calls=%d", err, b.Revived)
	}

	b.Received = append(b.Received, "state from before the kill")
	if err := sim.KillActor(b.ID()); err != nil {
		t.Fatal(err)
	}

	events, unsubscribe := sim.Subscribe()
	defer unsubscribe()
	if err := sim.ReviveActor(b.ID()); err != nil {
		t.Fatalf("ReviveActor: %v", err)
	}
	if b.Revived != 1 {
		t.Errorf("OnRevive called %d times, want 1", b.Revived)
	}
	if n := count(drain(events), EventActorRevived); n != 1 {
		t.Errorf("got %d actor.revived events, want 1", n)
	}
	if !actorAlive(sim, b.ID()) {
		t.Error("snapshot reports revived actor as dead")
	}

	// State is kept, and the actor takes part again.
	if err := sim.Send(a.ID(), b.ID(), "after revive"); err != nil {
		t.Fatal(err)
	}
	sim.Tick()
	if !reflect.DeepEqual(b.Received, []any{"state from before the kill", "after revive"}) {
		t.Errorf("b received %v", b.Received)
	}
	if got, err := sim.InvokeActor(b.ID(), "Ping", nil); err != nil || got != "pong" {
		t.Errorf("InvokeActor after revive = %v, %v", got, err)
	}
}

func TestOnReviveIsNotExposedAsMethod(t *testing.T) {
	sim, _, b := newLifecycleSim(t)
	for _, m := range actorMethods(b) {
		if m.Name == "OnRevive" {
			t.Error("OnRevive listed as an invocable method")
		}
	}
	if _, err := sim.InvokeActor(b.ID(), "OnRevive", nil); !errors.Is(err, ErrMethodNotFound) {
		t.Errorf("InvokeActor(OnRevive): got %v, want ErrMethodNotFound", err)
	}
}

type panickyReviver struct{ reflectActor }

func (a *panickyReviver) OnRevive() { panic("boom") }

func TestReviveRecoversFromPanic(t *testing.T) {
	sim := New()
	sim.RegisterActorType(reflect.TypeOf(&panickyReviver{}), "p")
	a := sim.SpawnActor("p", -1)
	if err := sim.KillActor(a.ID()); err != nil {
		t.Fatal(err)
	}
	if err := sim.ReviveActor(a.ID()); err == nil {
		t.Error("expected an error from the panicking OnRevive")
	}
	if !actorAlive(sim, a.ID()) {
		t.Error("actor should be alive even though OnRevive panicked")
	}
}
